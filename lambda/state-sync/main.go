package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	bucketName      = os.Getenv("STATE_BUCKET")
	stateKey        = "shared-state/app-state.json"
	connectionsTable = os.Getenv("CONNECTIONS_TABLE")
	wsEndpoint      = os.Getenv("WS_ENDPOINT")
)

func handler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if request.RequestContext.HTTP.Method == "OPTIONS" {
		return events.LambdaFunctionURLResponse{StatusCode: 200, Headers: headers}, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return errorResp(500, "AWS config error: "+err.Error(), headers), nil
	}
	s3Client := s3.NewFromConfig(cfg)

	method := request.RequestContext.HTTP.Method

	switch method {
	case "GET":
		result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &bucketName,
			Key:    strPtr(stateKey),
		})
		if err != nil {
			if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
				return events.LambdaFunctionURLResponse{
					StatusCode: 200,
					Headers:    headers,
					Body:       `{"exists":false}`,
				}, nil
			}
			return errorResp(500, "S3 read error: "+err.Error(), headers), nil
		}
		defer result.Body.Close()
		body, _ := io.ReadAll(result.Body)
		return events.LambdaFunctionURLResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       string(body),
		}, nil

	case "PUT", "POST":
		contentType := "application/json"
		bodyReader := strings.NewReader(request.Body)
		_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      &bucketName,
			Key:         strPtr(stateKey),
			Body:        bodyReader,
			ContentType: &contentType,
		})
		if err != nil {
			return errorResp(500, "S3 write error: "+err.Error(), headers), nil
		}

		// Notify all connected WebSocket clients
		go notifyClients(ctx, cfg)

		return events.LambdaFunctionURLResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       `{"success":true}`,
		}, nil

	default:
		return errorResp(405, "Method not allowed", headers), nil
	}
}

func notifyClients(ctx context.Context, cfg interface{}) {
	if connectionsTable == "" || wsEndpoint == "" {
		return
	}

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return
	}

	ddbClient := dynamodb.NewFromConfig(awsCfg)
	result, err := ddbClient.Scan(ctx, &dynamodb.ScanInput{
		TableName: &connectionsTable,
	})
	if err != nil || len(result.Items) == 0 {
		return
	}

	// Create API Gateway Management API client
	apiClient := apigatewaymanagementapi.NewFromConfig(awsCfg, func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = &wsEndpoint
	})

	msg := []byte(`{"event":"state_changed"}`)

	for _, item := range result.Items {
		connIDAttr, ok := item["connectionId"]
		if !ok {
			continue
		}
		connID := connIDAttr.(*types.AttributeValueMemberS).Value
		apiClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: &connID,
			Data:         msg,
		})
	}
}

func errorResp(code int, msg string, headers map[string]string) events.LambdaFunctionURLResponse {
	body, _ := json.Marshal(map[string]string{"error": msg})
	return events.LambdaFunctionURLResponse{StatusCode: code, Headers: headers, Body: string(body)}
}

func strPtr(s string) *string { return &s }

func main() {
	lambda.Start(handler)
}
