package ingestion

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

// mockS3Uploader is a test double for S3Uploader that records uploads.
type mockS3Uploader struct {
	uploads []uploadCall
	err     error
}

type uploadCall struct {
	Bucket string
	Key    string
	Data   []byte
}

func (m *mockS3Uploader) Upload(ctx context.Context, bucket, key string, data []byte) error {
	m.uploads = append(m.uploads, uploadCall{Bucket: bucket, Key: key, Data: data})
	return m.err
}

// buildRawEmail constructs a raw MIME email with the given attachments.
func buildRawEmail(from, subject string, attachments []testAttachment) []byte {
	boundary := "----=_Part_12345"
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("From: %s\r\n", from))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString(fmt.Sprintf("Message-ID: <test-msg-001@example.com>\r\n"))
	sb.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	sb.WriteString("\r\n")

	// Text body part
	sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	sb.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	sb.WriteString("\r\n")
	sb.WriteString("Please find the attached invoice.\r\n")
	sb.WriteString("\r\n")

	// Attachment parts
	for _, att := range attachments {
		sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		sb.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", att.ContentType, att.Filename))
		sb.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", att.Filename))
		sb.WriteString("Content-Transfer-Encoding: base64\r\n")
		sb.WriteString("\r\n")
		encoded := base64.StdEncoding.EncodeToString(att.Data)
		sb.WriteString(encoded)
		sb.WriteString("\r\n")
	}

	sb.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return []byte(sb.String())
}

type testAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

func TestProcessEmail_WithPDFAttachment(t *testing.T) {
	uploader := &mockS3Uploader{}
	handler := &EmailIngestionHandler{
		Uploader: uploader,
		Bucket:   "test-ingestion-bucket",
	}

	pdfData := []byte("%PDF-1.4 fake pdf content for testing")
	rawEmail := buildRawEmail(
		"sender@example.com",
		"Invoice #12345",
		[]testAttachment{
			{Filename: "invoice.pdf", ContentType: "application/pdf", Data: pdfData},
		},
	)

	result, err := handler.ProcessEmail(context.Background(), rawEmail)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Verify S3 path format
	if !strings.Contains(result.DocumentPath, "emails/") {
		t.Errorf("expected DocumentPath to contain 'emails/', got %s", result.DocumentPath)
	}
	if !strings.Contains(result.DocumentPath, "test-msg-001@example.com") {
		t.Errorf("expected DocumentPath to contain message ID, got %s", result.DocumentPath)
	}
	if !strings.Contains(result.DocumentPath, "invoice.pdf") {
		t.Errorf("expected DocumentPath to contain filename, got %s", result.DocumentPath)
	}

	// Verify metadata
	if result.Metadata.Channel != ChannelEMAIL {
		t.Errorf("expected channel EMAIL, got %s", result.Metadata.Channel)
	}
	if result.Metadata.SourceIdentifier != "sender@example.com" {
		t.Errorf("expected sender 'sender@example.com', got %s", result.Metadata.SourceIdentifier)
	}
	if result.Metadata.OriginalFilename != "invoice.pdf" {
		t.Errorf("expected original filename 'invoice.pdf', got %s", result.Metadata.OriginalFilename)
	}

	// Verify upload was called
	if len(uploader.uploads) != 1 {
		t.Fatalf("expected 1 upload, got %d", len(uploader.uploads))
	}
	if uploader.uploads[0].Bucket != "test-ingestion-bucket" {
		t.Errorf("expected bucket 'test-ingestion-bucket', got %s", uploader.uploads[0].Bucket)
	}
	if !strings.HasPrefix(uploader.uploads[0].Key, "emails/") {
		t.Errorf("expected key to start with 'emails/', got %s", uploader.uploads[0].Key)
	}

	// Verify PaymentID is set
	if result.PaymentID == "" {
		t.Error("expected non-empty PaymentID")
	}
}

func TestProcessEmail_WithMultipleAttachments(t *testing.T) {
	uploader := &mockS3Uploader{}
	handler := &EmailIngestionHandler{
		Uploader: uploader,
		Bucket:   "test-ingestion-bucket",
	}

	pdfData := []byte("%PDF-1.4 fake pdf content")
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG magic bytes

	rawEmail := buildRawEmail(
		"vendor@company.org",
		"Multiple documents",
		[]testAttachment{
			{Filename: "invoice.pdf", ContentType: "application/pdf", Data: pdfData},
			{Filename: "receipt.png", ContentType: "image/png", Data: pngData},
		},
	)

	result, err := handler.ProcessEmail(context.Background(), rawEmail)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Verify both attachments were uploaded
	if len(uploader.uploads) != 2 {
		t.Fatalf("expected 2 uploads, got %d", len(uploader.uploads))
	}

	// Verify first upload is the PDF
	if !strings.Contains(uploader.uploads[0].Key, "invoice.pdf") {
		t.Errorf("expected first upload key to contain 'invoice.pdf', got %s", uploader.uploads[0].Key)
	}

	// Verify second upload is the PNG
	if !strings.Contains(uploader.uploads[1].Key, "receipt.png") {
		t.Errorf("expected second upload key to contain 'receipt.png', got %s", uploader.uploads[1].Key)
	}

	// Result should reference the first attachment
	if !strings.Contains(result.DocumentPath, "invoice.pdf") {
		t.Errorf("expected result DocumentPath to reference first attachment, got %s", result.DocumentPath)
	}

	// Metadata should reference the sender
	if result.Metadata.SourceIdentifier != "vendor@company.org" {
		t.Errorf("expected sender 'vendor@company.org', got %s", result.Metadata.SourceIdentifier)
	}
}

func TestProcessEmail_NoAttachments_ReturnsError(t *testing.T) {
	uploader := &mockS3Uploader{}
	handler := &EmailIngestionHandler{
		Uploader: uploader,
		Bucket:   "test-ingestion-bucket",
	}

	// Build an email with only a text part (no supported attachments)
	boundary := "----=_Part_99999"
	var sb strings.Builder
	sb.WriteString("From: sender@example.com\r\n")
	sb.WriteString("Subject: No attachments\r\n")
	sb.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	sb.WriteString("\r\n")
	sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	sb.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	sb.WriteString("\r\n")
	sb.WriteString("This email has no attachments.\r\n")
	sb.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	rawEmail := []byte(sb.String())

	result, err := handler.ProcessEmail(context.Background(), rawEmail)
	if err == nil {
		t.Fatal("expected error for email with no attachments")
	}
	if result != nil {
		t.Error("expected nil result for email with no attachments")
	}
	if !strings.Contains(err.Error(), "no attachments found") {
		t.Errorf("expected error about no attachments, got: %v", err)
	}

	// Verify no uploads were made
	if len(uploader.uploads) != 0 {
		t.Errorf("expected 0 uploads, got %d", len(uploader.uploads))
	}
}

func TestProcessEmail_MetadataCorrectlyPopulated(t *testing.T) {
	uploader := &mockS3Uploader{}
	handler := &EmailIngestionHandler{
		Uploader: uploader,
		Bucket:   "ingestion-bucket",
	}

	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0} // JPEG magic bytes
	rawEmail := buildRawEmail(
		"John Doe <john.doe@agency.gov>",
		"Travel Voucher Scan",
		[]testAttachment{
			{Filename: "travel_voucher.jpg", ContentType: "image/jpeg", Data: jpegData},
		},
	)

	result, err := handler.ProcessEmail(context.Background(), rawEmail)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify channel is EMAIL
	if result.Metadata.Channel != ChannelEMAIL {
		t.Errorf("expected channel EMAIL, got %s", result.Metadata.Channel)
	}

	// Verify sender is extracted correctly (just the address, not the name)
	if result.Metadata.SourceIdentifier != "john.doe@agency.gov" {
		t.Errorf("expected sender 'john.doe@agency.gov', got %s", result.Metadata.SourceIdentifier)
	}

	// Verify original filename
	if result.Metadata.OriginalFilename != "travel_voucher.jpg" {
		t.Errorf("expected filename 'travel_voucher.jpg', got %s", result.Metadata.OriginalFilename)
	}

	// Verify ReceivedAt is set (not zero)
	if result.Metadata.ReceivedAt.IsZero() {
		t.Error("expected non-zero ReceivedAt timestamp")
	}

	// Verify PaymentID is a valid UUID
	if result.PaymentID == "" {
		t.Error("expected non-empty PaymentID")
	}
}

func TestProcessEmail_NonMultipart_ReturnsError(t *testing.T) {
	uploader := &mockS3Uploader{}
	handler := &EmailIngestionHandler{
		Uploader: uploader,
		Bucket:   "test-bucket",
	}

	// Simple text email without multipart
	rawEmail := []byte("From: sender@example.com\r\nSubject: Plain text\r\nContent-Type: text/plain\r\n\r\nJust a plain text email.\r\n")

	result, err := handler.ProcessEmail(context.Background(), rawEmail)
	if err == nil {
		t.Fatal("expected error for non-multipart email")
	}
	if result != nil {
		t.Error("expected nil result")
	}
	if !strings.Contains(err.Error(), "not multipart") {
		t.Errorf("expected error about non-multipart, got: %v", err)
	}
}

func TestProcessEmail_UnsupportedAttachmentTypes_Ignored(t *testing.T) {
	uploader := &mockS3Uploader{}
	handler := &EmailIngestionHandler{
		Uploader: uploader,
		Bucket:   "test-bucket",
	}

	// Build email with unsupported attachment type only
	boundary := "----=_Part_77777"
	var sb strings.Builder
	sb.WriteString("From: sender@example.com\r\n")
	sb.WriteString("Subject: Word doc\r\n")
	sb.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	sb.WriteString("\r\n")
	sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	sb.WriteString("Content-Type: text/plain\r\n")
	sb.WriteString("\r\n")
	sb.WriteString("Body text\r\n")
	sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	sb.WriteString("Content-Type: application/msword; name=\"doc.docx\"\r\n")
	sb.WriteString("Content-Disposition: attachment; filename=\"doc.docx\"\r\n")
	sb.WriteString("Content-Transfer-Encoding: base64\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(base64.StdEncoding.EncodeToString([]byte("fake docx content")))
	sb.WriteString("\r\n")
	sb.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	rawEmail := []byte(sb.String())

	result, err := handler.ProcessEmail(context.Background(), rawEmail)
	if err == nil {
		t.Fatal("expected error since no supported attachments")
	}
	if result != nil {
		t.Error("expected nil result")
	}
	if !strings.Contains(err.Error(), "no attachments found") {
		t.Errorf("expected error about no supported attachments, got: %v", err)
	}
}

func TestProcessEmail_S3UploadError(t *testing.T) {
	uploader := &mockS3Uploader{
		err: fmt.Errorf("simulated S3 error"),
	}
	handler := &EmailIngestionHandler{
		Uploader: uploader,
		Bucket:   "test-bucket",
	}

	pdfData := []byte("%PDF-1.4 content")
	rawEmail := buildRawEmail(
		"sender@example.com",
		"Upload failure test",
		[]testAttachment{
			{Filename: "doc.pdf", ContentType: "application/pdf", Data: pdfData},
		},
	)

	result, err := handler.ProcessEmail(context.Background(), rawEmail)
	if err == nil {
		t.Fatal("expected error when S3 upload fails")
	}
	if result != nil {
		t.Error("expected nil result on upload failure")
	}
	if !strings.Contains(err.Error(), "failed to upload") {
		t.Errorf("expected upload error message, got: %v", err)
	}
}
