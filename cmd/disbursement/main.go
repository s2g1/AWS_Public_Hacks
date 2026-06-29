package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"

	"federal-payment-processing/internal/agents/disbursement"
	"federal-payment-processing/internal/models"
)

// simulated treasury client for the hackathon demo
type simulatedTreasury struct{}

func (t *simulatedTreasury) ExecuteTransfer(from, to disbursement.AccountInfo, amount float64, reference, memo string) (*disbursement.TransferResult, error) {
	// In a real implementation, this would call the actual treasury API.
	// For the hackathon demo, we simulate a successful transfer.
	return &disbursement.TransferResult{
		TransactionID: fmt.Sprintf("TREAS-%s", reference),
		Status:        "SUCCESS",
	}, nil
}

// simulated account lookup for the hackathon demo
type simulatedAccountLookup struct{}

func (a *simulatedAccountLookup) GetPayeeAccount(payee string) (*disbursement.AccountInfo, error) {
	// In a real implementation, this would query a payee registry.
	// For the hackathon demo, we return a simulated account.
	return &disbursement.AccountInfo{
		AccountNumber: "****1234",
		RoutingNumber: "021000021",
		BankName:      "Federal Reserve Bank",
	}, nil
}

func main() {
	// Create handler with simulated dependencies
	handler := disbursement.NewDisbursementHandler(
		&simulatedTreasury{},
		&simulatedAccountLookup{},
		disbursement.AccountInfo{
			AccountNumber: "****0001",
			RoutingNumber: "021000021",
			BankName:      "US Treasury",
		},
	)

	// Start Lambda handler
	lambda.Start(func(ctx context.Context, event json.RawMessage) (json.RawMessage, error) {
		var record models.PaymentRecord
		if err := json.Unmarshal(event, &record); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payment record: %w", err)
		}

		result := handler.Execute(&record)

		responseBytes, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		log.Printf("Disbursement result for payment %s: status=%s", record.PaymentID, result.Status)
		return responseBytes, nil
	})
}
