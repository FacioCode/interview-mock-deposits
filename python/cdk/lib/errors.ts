import * as awsSns from 'aws-cdk-lib/aws-sns'
import * as awsLogs from 'aws-cdk-lib/aws-logs'
import * as awsLambda from 'aws-cdk-lib/aws-lambda'
import { Duration } from 'aws-cdk-lib'
import * as awsCloudwatch from 'aws-cdk-lib/aws-cloudwatch'
import * as awsCloudwatchActions from 'aws-cdk-lib/aws-cloudwatch-actions'
import { isProd } from './utils'
import { FilterPattern, MetricFilter } from 'aws-cdk-lib/aws-logs'
import { Alarm, ComparisonOperator, TreatMissingData } from 'aws-cdk-lib/aws-cloudwatch'
import { SnsAction } from 'aws-cdk-lib/aws-cloudwatch-actions'
import { StateMachine } from 'aws-cdk-lib/aws-stepfunctions'
import { Construct } from 'constructs'

// Lambda Python runtime prefixes log lines with the level in brackets,
// e.g. "[ERROR] 2026-01-01T12:00:00Z ...". CloudWatch text filters match
// those literal tokens.
export const PYTHON_DEFAULT_ERROR_PATTERN = '"[ERROR]"'
export const PYTHON_PANIC_ERROR_PATTERN = '"[CRITICAL]"'

export type LambdaAlarmConfig = {
  errorPattern: string
  threshold: number
  isCritical: boolean
  evaluationPeriods: number
  metricPeriod: Duration
}

export const CriticalAlarmConfig = {
  errorPattern: PYTHON_DEFAULT_ERROR_PATTERN,
  threshold: 95,
  isCritical: true,
  evaluationPeriods: 4,
  metricPeriod: Duration.minutes(15)
}

export const DefaultAlarmConfig = {
  errorPattern: PYTHON_DEFAULT_ERROR_PATTERN,
  threshold: 100,
  isCritical: false,
  evaluationPeriods: 1,
  metricPeriod: Duration.minutes(1)
}

const metricNamespace = 'InterviewMockDepositsPython'

let criticalTopic: any = null
let defaultTopic: any = null

function getCriticalNotificationTopic (scope: Construct) {
  if (criticalTopic == null) {
    criticalTopic = awsSns.Topic.fromTopicArn(scope, 'Critical_Alarms', process.env.ONCALL_ALARM_NOTIFICATION_ARN!)
  }
  return criticalTopic
}

function getDefaultNotificationTopic (scope: Construct) {
  if (defaultTopic == null) {
    defaultTopic = awsSns.Topic.fromTopicArn(scope, 'CloudWatch_Alarms_Topic', process.env.DEFAULT_ALARM_NOTIFICATION_ARN!)
  }
  return defaultTopic
}

function getTopic (scope: Construct, isCritical: boolean) {
  if (isCritical) {
    return getCriticalNotificationTopic(scope)
  }
  return getDefaultNotificationTopic(scope)
}

function defineLowSuccessRateAlarm (scope: Construct, name: string, lambdaFunc: awsLambda.Function, config: LambdaAlarmConfig) {
  const errorFilter = new awsLogs.MetricFilter(scope, name + 'ErrorMetricFilter', {
    logGroup: lambdaFunc.logGroup,
    metricName: name + 'Error',
    filterPattern: awsLogs.FilterPattern.literal(config.errorPattern),
    metricNamespace
  })
  const timeoutFilter = new awsLogs.MetricFilter(scope, name + 'TimeoutMetricFilter', {
    logGroup: lambdaFunc.logGroup,
    metricName: name + 'TimeOut',
    filterPattern: awsLogs.FilterPattern.literal('Task timed out'),
    metricNamespace
  })

  const expression = new awsCloudwatch.MathExpression({
    label: 'successRate',
    expression: '100 - 100 * (errors + timeout) / MAX([errors + timeout, invocations])',
    usingMetrics: {
      errors: errorFilter.metric().with({ statistic: 'sum' }),
      timeout: timeoutFilter.metric().with({ statistic: 'sum' }),
      invocations: lambdaFunc.metricInvocations()
    },
    period: config.metricPeriod
  })

  const alarm = new awsCloudwatch.Alarm(scope, name + 'LowSuccessRateAlarm', {
    metric: expression,
    threshold: config.threshold,
    evaluationPeriods: config.evaluationPeriods,
    comparisonOperator: awsCloudwatch.ComparisonOperator.LESS_THAN_THRESHOLD,
    treatMissingData: awsCloudwatch.TreatMissingData.NOT_BREACHING,
    actionsEnabled: true
  })

  return alarm
}

function definePanicAlarm (scope: Construct, name: string, lambdaFunc: awsLambda.Function) {
  const panicFilter = new awsLogs.MetricFilter(scope, name + 'PanicMetricFilter', {
    logGroup: lambdaFunc.logGroup,
    metricName: name + 'Panic',
    filterPattern: awsLogs.FilterPattern.literal(PYTHON_PANIC_ERROR_PATTERN),
    metricNamespace
  })

  const panicAlarm = new awsCloudwatch.Alarm(scope, name + 'PanicAlarm', {
    metric: panicFilter.metric(),
    threshold: 1,
    evaluationPeriods: 1,
    comparisonOperator: awsCloudwatch.ComparisonOperator.GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
    treatMissingData: awsCloudwatch.TreatMissingData.NOT_BREACHING,
    actionsEnabled: true
  })

  return panicAlarm
}

export function createLambdaAlarms (scope: Construct, name: string, lambdaFunc: awsLambda.Function, config: LambdaAlarmConfig) {
  if (!isProd()) {
    return
  }

  if (config.isCritical) {
    name = name + 'Critical'
  }

  const topic = getTopic(scope, config.isCritical)
  const lowSuccessAlarm = defineLowSuccessRateAlarm(scope, name, lambdaFunc, config)
  const panicAlarm = definePanicAlarm(scope, name, lambdaFunc)

  lowSuccessAlarm.addAlarmAction(new awsCloudwatchActions.SnsAction(topic))
  panicAlarm.addAlarmAction(new awsCloudwatchActions.SnsAction(topic))
}

export const createLambdaErrorAlarm = function (scope: Construct, name: string, lambda: awsLambda.Function, pattern: string) {
  if (!isProd()) {
    return
  }
  const metricFilter = new MetricFilter(scope, name + 'MetricFilter', {
    logGroup: lambda.logGroup,
    metricName: name,
    metricNamespace,
    filterPattern: FilterPattern.literal(pattern)
  })

  const alarm = new Alarm(scope, name + 'Alarm', {
    metric: metricFilter.metric({ statistic: 'Sum' }),
    threshold: 1,
    evaluationPeriods: 1,
    comparisonOperator: ComparisonOperator.GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
    treatMissingData: TreatMissingData.NOT_BREACHING,
    actionsEnabled: true
  })

  alarm.addAlarmAction(new SnsAction(getDefaultNotificationTopic(scope)))
}

export function createStateMachineAlarm (scope: Construct, id: string, stateMachine: StateMachine, alarmConfig: LambdaAlarmConfig = DefaultAlarmConfig) {
  if (!isProd()) {
    return
  }

  const failureAlarm = new Alarm(scope, pascalCase(id) + 'Failure', {
    metric: stateMachine.metricFailed(),
    threshold: 1,
    evaluationPeriods: 1,
    comparisonOperator: ComparisonOperator.GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
    treatMissingData: TreatMissingData.NOT_BREACHING,
    actionsEnabled: true
  })

  const timeoutAlarm = new Alarm(scope, pascalCase(id) + 'TimeOut', {
    metric: stateMachine.metricTimedOut(),
    threshold: 1,
    evaluationPeriods: 1,
    comparisonOperator: ComparisonOperator.GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
    treatMissingData: TreatMissingData.NOT_BREACHING,
    actionsEnabled: true
  })

  const topic = getTopic(scope, alarmConfig.isCritical)

  const snsAction = new SnsAction(topic)
  failureAlarm.addAlarmAction(snsAction)
  timeoutAlarm.addAlarmAction(snsAction)
}

export function createLambdaDurationAlarm (scope: Construct, name: string, lambdaFunc: awsLambda.Function, threshold: number) {
  if (!isProd()) {
    return
  }

  const durationAlarm = new Alarm(scope, name + 'Duration', {
    metric: lambdaFunc.metricDuration({ period: Duration.minutes(1) }),
    threshold,
    evaluationPeriods: 1,
    comparisonOperator: ComparisonOperator.GREATER_THAN_OR_EQUAL_TO_THRESHOLD
  })

  durationAlarm.addAlarmAction(new SnsAction(getDefaultNotificationTopic(scope)))
}

function pascalCase (text: string): string {
  text = text.replace(/[-_\s.]+(.)?/g, (_, c) => c ? c.toUpperCase() : '')
  return text.substring(0, 1).toUpperCase() + text.substring(1)
}
