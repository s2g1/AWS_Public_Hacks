import * as cdk from 'aws-cdk-lib';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as s3deploy from 'aws-cdk-lib/aws-s3-deployment';
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront';
import * as origins from 'aws-cdk-lib/aws-cloudfront-origins';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';
import * as path from 'path';

export class InfraStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // =========================================================================
    // Frontend Hosting (S3 + CloudFront) — existing
    // =========================================================================

    // S3 bucket for hosting the React frontend
    const websiteBucket = new s3.Bucket(this, 'FedPayPortalBucket', {
      bucketName: `fedpay-portal-${this.account}-${this.region}`,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      autoDeleteObjects: true,
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
    });

    // CloudFront distribution for the frontend
    const distribution = new cloudfront.Distribution(this, 'FedPayDistribution', {
      defaultBehavior: {
        origin: origins.S3BucketOrigin.withOriginAccessControl(websiteBucket),
        viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
        cachePolicy: cloudfront.CachePolicy.CACHING_OPTIMIZED,
      },
      defaultRootObject: 'index.html',
      errorResponses: [
        {
          httpStatus: 403,
          responseHttpStatus: 200,
          responsePagePath: '/index.html',
          ttl: cdk.Duration.minutes(5),
        },
        {
          httpStatus: 404,
          responseHttpStatus: 200,
          responsePagePath: '/index.html',
          ttl: cdk.Duration.minutes(5),
        },
      ],
    });

    // Deploy the built frontend to S3
    new s3deploy.BucketDeployment(this, 'DeployFrontend', {
      sources: [s3deploy.Source.asset(path.join(__dirname, '../../frontend/dist'))],
      destinationBucket: websiteBucket,
      distribution,
      distributionPaths: ['/*'],
    });

    // =========================================================================
    // DynamoDB Tables
    // =========================================================================

    // Payments table — stores payment records through the processing pipeline
    const paymentsTable = new dynamodb.Table(this, 'PaymentsTable', {
      tableName: `fedpay-payments-${this.account}`,
      partitionKey: { name: 'paymentId', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      pointInTimeRecovery: true,
    });

    // GSI: query payments by status (e.g., find all ESCALATED payments)
    paymentsTable.addGlobalSecondaryIndex({
      indexName: 'status-index',
      partitionKey: { name: 'status', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'updatedAt', type: dynamodb.AttributeType.STRING },
      projectionType: dynamodb.ProjectionType.ALL,
    });

    // GSI: query payments by payee (for duplicate detection and history)
    paymentsTable.addGlobalSecondaryIndex({
      indexName: 'payee-index',
      partitionKey: { name: 'payee', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'createdAt', type: dynamodb.AttributeType.STRING },
      projectionType: dynamodb.ProjectionType.ALL,
    });

    // Contracts table — stores contract financial data
    const contractsTable = new dynamodb.Table(this, 'ContractsTable', {
      tableName: `fedpay-contracts-${this.account}`,
      partitionKey: { name: 'contractId', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      pointInTimeRecovery: true,
    });

    // CLINs table — contract line items
    const clinsTable = new dynamodb.Table(this, 'CLINsTable', {
      tableName: `fedpay-clins-${this.account}`,
      partitionKey: { name: 'contractId', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'clinId', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      pointInTimeRecovery: true,
    });

    // REAs table — Requests for Equitable Adjustment
    const reasTable = new dynamodb.Table(this, 'REAsTable', {
      tableName: `fedpay-reas-${this.account}`,
      partitionKey: { name: 'reaId', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'contractId', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      pointInTimeRecovery: true,
    });

    // =========================================================================
    // S3 Bucket: Document Ingestion
    // =========================================================================

    const ingestionBucket = new s3.Bucket(this, 'DocumentIngestionBucket', {
      bucketName: `fedpay-ingestion-${this.account}-${this.region}`,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      autoDeleteObjects: true,
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      encryption: s3.BucketEncryption.S3_MANAGED,
      versioned: true,
      lifecycleRules: [
        {
          id: 'TransitionToIA',
          transitions: [
            {
              storageClass: s3.StorageClass.INFREQUENT_ACCESS,
              transitionAfter: cdk.Duration.days(30),
            },
          ],
        },
        {
          id: 'ExpireOldVersions',
          noncurrentVersionExpiration: cdk.Duration.days(90),
        },
        {
          id: 'AbortIncompleteMultipart',
          abortIncompleteMultipartUploadAfter: cdk.Duration.days(7),
        },
      ],
    });

    // =========================================================================
    // IAM Role: Lambda Agent Execution Role
    // =========================================================================

    const agentExecutionRole = new iam.Role(this, 'AgentExecutionRole', {
      roleName: `fedpay-agent-execution-${this.region}`,
      assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com'),
      description: 'Execution role for payment processing agent Lambda functions',
      managedPolicies: [
        iam.ManagedPolicy.fromAwsManagedPolicyName('service-role/AWSLambdaBasicExecutionRole'),
      ],
    });

    // Grant DynamoDB access to all tables
    paymentsTable.grantReadWriteData(agentExecutionRole);
    contractsTable.grantReadWriteData(agentExecutionRole);
    clinsTable.grantReadWriteData(agentExecutionRole);
    reasTable.grantReadWriteData(agentExecutionRole);

    // Grant S3 access to the ingestion bucket
    ingestionBucket.grantReadWrite(agentExecutionRole);

    // Grant Amazon Bedrock invoke access
    agentExecutionRole.addToPolicy(new iam.PolicyStatement({
      sid: 'BedrockInvokeModel',
      effect: iam.Effect.ALLOW,
      actions: [
        'bedrock:InvokeModel',
        'bedrock:InvokeModelWithResponseStream',
      ],
      resources: ['*'],
    }));

    // =========================================================================
    // Outputs
    // =========================================================================

    new cdk.CfnOutput(this, 'DistributionDomainName', {
      value: `https://${distribution.distributionDomainName}`,
      description: 'CloudFront URL for the Federal Payment Processing Portal',
    });

    new cdk.CfnOutput(this, 'BucketName', {
      value: websiteBucket.bucketName,
      description: 'S3 bucket hosting the frontend assets',
    });

    new cdk.CfnOutput(this, 'IngestionBucketName', {
      value: ingestionBucket.bucketName,
      description: 'S3 bucket for document ingestion',
    });

    new cdk.CfnOutput(this, 'PaymentsTableName', {
      value: paymentsTable.tableName,
      description: 'DynamoDB table for payment records',
    });

    new cdk.CfnOutput(this, 'ContractsTableName', {
      value: contractsTable.tableName,
      description: 'DynamoDB table for contract data',
    });

    new cdk.CfnOutput(this, 'CLINsTableName', {
      value: clinsTable.tableName,
      description: 'DynamoDB table for contract line items',
    });

    new cdk.CfnOutput(this, 'REAsTableName', {
      value: reasTable.tableName,
      description: 'DynamoDB table for REAs',
    });

    new cdk.CfnOutput(this, 'AgentExecutionRoleArn', {
      value: agentExecutionRole.roleArn,
      description: 'IAM role ARN for agent Lambda functions',
    });
  }
}
