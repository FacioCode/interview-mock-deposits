import { App, SecretValue, Stack, StackProps } from 'aws-cdk-lib'
import { CodeBuildAction, GitHubSourceAction } from 'aws-cdk-lib/aws-codepipeline-actions'
import { Artifact, Pipeline } from 'aws-cdk-lib/aws-codepipeline'
import { BuildSpec, LinuxBuildImage, PipelineProject, Project, Source } from 'aws-cdk-lib/aws-codebuild'
import * as awsIam from 'aws-cdk-lib/aws-iam'
import { Topic } from 'aws-cdk-lib/aws-sns'
import { DetailType, NotificationRule } from 'aws-cdk-lib/aws-codestarnotifications'

export interface CDPipelineStackProps extends StackProps {
  readonly projectName: string
  readonly repoName: string
}

export class CDPipelineStack extends Stack {
  constructor (app: App, id: string, props: CDPipelineStackProps) {
    /* eslint-disable no-new */
    super(app, id, props)
    const gitHubAccount = 'example-org'
    const gitHubOAuthToken = SecretValue.secretsManager('github-ci-token', {
      jsonField: 'oAuthToken'
    })

    const sourceOutput = new Artifact('SourceArtifact')
    const cdkBuildOutput = new Artifact('BuildArtifact')

    const roleDev = awsIam.Role.fromRoleArn(this, 'roleDev', 'arn:aws:iam::333333333333:role/service-role/codebuild-CDK-Deploy-Dev-service-role', { mutable: false }) as awsIam.Role
    const devProj = this.getCodeBuildProject('Dev', props.projectName, roleDev)

    const adminLocalRole = new awsIam.Role(this, 'localAdminRole', {
      managedPolicies: [awsIam.ManagedPolicy.fromAwsManagedPolicyName('AdministratorAccess')],
      assumedBy: new awsIam.ServicePrincipal('codebuild.amazonaws.com')
    })
    const buildProj = this.getCodeBuildProject('Build', props.projectName, adminLocalRole)

    const pipeline = new Pipeline(this, 'CD-Pipeline-' + props.projectName, {
      pipelineName: props.projectName,
      stages: [
        {
          stageName: 'Source',
          actions: [
            new GitHubSourceAction({
              actionName: 'Source',
              owner: gitHubAccount,
              repo: props.repoName,
              output: sourceOutput,
              oauthToken: gitHubOAuthToken,
              branch: 'main'
            })
          ]
        },
        {
          stageName: 'Build',
          actions: [
            new CodeBuildAction({
              actionName: 'Build',
              project: buildProj,
              input: sourceOutput,
              outputs: [cdkBuildOutput]
            })
          ]
        },
        {
          stageName: 'DeployDev',
          actions: [
            new CodeBuildAction({
              actionName: 'Dev',
              project: devProj,
              input: cdkBuildOutput
            })
          ]
        }
      ]
    })

    this.createNotification(pipeline)

    const prProjectName = props.projectName + '-PR'
    new Project(this, prProjectName, {
      role: adminLocalRole,
      projectName: prProjectName,
      source: Source.gitHub({
        owner: gitHubAccount,
        repo: props.repoName
      }),
      buildSpec: BuildSpec.fromSourceFilename('cdk/buildspec/build.yml'),
      environment: {
        buildImage: LinuxBuildImage.STANDARD_7_0
      }
    })
  }

  private getCodeBuildProject (stepName: string, projectName: string, role?: awsIam.IRole) {
    const name = projectName + '-' + stepName
    return new PipelineProject(this, name, {
      projectName: name,
      buildSpec: BuildSpec.fromSourceFilename('cdk/buildspec/' + stepName.toLowerCase() + '.yml'),
      environment: {
        buildImage: LinuxBuildImage.STANDARD_7_0
      },
      role
    })
  }

  private createNotification (pipeline: Pipeline) {
    if (!process.env.SLACK_NOTIFICATION_SNS_ARN) {
      return
    }

    const targetTopic = Topic.fromTopicArn(this, 'topic', process.env.SLACK_NOTIFICATION_SNS_ARN!)
    new NotificationRule(this, 'slackCiCdNotification', {
      detailType: DetailType.BASIC,
      events: [
        'codepipeline-pipeline-pipeline-execution-started',
        'codepipeline-pipeline-pipeline-execution-resumed',
        'codepipeline-pipeline-pipeline-execution-succeeded',

        'codepipeline-pipeline-stage-execution-failed',
        'codepipeline-pipeline-stage-execution-canceled'
      ],
      targets: [targetTopic],
      source: pipeline
    })
  }
}
