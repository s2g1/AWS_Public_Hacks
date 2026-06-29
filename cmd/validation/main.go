package main

import (
	"context"
	"encoding/json"
	"fmt"

	"federal-payment-processing/internal/agents/validation"
	"federal-payment-processing/internal/models"
)

// Handler is the Lambda function handler for the Validation Agent.
// It accepts an ExtractionResult as input and returns a ValidationResult.
func Handler(ctx context.Context, event json.RawMessage) (models.ValidationResult, error) {
	var extraction models.ExtractionResult
	if err := json.Unmarshal(event, &extraction); err != nil {
		return models.ValidationResult{}, fmt.Errorf("failed to unmarshal extraction result: %w", err)
	}

	result := validation.ValidatePayment(extraction)
	return result, nil
}

func main() {
	// In production, this would use the AWS Lambda Go runtime:
	// lambda.Start(Handler)
	// For now, kept as a placeholder for the Lambda entry point.
	fmt.Println("Validation Agent Lambda - use with AWS Lambda runtime")
}
