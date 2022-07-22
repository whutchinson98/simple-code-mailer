// Responsible for queueing and processing email requests
import {Stack, StackProps, Duration} from 'aws-cdk-lib';
import {Construct} from 'constructs';
import {Queue} from 'aws-cdk-lib/aws-sqs';
import {Runtime, Code, Function} from 'aws-cdk-lib/aws-lambda';
import {
  SecurityGroup,
  Peer,
  Port,
  IVpc,
  Vpc,
  SubnetType,
} from 'aws-cdk-lib/aws-ec2';
import {SqsEventSource} from 'aws-cdk-lib/aws-lambda-event-sources';
import {PolicyStatement, Effect} from 'aws-cdk-lib/aws-iam';

export class EmailerStack extends Stack {
  readonly mailerQueue: Queue;
  readonly vpc: IVpc;
  readonly lambdaSecurityGroup: SecurityGroup;

  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    this.vpc = Vpc.fromLookup(this, `baseVpc`, {
      vpcId: `${process.env.VPC_ID}`,
    });

    this.lambdaSecurityGroup = new SecurityGroup(
      this,
      'code-email-sender-sec-group',
      {
        vpc: this.vpc,
        allowAllOutbound: true,
        securityGroupName: 'code-email-sender-sec-group',
      },
    );

    this.lambdaSecurityGroup.addIngressRule(Peer.anyIpv4(), Port.allTraffic());

    this.mailerQueue = new Queue(this, 'mailer-queue', {
      queueName: `mailer-queue`,
      visibilityTimeout: Duration.seconds(300),
    });

    const emailSender = new Function(this, `code-email-sender-lambda`, {
      functionName: `code-email-sender-lambda`,
      runtime: Runtime.GO_1_X,
      code: Code.fromAsset('app/emailSender', {}),
      handler: 'emailSender',
      vpc: this.vpc,
      vpcSubnets: {
        subnetType: SubnetType.PRIVATE_WITH_NAT,
      },
      securityGroups: [this.lambdaSecurityGroup],
      environment: {},
    });

    const eventSource = new SqsEventSource(this.mailerQueue);

    emailSender.addEventSource(eventSource);

    emailSender.addToRolePolicy(
      new PolicyStatement({
        effect: Effect.ALLOW,
        actions: [
          'ses:SendEmail',
          'ses:SendRawEmail',
          'ses:SendTemplatedEmail',
        ],
        resources: [`*`],
      }),
    );
  }
}
