package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"federal-payment-processing/internal/agents/extraction"
)

// bedrockClientAdapter wraps the AWS Bedrock Runtime client to satisfy the BedrockClient interface.
type bedrockClientAdapter struct {
	client *bedrockruntime.Client
}

func (a *bedrockClientAdapter) InvokeModel(ctx context.Context, modelID string, payload []byte) ([]byte, error) {
	output, err := a.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     &modelID,
		Body:        payload,
		ContentType: strPtr("application/json"),
		Accept:      strPtr("application/json"),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

// s3ClientAdapter wraps the AWS S3 client to satisfy the S3Client interface.
type s3ClientAdapter struct {
	client *s3.Client
}

func (a *s3ClientAdapter) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	output, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer output.Body.Close()
	return io.ReadAll(output.Body)
}

func strPtr(s string) *string {
	return &s
}

func main() {
	// Determine model ID from environment or use default
	modelID := os.Getenv("BEDROCK_MODEL_ID")
	if modelID == "" {
		modelID = "anthropic.claude-3-sonnet-20240229-v1:0"
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	// Create clients
	bedrockClient := &bedrockClientAdapter{
		client: bedrockruntime.NewFromConfig(cfg),
	}
	s3Client := &s3ClientAdapter{
		client: s3.NewFromConfig(cfg),
	}

	// Create handler
	handler := &extraction.Handler{
		BedrockClient: bedrockClient,
		S3Client:      s3Client,
		ModelID:       modelID,
	}

	// Start Lambda handler
	lambda.Start(func(ctx context.Context, event json.RawMessage) (json.RawMessage, error) {
		var extractionEvent extraction.ExtractionEvent
		if err := json.Unmarshal(event, &extractionEvent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event: %w", err)
		}

		result, err := handler.Handle(ctx, extractionEvent)
		if err != nil {
			return nil, err
		}

		responseBytes, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		return responseBytes, nil
	})
}
