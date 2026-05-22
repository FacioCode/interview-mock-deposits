import * as cdk from 'aws-cdk-lib'
import { ServiceStack } from '../cdk/lib/service-stack'

test('Empty Stack', () => {
  const app = new cdk.App()
  // WHEN
  const stack = new ServiceStack(app, 'MyTestStack')
  // THEN
  const actual = app.synth().getStackArtifact(stack.artifactId).template
  expect(actual.Resources ?? {}).toEqual({})
})
