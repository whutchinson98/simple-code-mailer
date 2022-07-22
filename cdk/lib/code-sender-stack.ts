// Responsible for sending email to queue and generating code in redis
import {Stack, StackProps} from 'aws-cdk-lib';
import {Construct} from 'constructs';
import {CfnCacheCluster} from 'aws-cdk-lib/aws-elasticache';
import {
  IVpc,
  Vpc,
  SubnetType,
  SecurityGroup,
  Peer,
  Port,
} from 'aws-cdk-lib/aws-ec2';
import {Runtime, Code, Function} from 'aws-cdk-lib/aws-lambda';
import {PolicyStatement, Effect} from 'aws-cdk-lib/aws-iam';

interface CodeSenderStackProps extends StackProps {
  queueUrl: string;
}

export class CodeSenderStack extends Stack {
  readonly redisCache: CfnCacheCluster;
  readonly lambdaSecurityGroup: SecurityGroup;
  readonly vpc: IVpc;
  constructor(scope: Construct, id: string, props: CodeSenderStackProps) {
    super(scope, id, props);

    this.vpc = Vpc.fromLookup(this, `baseVpc`, {
      vpcId: `${process.env.VPC_ID}`,
    });

    this.lambdaSecurityGroup = new SecurityGroup(
      this,
      'code-email-gen-sec-group',
      {
        vpc: this.vpc,
        allowAllOutbound: true,
        securityGroupName: 'code-email-gen-sec-group',
      },
    );

    this.lambdaSecurityGroup.addIngressRule(Peer.anyIpv4(), Port.allTraffic());

    const codeGenerator = new Function(this, `code-email-generator-lambda`, {
      functionName: `code-email-generator-lambda`,
      runtime: Runtime.GO_1_X,
      code: Code.fromAsset('app/emailGenerator', {}),
      handler: 'emailGenerator',
      vpc: this.vpc,
      vpcSubnets: {
        subnetType: SubnetType.PRIVATE_WITH_NAT,
      },
      securityGroups: [this.lambdaSecurityGroup],
      environment: {
        QUEUE_URL: props.queueUrl,
        REDIS_CACHE: `${process.env.REDIS_CACHE}`,
      },
    });

    codeGenerator.addToRolePolicy(
      new PolicyStatement({
        effect: Effect.ALLOW,
        actions: ['sqs:*'],
        resources: [`*`],
      }),
    );
  }
}
