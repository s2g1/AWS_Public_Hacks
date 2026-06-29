package models

// DocumentType classifies the type of document being processed.
type DocumentType string

const (
	DocumentTypeInvoice         DocumentType = "INVOICE"
	DocumentTypePurchaseOrder   DocumentType = "PURCHASE_ORDER"
	DocumentTypeTravelVoucher   DocumentType = "TRAVEL_VOUCHER"
	DocumentTypeGrantPayment    DocumentType = "GRANT_PAYMENT"
	DocumentTypeContractPayment DocumentType = "CONTRACT_PAYMENT"
	DocumentTypeUnknown         DocumentType = "UNKNOWN"
)

// RequiredFieldsByDocType maps each document type to its required fields.
var RequiredFieldsByDocType = map[DocumentType][]string{
	DocumentTypeInvoice:         {"payee", "amount", "invoiceNumber", "date"},
	DocumentTypePurchaseOrder:   {"vendor", "items", "totalAmount", "poNumber"},
	DocumentTypeTravelVoucher:   {"traveler", "dates", "expenses", "totalClaim"},
	DocumentTypeGrantPayment:    {"payee", "amount", "grantNumber", "date"},
	DocumentTypeContractPayment: {"payee", "amount", "contractNumber", "date"},
	DocumentTypeUnknown:         {},
}

// BoundingBox represents the location of an extracted field within a document.
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ExtractedField represents a single field extracted from a document with its confidence.
type ExtractedField struct {
	Value         string       `json:"value"`
	Confidence    float64      `json:"confidence"`
	BoundingBox   *BoundingBox `json:"boundingBox,omitempty"`
	Normalized    string       `json:"normalized"`
	IsHandwritten bool         `json:"isHandwritten,omitempty"`
}

// ExtractionResult contains the full output of the document extraction agent.
type ExtractionResult struct {
	DocumentType      DocumentType            `json:"documentType"`
	Fields            map[string]ExtractedField `json:"fields"`
	OverallConfidence float64                 `json:"overallConfidence"`
	RawText           string                  `json:"rawText"`
	ProcessingTimeMs  int64                   `json:"processingTimeMs"`
}
