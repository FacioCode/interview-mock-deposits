import { App, Stack, StackProps, Duration } from 'aws-cdk-lib'
import { Table, AttributeType } from 'aws-cdk-lib/aws-dynamodb'
import { StartingPosition } from 'aws-cdk-lib/aws-lambda'
import { DomainNameAttributes } from 'aws-cdk-lib/aws-apigateway/lib/domain-name'
import * as awsApigateway from 'aws-cdk-lib/aws-apigateway'
import * as awsIam from 'aws-cdk-lib/aws-iam'
import * as awsEvents from 'aws-cdk-lib/aws-events'
import * as awsEventsTargets from 'aws-cdk-lib/aws-events-targets'
import { DynamoEventSource } from 'aws-cdk-lib/aws-lambda-event-sources'
import * as utils from './utils'
import { CriticalAlarmConfig } from './errors'

export interface ApplicationStackProps {
  domainName: DomainNameAttributes | undefined
}

const TABLE_NAME = 'InterviewMockDepositsTable'
const BASE_PATH = 'deposits'
const EVENT_BUS_NAME = 'InterviewMockDepositsBus'
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

    this.createDepositLambda(api, table, env)
    this.createWebhookLambda(api, table, env)
    this.createStreamConsumerLambda(table, env)
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

    table.addGlobalSecondaryIndex({
      indexName: 'endToEndIdIndex',
      partitionKey: { name: 'endToEndId', type: AttributeType.STRING }
    })

    return table
  }

  private createDepositLambda (api: awsApigateway.RestApi, table: Table, env: Record<string, any>) {
    const lambda = utils.createLambdaWithDynamoAccess(
      this, 'deposit', 'deposit', table, [], env, Duration.seconds(5), CriticalAlarmConfig
    )
    const integration = new awsApigateway.LambdaIntegration(lambda)
    api.root.addMethod('GET', integration, { apiKeyRequired: true })
    api.root.addResource('{depositId}').addMethod('GET', integration, { apiKeyRequired: true })
  }

  private createWebhookLambda (api: awsApigateway.RestApi, table: Table, env: Record<string, any>) {
    const lambda = utils.createLambdaWithDynamoAccess(
      this, 'webhook', 'webhook', table, [], env, Duration.seconds(30), CriticalAlarmConfig
    )
    api.root.addResource('webhook').addMethod('POST', new awsApigateway.LambdaIntegration(lambda))
  }

  private createStreamConsumerLambda (table: Table, env: Record<string, any>) {
    const initialPolicy = [
      new awsIam.PolicyStatement({
        actions: ['events:PutEvents'],
        resources: ['*'],
        effect: awsIam.Effect.ALLOW
      })
    ]
    const lambda = utils.createLambdaWithDynamoAccess(
      this, 'streamConsumer', 'stream_consumer', table, initialPolicy, env, Duration.minutes(15), CriticalAlarmConfig, 2
    )
    lambda.addEventSource(new DynamoEventSource(table, {
      startingPosition: StartingPosition.TRIM_HORIZON,
      batchSize: 1,
      retryAttempts: 2
    }))
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
