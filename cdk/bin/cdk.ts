#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import {EmailerStack} from '../lib/emailer-stack';
import {CodeSenderStack} from '../lib/code-sender-stack';

const app = new cdk.App();

const env = {
  account: 'YOUR_AWS_ACCOUNT_ID_HERE',
  region: 'us-east-1',
};

// Add additional tags here
const tags = {
  project: 'auth-code-mailer',
};

const emailStack = new EmailerStack(app, `emailer-stack`, {
  env,
  tags,
});

new CodeSenderStack(app, `code-sender-stack`, {
  env,
  queueUrl: emailStack.mailerQueue.queueUrl,
  tags,
});
