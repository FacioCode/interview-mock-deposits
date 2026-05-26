import { App, Stack, StackProps, Duration } from 'aws-cdk-lib'
import { Table, AttributeType } from 'aws-cdk-lib/aws-dynamodb'
import { DomainNameAttributes } from 'aws-cdk-lib/aws-apigateway'
import * as awsApigateway from 'aws-cdk-lib/aws-apigateway'
import * as awsIam from 'aws-cdk-lib/aws-iam'
import * as awsEvents from 'aws-cdk-lib/aws-events'
import * as awsEventsTargets from 'aws-cdk-lib/aws-events-targets'
import * as utils from './utils'
import { CriticalAlarmConfig } from './errors'

export interface ApplicationStackProps {
  domainName: DomainNameAttributes | undefined
}

const TABLE_NAME = 'InterviewMockDepositsTable'
const BASE_PATH = 'deposits'
const EVENT_BUS_NAME = 'InterviewMockDepositsBus'

const PENDING_TRANSACTION_SOURCE = 'interview.mock.transactions'
const PENDING_TRANSACTION_DETAIL_TYPE = 'transaction-pending'

const EVENT_SOURCE = 'interview.mock.deposits'
const DEPOSIT_REQUESTED_DETAIL_TYPE = 'deposit-requested'

export class ServiceStack extends Stack {
  private readonly id: string
  private readonly appProps: ApplicationStackProps | undefined

  constructor (scope: App, id: string, props?: StackProps, appProps?: ApplicationStackProps) {
    super(scope, id, props)
    this.id = id
    this.appProps = appProps
  }

  public run () {
    const env = utils.getIntegrationEnv(this)
    const table = this.createTable()
    const api = utils.createApiGateway(this, this.id + 'Api', BASE_PATH, this.appProps?.domainName)
    const eventBus = new awsEvents.EventBus(this, 'eventBus', { eventBusName: EVENT_BUS_NAME })

    this.createNewDepositLambda(eventBus, table, env)
    this.createWebhookLambda(api, table, env)
    this.createDepositInProviderLambda(eventBus, table, env)
  }

  private createTable (): Table {
    const table = utils.createTable(
      this,
      TABLE_NAME,
      { name: 'depositId', type: AttributeType.STRING },
      undefined,
      'ttl'
    )

    table.addGlobalSecondaryIndex({
      indexName: 'userIndex',
      partitionKey: { name: 'userId', type: AttributeType.STRING }
    })

    return table
  }

  // Reference pattern: EventBridge-listener lambda that creates a new
  // deposit on an upstream "transaction-pending" event. Handler + tests +
  // delegation to a src/ package mirror the deposit_in_provider shape so
  // candidates have a complete worked example to model on.
  private createNewDepositLambda (eventBus: awsEvents.EventBus, table: Table, env: Record<string, any>) {
    const initialPolicy = [
      new awsIam.PolicyStatement({
        actions: ['events:PutEvents'],
        resources: [eventBus.eventBusArn],
        effect: awsIam.Effect.ALLOW
      })
    ]
    const lambda = utils.createLambdaWithDynamoAccess(
      this, 'createNewDeposit', 'create_new_deposit', table, initialPolicy, env, Duration.seconds(30), CriticalAlarmConfig
    )
    const rule = new awsEvents.Rule(this, 'pendingTransactionRule', {
      eventBus,
      eventPattern: {
        source: [PENDING_TRANSACTION_SOURCE],
        detailType: [PENDING_TRANSACTION_DETAIL_TYPE]
      }
    })
    rule.addTarget(new awsEventsTargets.LambdaFunction(lambda, { retryAttempts: 0 }))
  }

  private createWebhookLambda (api: awsApigateway.RestApi, table: Table, env: Record<string, any>) {
    const lambda = utils.createLambdaWithDynamoAccess(
      this, 'webhook', 'webhook', table, [], env, Duration.seconds(30), CriticalAlarmConfig
    )
    api.root.addResource('webhook').addMethod('POST', new awsApigateway.LambdaIntegration(lambda))
  }

  private createDepositInProviderLambda (eventBus: awsEvents.EventBus, table: Table, env: Record<string, any>) {
    const lambda = utils.createLambdaWithDynamoAccess(
      this, 'depositInProvider', 'deposit_in_provider', table, [], env, Duration.seconds(120), CriticalAlarmConfig
    )
    const rule = new awsEvents.Rule(this, 'newDepositRule', {
      eventBus,
      eventPattern: {
        source: [EVENT_SOURCE],
        detailType: [DEPOSIT_REQUESTED_DETAIL_TYPE]
      }
    })
    rule.addTarget(new awsEventsTargets.LambdaFunction(lambda, { retryAttempts: 0 }))
  }
}
