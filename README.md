# auth-code-mailer

Auth code mailer. Created at Hackathon July 2022 by Hutch

## Prerequisites

- Go
- CDK

## Setup

1. Setup a redis cluster either in AWS via Elasticache or elsewhere
1. Create `.env` file in the root of this repo. Provide the following values

```
VPC_ID=vpc-you-plan-to-use
REDIS_CACHE=redis-cluster-connection-string
```

1. Run `./build.sh` to build the lambdas
1. Run `npm run deploy`
