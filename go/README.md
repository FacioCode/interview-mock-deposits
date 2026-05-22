# Interview Mock Deposits — Go

Go implementation of the interview exercise. CDK-deployed lambda service with DynamoDB-backed deposit handlers.

## Commands

```
npm install         # install CDK/typescript deps
npm run build       # build Go lambda artifacts (./src/...) via build.sh
./test.sh           # run Go tests against DynamoDB Local
npm run test:cdk    # run CDK stack tests
```
