package ingestion

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// IngestionChannel represents the source channel of a document.
type IngestionChannel string

const (
	ChannelEMAIL  IngestionChannel = "EMAIL"
	ChannelFAX    IngestionChannel = "FAX"
	ChannelMAIL   IngestionChannel = "MAIL"
	ChannelPORTAL IngestionChannel = "PORTAL"
)

// IngestionMetadata holds metadata about how a document was ingested.
type IngestionMetadata struct {
	Channel          IngestionChannel `json:"channel"`
	SourceIdentifier string           `json:"sourceIdentifier"` // email address, fax number, tracking ID
	ReceivedAt       time.Time        `json:"receivedAt"`
	OriginalFilename string           `json:"originalFilename"`
}

// IngestionResult represents the result of ingesting a document.
type IngestionResult struct {
	DocumentPath string            `json:"documentPath"` // S3 path
	Metadata     IngestionMetadata `json:"metadata"`
	PaymentID    string            `json:"paymentId"`
}

// S3Uploader defines the interface for uploading objects to S3.
type S3Uploader interface {
	Upload(ctx context.Context, bucket, key string, data []byte) error
}

// EmailIngestionHandler processes incoming emails and extracts attachments
// for the payment processing pipeline.
type EmailIngestionHandler struct {
	Uploader S3Uploader
	Bucket   string
}

// supportedAttachmentTypes lists MIME types that are accepted as payment documents.
var supportedAttachmentTypes = map[string]bool{
	"application/pdf": true,
	"image/png":       true,
	"image/jpeg":      true,
	"image/tiff":      true,
}

// ProcessEmail parses a raw MIME email, extracts PDF/image attachments,
// uploads them to S3, and returns an IngestionResult for each attachment.
func (h *EmailIngestionHandler) ProcessEmail(ctx context.Context, rawEmail []byte) (*IngestionResult, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(rawEmail))
	if err != nil {
		return nil, fmt.Errorf("failed to parse email: %w", err)
	}

	sender := extractSender(msg)
	messageID := extractMessageID(msg)
	if messageID == "" {
		messageID = uuid.New().String()
	}

	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		return nil, fmt.Errorf("no attachments found: missing Content-Type header")
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Content-Type: %w", err)
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		return nil, fmt.Errorf("no attachments found: email is not multipart (type: %s)", mediaType)
	}

	boundary := params["boundary"]
	if boundary == "" {
		return nil, fmt.Errorf("no attachments found: missing boundary parameter")
	}

	attachments, err := extractAttachments(msg.Body, boundary)
	if err != nil {
		return nil, fmt.Errorf("failed to extract attachments: %w", err)
	}

	if len(attachments) == 0 {
		return nil, fmt.Errorf("no attachments found: email contains no supported attachments (PDF, PNG, JPEG, TIFF)")
	}

	now := time.Now().UTC()
	dateStr := now.Format("2006-01-02")

	// Upload each attachment and return the result for the first one
	// (multiple attachments are all uploaded, result references first)
	var firstResult *IngestionResult
	for _, att := range attachments {
		key := fmt.Sprintf("emails/%s/%s/%s", dateStr, messageID, att.Filename)
		s3Path := fmt.Sprintf("s3://%s/%s", h.Bucket, key)

		if err := h.Uploader.Upload(ctx, h.Bucket, key, att.Data); err != nil {
			return nil, fmt.Errorf("failed to upload attachment %s to S3: %w", att.Filename, err)
		}

		result := &IngestionResult{
			DocumentPath: s3Path,
			Metadata: IngestionMetadata{
				Channel:          ChannelEMAIL,
				SourceIdentifier: sender,
				ReceivedAt:       now,
				OriginalFilename: att.Filename,
			},
			PaymentID: uuid.New().String(),
		}

		if firstResult == nil {
			firstResult = result
		}
	}

	return firstResult, nil
}

// attachment represents an extracted email attachment.
type attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// extractAttachments parses the multipart body and extracts supported attachments.
func extractAttachments(body io.Reader, boundary string) ([]attachment, error) {
	reader := multipart.NewReader(body, boundary)
	var attachments []attachment

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading multipart part: %w", err)
		}

		partContentType := part.Header.Get("Content-Type")
		if partContentType == "" {
			continue
		}

		parsedType, _, err := mime.ParseMediaType(partContentType)
		if err != nil {
			continue
		}

		if !supportedAttachmentTypes[parsedType] {
			continue
		}

		filename := part.FileName()
		if filename == "" {
			// Generate a filename based on content type
			ext := extensionForType(parsedType)
			filename = fmt.Sprintf("attachment%s", ext)
		}

		// Read the part data
		data, err := io.ReadAll(part)
		if err != nil {
			return nil, fmt.Errorf("error reading attachment data: %w", err)
		}

		// Check if the data is base64 encoded
		encoding := part.Header.Get("Content-Transfer-Encoding")
		if strings.EqualFold(encoding, "base64") {
			decoded, err := base64.StdEncoding.DecodeString(string(data))
			if err != nil {
				// Try with line breaks removed
				cleaned := strings.ReplaceAll(string(data), "\r\n", "")
				cleaned = strings.ReplaceAll(cleaned, "\n", "")
				decoded, err = base64.StdEncoding.DecodeString(cleaned)
				if err != nil {
					return nil, fmt.Errorf("failed to decode base64 attachment: %w", err)
				}
			}
			data = decoded
		}

		attachments = append(attachments, attachment{
			Filename:    filepath.Base(filename),
			ContentType: parsedType,
			Data:        data,
		})
	}

	return attachments, nil
}

// extractSender extracts the sender email address from the message headers.
func extractSender(msg *mail.Message) string {
	from := msg.Header.Get("From")
	if from == "" {
		return "unknown"
	}

	addr, err := mail.ParseAddress(from)
	if err != nil {
		// Fall back to raw value if parsing fails
		return from
	}
	return addr.Address
}

// extractMessageID extracts and cleans the Message-ID from headers.
func extractMessageID(msg *mail.Message) string {
	msgID := msg.Header.Get("Message-Id")
	if msgID == "" {
		msgID = msg.Header.Get("Message-ID")
	}
	// Remove angle brackets
	msgID = strings.TrimPrefix(msgID, "<")
	msgID = strings.TrimSuffix(msgID, ">")
	// Replace characters that aren't valid in S3 keys
	msgID = strings.ReplaceAll(msgID, "/", "_")
	msgID = strings.ReplaceAll(msgID, "\\", "_")
	return msgID
}

// extensionForType returns a file extension for a given MIME type.
func extensionForType(mimeType string) string {
	switch mimeType {
	case "application/pdf":
		return ".pdf"
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/tiff":
		return ".tiff"
	default:
		return ".bin"
	}
}
