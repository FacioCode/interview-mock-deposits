import { Aws, Duration, RemovalPolicy } from 'aws-cdk-lib'
import * as awsSqs from 'aws-cdk-lib/aws-sqs'
import * as awsLambda from 'aws-cdk-lib/aws-lambda'
import * as awsDynamodb from 'aws-cdk-lib/aws-dynamodb'
import * as awsIam from 'aws-cdk-lib/aws-iam'
import * as awsApigateway from 'aws-cdk-lib/aws-apigateway'
import { Secret } from 'aws-cdk-lib/aws-secretsmanager'
import { Construct } from 'constructs'
import { DefaultAlarmConfig, LambdaAlarmConfig, createLambdaAlarms } from './errors'

export function isProd () {
  return process.env.APP_ENV === 'prod'
}

export function isDev () {
  return process.env.APP_ENV === 'dev'
}

export function createApiProxyToSqs (scope: Construct, resource: awsApigateway.Resource, queue: awsSqs.Queue) {
  const apiProxyToSqsRole = new awsIam.Role(scope, 'apiProxyToSqsRole' + resource.path, {
    assumedBy: new awsIam.ServicePrincipal('apigateway.amazonaws.com')
  })

  apiProxyToSqsRole.addToPolicy(new awsIam.PolicyStatement({
    resources: [
      queue.queueArn
    ],
    actions: [
      'sqs:SendMessage'
    ]
  }))

  const awsSqsIntegration = new awsApigateway.AwsIntegration({
    service: 'sqs',
    path: Aws.ACCOUNT_ID + '/' + queue.queueName,
    options: {
      passthroughBehavior: awsApigateway.PassthroughBehavior.NEVER,
      credentialsRole: apiProxyToSqsRole,
      requestParameters: {
        'integration.request.header.Content-Type': "'application/x-www-form-urlencoded'"
      },
      requestTemplates: {
        'application/json': 'Action=SendMessage&MessageBody=$input.body'
      },
      integrationResponses: [
        {
          statusCode: '200',
          responseTemplates: {
            'application/json': JSON.stringify({ success: true })
          },
          selectionPattern: '200'
        },
        {
          statusCode: '500',
          responseTemplates: {
            'application/json': JSON.stringify({ success: false })
          },
          selectionPattern: '500'
        }
      ]
    }
  })
  resource.addMethod('POST', awsSqsIntegration, {
    apiKeyRequired: true,
    methodResponses: [{ statusCode: '200' }, { statusCode: '500' }]
  })
}

export function createApiGateway (scope: Construct, id: string, basePath: string, domainName?: awsApigateway.DomainNameAttributes): awsApigateway.RestApi {
  const api = new awsApigateway.RestApi(scope, id, {
    cloudWatchRoleRemovalPolicy: isProd() ? RemovalPolicy.RETAIN : RemovalPolicy.DESTROY
  })
  const apiKey = api.addApiKey(id + ' api key')
  const usagePlan = api.addUsagePlan('usage_plan')
  usagePlan.addApiKey(apiKey)
  usagePlan.addApiStage({ stage: api.deploymentStage })
  setDomainName(scope, basePath, domainName, api)
  return api
}

export function createTable (scope: Construct, tableName: string, partitionKey: awsDynamodb.Attribute, sortKey?: awsDynamodb.Attribute, timeToLiveAttribute?: string): awsDynamodb.Table {
  const table = new awsDynamodb.Table(scope, tableName, {
    tableName,
    partitionKey,
    sortKey,
    timeToLiveAttribute,
    billingMode: awsDynamodb.BillingMode.PAY_PER_REQUEST,
    stream: awsDynamodb.StreamViewType.NEW_AND_OLD_IMAGES,
    removalPolicy: isProd() ? RemovalPolicy.RETAIN : RemovalPolicy.DESTROY,
    pointInTimeRecoverySpecification: {
      pointInTimeRecoveryEnabled: isProd()
    }
  })

  return table
}

export function createPythonLambda (scope: Construct, id: string, path: string, env?: any, timeout?: Duration, initialPolicy?: awsIam.PolicyStatement[], alarmConfig?: LambdaAlarmConfig, retry?: any): awsLambda.Function {
  const name = pascalCase(id)
  const lambda = new awsLambda.Function(scope, name, {
    runtime: awsLambda.Runtime.PYTHON_3_12,
    code: awsLambda.Code.fromAsset(`build/${path}/main.zip`),
    handler: `${path}.handler.handle_request`,
    environment: env,
    timeout,
    initialPolicy,
    retryAttempts: retry
  })

  createLambdaAlarms(scope, id, lambda, alarmConfig ?? DefaultAlarmConfig)
  return lambda
}

export function createLambda (scope: Construct, id: string, path: string, handler: string, initialPolicy?: any, env?: any, timeout?: Duration, alarmConfig?: LambdaAlarmConfig): awsLambda.Function {
  const lambda = new awsLambda.Function(scope, id, {
    runtime: awsLambda.Runtime.NODEJS_20_X,
    code: awsLambda.Code.fromAsset(path),
    handler,
    initialPolicy,
    environment: env,
    timeout
  })

  createLambdaAlarms(scope, id, lambda, alarmConfig ?? DefaultAlarmConfig)
  return lambda
}

export function createLambdaWithDynamoAccess (scope: Construct, id: string, path: string, table: awsDynamodb.Table, initialPolicy?: awsIam.PolicyStatement[], env?: any, timeout?: Duration, alarmConfig?: LambdaAlarmConfig | undefined, retry?: any): awsLambda.Function {
  const lambda = createPythonLambda(scope, id, path, { ...env, TABLE_NAME: table.tableName }, timeout, initialPolicy, alarmConfig, retry)
  const tablePolicy = new awsIam.PolicyStatement({
    actions: ['dynamodb:*'],
    resources: [table.tableArn, `${table.tableArn}/*`],
    effect: awsIam.Effect.ALLOW
  })
  lambda.addToRolePolicy(tablePolicy)
  return lambda
}

export function createSqsQueue (scope: Construct, queueName: string): awsSqs.Queue {
  const deadLetter = new awsSqs.Queue(scope, queueName + 'DeadLetter')
  return new awsSqs.Queue(scope, queueName, {
    deadLetterQueue: {
      queue: deadLetter,
      maxReceiveCount: 1
    }
  })
}

// returns an ISecret object
// use .secretValueFromJson(key) to interpret it as JSON and get the SecretValue with given key
export function getSecret (scope: Construct, secretId: string, secretName: string) {
  return Secret.fromSecretNameV2(scope, secretId, secretName)
}

function setDomainName (scope: Construct, basePath: string, domain: awsApigateway.DomainNameAttributes | undefined, api: awsApigateway.RestApi) {
  /* eslint-disable no-new */
  if (domain === undefined) return
  const domainName = awsApigateway.DomainName.fromDomainNameAttributes(scope, 'domainName',
    {
      domainName: domain.domainName,
      domainNameAliasTarget: domain.domainNameAliasTarget,
      domainNameAliasHostedZoneId: domain.domainNameAliasHostedZoneId
    })
  new awsApigateway.BasePathMapping(scope, 'basePath', { domainName, restApi: api, basePath })
}

function pascalCase (text: string): string {
  text = text.replace(/[-_\s.]+(.)?/g, (_, c) => c ? c.toUpperCase() : '')
  return text.substring(0, 1).toUpperCase() + text.substring(1)
}

export function getIntegrationEnv (scope: Construct) {
  if (isProd()) {
    return {
      CLOUD_URL: 'https://api.example.com',
      EVENT_BRIDGE_DLQ_URL: 'https://sqs.sa-east-1.amazonaws.com/222222222222/eventBus-dlq'
    }
  }
  return {
    CLOUD_URL: 'https://api.example.dev',
    EVENT_BRIDGE_DLQ_URL: 'https://sqs.us-east-1.amazonaws.com/111111111111/eventBus-dlq'
  }
}
