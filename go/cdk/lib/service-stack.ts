import { App, Stack, StackProps } from 'aws-cdk-lib'
import { Table, AttributeType } from 'aws-cdk-lib/aws-dynamodb'
import { DomainNameAttributes } from 'aws-cdk-lib/aws-apigateway/lib/domain-name'
import * as utils from './utils'

export interface ApplicationStackProps {
  domainName: DomainNameAttributes | undefined
}

const TABLE_NAME = 'InterviewMockDepositsTable'
const BASE_PATH = 'deposits'

export class ServiceStack extends Stack {
  private readonly id: string
  private readonly appProps: ApplicationStackProps | undefined

  constructor (scope: App, id: string, props?: StackProps, appProps?: ApplicationStackProps) {
    super(scope, id, props)
    this.id = id
    this.appProps = appProps
  }

  public run () {
    this.createTable()
    const api = utils.createApiGateway(this, this.id + 'Api', BASE_PATH, this.appProps?.domainName)
    api.root.addMethod('GET')
  }

  private createTable (): Table {
    return utils.createTable(this, TABLE_NAME,
      {
        name: 'depositId',
        type: AttributeType.STRING
      },
      {
        name: 'itemType',
        type: AttributeType.STRING
      }
    )
  }
}
