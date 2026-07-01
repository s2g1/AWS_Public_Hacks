package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func handler(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	cfg, _ := config.LoadDefaultConfig(ctx)
	client := dynamodb.NewFromConfig(cfg)

	tableName := os.Getenv("CONNECTIONS_TABLE")
	connID := request.RequestContext.ConnectionID

	switch request.RequestContext.RouteKey {
	case "$connect":
		client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: &tableName,
			Item: map[string]types.AttributeValue{
				"connectionId": &types.AttributeValueMemberS{Value: connID},
			},
		})
	case "$disconnect":
		client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &tableName,
			Key: map[string]types.AttributeValue{
				"connectionId": &types.AttributeValueMemberS{Value: connID},
			},
		})
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}
