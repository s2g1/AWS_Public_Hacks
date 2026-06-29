package ingestion

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestProcessFax_ValidFax(t *testing.T) {
	uploader := &mockS3Uploader{}
	handler := &FaxIngestionHandler{
		Uploader: uploader,
		Bucket:   "test-ingestion-bucket",
	}

	faxData := []byte{0x49, 0x49, 0x2A, 0x00} // TIFF magic bytes (little-endian)
	senderNumber := "+15551234567"

	result, err := handler.ProcessFax(context.Background(), faxData, senderNumber)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Verify S3 path format: faxes/{date}/{uuid}/{sender_number}.tiff
	if !strings.Contains(result.DocumentPath, "faxes/") {
		t.Errorf("expected DocumentPath to contain 'faxes/', got %s", result.DocumentPath)
	}
	if !strings.Contains(result.DocumentPath, senderNumber+".tiff") {
		t.Errorf("expected DocumentPath to contain sender number with .tiff extension, got %s", result.DocumentPath)
	}

	// Verify channel is FAX
	if result.Metadata.Channel != ChannelFAX {
		t.Errorf("expected channel FAX, got %s", result.Metadata.Channel)
	}

	// Verify source identifier is the sender fax number
	if result.Metadata.SourceIdentifier != senderNumber {
		t.Errorf("expected source identifier %q, got %q", senderNumber, result.Metadata.SourceIdentifier)
	}

	// Verify upload was called correctly
	if len(uploader.uploads) != 1 {
		t.Fatalf("expected 1 upload, got %d", len(uploader.uploads))
	}
	if uploader.uploads[0].Bucket != "test-ingestion-bucket" {
		t.Errorf("expected bucket 'test-ingestion-bucket', got %s", uploader.uploads[0].Bucket)
	}
	if !strings.HasPrefix(uploader.uploads[0].Key, "faxes/") {
		t.Errorf("expected key to start with 'faxes/', got %s", uploader.uploads[0].Key)
	}

	// Verify PaymentID is generated
	if result.PaymentID == "" {
		t.Error("expected non-empty PaymentID")
	}
}

func TestProcessMail_ValidMail(t *testing.T) {
	uploader := &mockS3Uploader{}
	handler := &MailIngestionHandler{
		Uploader: uploader,
		Bucket:   "test-ingestion-bucket",
	}

	scanData := []byte("%PDF-1.4 scanned document content")
	trackingID := "USPS-TRACK-9876543210"

	result, err := handler.ProcessMail(context.Background(), scanData, trackingID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Verify S3 path format: mail/{date}/{trackingID}/scan.pdf
	if !strings.Contains(result.DocumentPath, "mail/") {
		t.Errorf("expected DocumentPath to contain 'mail/', got %s", result.DocumentPath)
	}
	if !strings.Contains(result.DocumentPath, trackingID) {
		t.Errorf("expected DocumentPath to contain tracking ID, got %s", result.DocumentPath)
	}
	if !strings.Contains(result.DocumentPath, "/scan.pdf") {
		t.Errorf("expected DocumentPath to contain '/scan.pdf', got %s", result.DocumentPath)
	}

	// Verify channel is MAIL
	if result.Metadata.Channel != ChannelMAIL {
		t.Errorf("expected channel MAIL, got %s", result.Metadata.Channel)
	}

	// Verify source identifier is the tracking ID
	if result.Metadata.SourceIdentifier != trackingID {
		t.Errorf("expected source identifier %q, got %q", trackingID, result.Metadata.SourceIdentifier)
	}

	// Verify upload was called correctly
	if len(uploader.uploads) != 1 {
		t.Fatalf("expected 1 upload, got %d", len(uploader.uploads))
	}
	if uploader.uploads[0].Bucket != "test-ingestion-bucket" {
		t.Errorf("expected bucket 'test-ingestion-bucket', got %s", uploader.uploads[0].Bucket)
	}
	if !strings.HasPrefix(uploader.uploads[0].Key, "mail/") {
		t.Errorf("expected key to start with 'mail/', got %s", uploader.uploads[0].Key)
	}

	// Verify PaymentID is generated
	if result.PaymentID == "" {
		t.Error("expected non-empty PaymentID")
	}
}

func TestProcessFax_EmptyData_ReturnsError(t *testing.T) {
	uploader := &mockS3Uploader{}
	handler := &FaxIngestionHandler{
		Uploader: uploader,
		Bucket:   "test-bucket",
	}

	result, err := handler.ProcessFax(context.Background(), []byte{}, "+15551234567")
	if err == nil {
		t.Fatal("expected error for empty fax data")
	}
	if result != nil {
		t.Error("expected nil result for empty fax data")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected error about empty data, got: %v", err)
	}

	// Verify no uploads were made
	if len(uploader.uploads) != 0 {
		t.Errorf("expected 0 uploads, got %d", len(uploader.uploads))
	}
}

func TestProcessMail_EmptyData_ReturnsError(t *testing.T) {
	uploader := &mockS3Uploader{}
	handler := &MailIngestionHandler{
		Uploader: uploader,
		Bucket:   "test-bucket",
	}

	result, err := handler.ProcessMail(context.Background(), []byte{}, "TRACK-123")
	if err == nil {
		t.Fatal("expected error for empty scan data")
	}
	if result != nil {
		t.Error("expected nil result for empty scan data")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected error about empty data, got: %v", err)
	}

	// Verify no uploads were made
	if len(uploader.uploads) != 0 {
		t.Errorf("expected 0 uploads, got %d", len(uploader.uploads))
	}
}

func TestProcessFax_MetadataPopulatedCorrectly(t *testing.T) {
	uploader := &mockS3Uploader{}
	handler := &FaxIngestionHandler{
		Uploader: uploader,
		Bucket:   "ingestion-bucket",
	}

	faxData := []byte{0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00} // TIFF data
	senderNumber := "+18005551234"

	result, err := handler.ProcessFax(context.Background(), faxData, senderNumber)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all metadata fields are populated
	if result.Metadata.Channel != ChannelFAX {
		t.Errorf("expected channel FAX, got %s", result.Metadata.Channel)
	}
	if result.Metadata.SourceIdentifier != senderNumber {
		t.Errorf("expected source identifier %q, got %q", senderNumber, result.Metadata.SourceIdentifier)
	}
	if result.Metadata.ReceivedAt.IsZero() {
		t.Error("expected non-zero ReceivedAt timestamp")
	}
	if result.Metadata.OriginalFilename == "" {
		t.Error("expected non-empty OriginalFilename")
	}
	if result.PaymentID == "" {
		t.Error("expected non-empty PaymentID")
	}

	// Verify the S3 upload contains the fax data
	if len(uploader.uploads) != 1 {
		t.Fatalf("expected 1 upload, got %d", len(uploader.uploads))
	}
	if len(uploader.uploads[0].Data) != len(faxData) {
		t.Errorf("expected uploaded data length %d, got %d", len(faxData), len(uploader.uploads[0].Data))
	}
}

func TestProcessMail_S3UploadError(t *testing.T) {
	uploader := &mockS3Uploader{
		err: fmt.Errorf("simulated S3 error"),
	}
	handler := &MailIngestionHandler{
		Uploader: uploader,
		Bucket:   "test-bucket",
	}

	scanData := []byte("%PDF-1.4 content")

	result, err := handler.ProcessMail(context.Background(), scanData, "TRACK-456")
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
