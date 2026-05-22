#!/usr/bin/env node
import 'source-map-support/register'
import * as cdk from 'aws-cdk-lib'
import { ServiceStack } from '../lib/service-stack'
import { CDPipelineStack } from '../lib/cd-pipeline-stack'
import * as utils from '../lib/utils'

const app = new cdk.App()

const STACK_NAME = 'InterviewMockDepositsPythonStack'
const REPO_NAME = 'interview-mock-deposits-python'

function run () {
  /* eslint-disable no-new */

  let domainName
  let env = { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION }

  if (utils.isProd()) {
    env = { region: 'sa-east-1', account: '222222222222' }
    domainName = {
      domainName: 'api.example.com',
      domainNameAliasHostedZoneId: 'Z2FDTNDATAQYW2',
      domainNameAliasTarget: 'dEXAMPLEPROD.cloudfront.net'
    }
  } else if (utils.isDev()) {
    env = { region: 'us-east-1', account: '111111111111' }
    domainName = {
      domainName: 'api.example.dev',
      domainNameAliasHostedZoneId: 'Z2FDTNDATAQYW2',
      domainNameAliasTarget: 'dEXAMPLEDEV.cloudfront.net'
    }
  }

  new ServiceStack(app, STACK_NAME, {
    env,
    stackName: STACK_NAME
  }, { domainName }).run()

  new CDPipelineStack(app, 'CDPipeline' + STACK_NAME, {
    projectName: STACK_NAME,
    repoName: REPO_NAME
  })
}

run()
