// Package internal provides a dependency anchor for go mod tidy.
// This file ensures AWS SDK v2 dependencies are retained in go.mod.
// It will be superseded by actual implementation files.
package internal

import (
	_ "github.com/aws/aws-sdk-go-v2/config"
	_ "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	_ "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	_ "github.com/aws/aws-sdk-go-v2/service/lambda"
	_ "github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/aws/aws-sdk-go-v2/service/sfn"
)
