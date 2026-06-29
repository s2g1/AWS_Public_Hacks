package ingestion

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FaxIngestionHandler processes incoming fax documents for the payment processing pipeline.
type FaxIngestionHandler struct {
	Uploader S3Uploader
	Bucket   string
}

// ProcessFax accepts base64-decoded fax image data and uploads it to S3
// for processing by the payment pipeline.
func (h *FaxIngestionHandler) ProcessFax(ctx context.Context, faxData []byte, senderNumber string) (*IngestionResult, error) {
	if len(faxData) == 0 {
		return nil, fmt.Errorf("fax data is empty")
	}

	now := time.Now().UTC()
	dateStr := now.Format("2006-01-02")
	docID := uuid.New().String()

	key := fmt.Sprintf("faxes/%s/%s/%s.tiff", dateStr, docID, senderNumber)
	s3Path := fmt.Sprintf("s3://%s/%s", h.Bucket, key)

	if err := h.Uploader.Upload(ctx, h.Bucket, key, faxData); err != nil {
		return nil, fmt.Errorf("failed to upload fax to S3: %w", err)
	}

	return &IngestionResult{
		DocumentPath: s3Path,
		Metadata: IngestionMetadata{
			Channel:          ChannelFAX,
			SourceIdentifier: senderNumber,
			ReceivedAt:       now,
			OriginalFilename: fmt.Sprintf("%s.tiff", senderNumber),
		},
		PaymentID: uuid.New().String(),
	}, nil
}

// MailIngestionHandler processes scanned physical mail documents for the payment processing pipeline.
type MailIngestionHandler struct {
	Uploader S3Uploader
	Bucket   string
}

// ProcessMail accepts scanned document data from a mail scanning station
// and uploads it to S3 for processing by the payment pipeline.
func (h *MailIngestionHandler) ProcessMail(ctx context.Context, scanData []byte, trackingID string) (*IngestionResult, error) {
	if len(scanData) == 0 {
		return nil, fmt.Errorf("scan data is empty")
	}

	now := time.Now().UTC()
	dateStr := now.Format("2006-01-02")

	key := fmt.Sprintf("mail/%s/%s/scan.pdf", dateStr, trackingID)
	s3Path := fmt.Sprintf("s3://%s/%s", h.Bucket, key)

	if err := h.Uploader.Upload(ctx, h.Bucket, key, scanData); err != nil {
		return nil, fmt.Errorf("failed to upload scanned mail to S3: %w", err)
	}

	return &IngestionResult{
		DocumentPath: s3Path,
		Metadata: IngestionMetadata{
			Channel:          ChannelMAIL,
			SourceIdentifier: trackingID,
			ReceivedAt:       now,
			OriginalFilename: "scan.pdf",
		},
		PaymentID: uuid.New().String(),
	}, nil
}
