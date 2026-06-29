package ingestion

import (
	"context"
	"strings"
	"testing"
)

// buildTestEmailForRouting constructs a minimal multipart MIME email with a PDF attachment.
func buildTestEmailForRouting(from, messageID string) []byte {
	boundary := "----TestBoundary456"
	email := "From: " + from + "\r\n" +
		"To: payments@agency.gov\r\n" +
		"Subject: Invoice submission\r\n" +
		"Message-ID: <" + messageID + ">\r\n" +
		"Content-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Please process attached invoice.\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: application/pdf\r\n" +
		"Content-Disposition: attachment; filename=\"invoice.pdf\"\r\n" +
		"\r\n" +
		"fake-pdf-content\r\n" +
		"--" + boundary + "--\r\n"
	return []byte(email)
}

func TestEmailChannel_ProducesEmailMetadata(t *testing.T) {
	mock := &mockS3Uploader{}
	handler := &EmailIngestionHandler{
		Uploader: mock,
		Bucket:   "test-bucket",
	}

	rawEmail := buildTestEmailForRouting("sender@example.com", "msg-route-001")
	result, err := handler.ProcessEmail(context.Background(), rawEmail)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Metadata.Channel != ChannelEMAIL {
		t.Errorf("expected channel EMAIL, got %s", result.Metadata.Channel)
	}
	if result.Metadata.SourceIdentifier != "sender@example.com" {
		t.Errorf("expected source identifier 'sender@example.com', got %s", result.Metadata.SourceIdentifier)
	}
}

func TestFaxChannel_ProducesFaxMetadata(t *testing.T) {
	mock := &mockS3Uploader{}
	handler := &FaxIngestionHandler{
		Uploader: mock,
		Bucket:   "test-bucket",
	}

	faxData := []byte("fake-tiff-data")
	result, err := handler.ProcessFax(context.Background(), faxData, "+15551234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Metadata.Channel != ChannelFAX {
		t.Errorf("expected channel FAX, got %s", result.Metadata.Channel)
	}
	if result.Metadata.SourceIdentifier != "+15551234567" {
		t.Errorf("expected source identifier '+15551234567', got %s", result.Metadata.SourceIdentifier)
	}
}

func TestMailChannel_ProducesMailMetadata(t *testing.T) {
	mock := &mockS3Uploader{}
	handler := &MailIngestionHandler{
		Uploader: mock,
		Bucket:   "test-bucket",
	}

	scanData := []byte("fake-pdf-scan-data")
	result, err := handler.ProcessMail(context.Background(), scanData, "TRACK-2024-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Metadata.Channel != ChannelMAIL {
		t.Errorf("expected channel MAIL, got %s", result.Metadata.Channel)
	}
	if result.Metadata.SourceIdentifier != "TRACK-2024-001" {
		t.Errorf("expected source identifier 'TRACK-2024-001', got %s", result.Metadata.SourceIdentifier)
	}
}

func TestAllChannels_ProduceValidS3PathsAndPaymentIDs(t *testing.T) {
	tests := []struct {
		name    string
		channel IngestionChannel
		run     func(uploader S3Uploader) (*IngestionResult, error)
	}{
		{
			name:    "Email channel",
			channel: ChannelEMAIL,
			run: func(uploader S3Uploader) (*IngestionResult, error) {
				h := &EmailIngestionHandler{Uploader: uploader, Bucket: "bucket"}
				return h.ProcessEmail(context.Background(), buildTestEmailForRouting("test@test.com", "id-route-123"))
			},
		},
		{
			name:    "Fax channel",
			channel: ChannelFAX,
			run: func(uploader S3Uploader) (*IngestionResult, error) {
				h := &FaxIngestionHandler{Uploader: uploader, Bucket: "bucket"}
				return h.ProcessFax(context.Background(), []byte("fax-data"), "+1555000")
			},
		},
		{
			name:    "Mail channel",
			channel: ChannelMAIL,
			run: func(uploader S3Uploader) (*IngestionResult, error) {
				h := &MailIngestionHandler{Uploader: uploader, Bucket: "bucket"}
				return h.ProcessMail(context.Background(), []byte("scan-data"), "TRACK-001")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockS3Uploader{}
			result, err := tc.run(mock)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify S3 path is valid
			if !strings.HasPrefix(result.DocumentPath, "s3://bucket/") {
				t.Errorf("expected S3 path to start with 's3://bucket/', got %s", result.DocumentPath)
			}
			if result.DocumentPath == "s3://bucket/" {
				t.Error("S3 path should have a key after the bucket")
			}

			// Verify PaymentID is non-empty
			if result.PaymentID == "" {
				t.Error("expected non-empty PaymentID")
			}

			// Verify PaymentID looks like a UUID (36 chars with hyphens)
			if len(result.PaymentID) != 36 {
				t.Errorf("expected PaymentID to be a UUID (36 chars), got %d chars: %s", len(result.PaymentID), result.PaymentID)
			}
		})
	}
}
