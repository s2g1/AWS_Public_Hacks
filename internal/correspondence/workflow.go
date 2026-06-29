package correspondence

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CorrespondenceItem represents a single piece of correspondence in the review workflow.
type CorrespondenceItem struct {
	ID           string
	GeneratedCorrespondence
	SentAt       *time.Time
	ReviewedBy   string
	ModifiedBody string
}

// CorrespondenceQueue manages correspondence items through the review and send workflow.
type CorrespondenceQueue struct {
	mu    sync.RWMutex
	items map[string]*CorrespondenceItem
}

// NewCorrespondenceQueue creates a new in-memory correspondence queue.
func NewCorrespondenceQueue() *CorrespondenceQueue {
	return &CorrespondenceQueue{
		items: make(map[string]*CorrespondenceItem),
	}
}

// AddDraft adds a generated correspondence to the queue with DRAFT status.
func (q *CorrespondenceQueue) AddDraft(correspondence *GeneratedCorrespondence) *CorrespondenceItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	item := &CorrespondenceItem{
		ID: uuid.New().String(),
		GeneratedCorrespondence: GeneratedCorrespondence{
			Subject:     correspondence.Subject,
			Body:        correspondence.Body,
			Format:      correspondence.Format,
			GeneratedAt: correspondence.GeneratedAt,
			Status:      CorrespondenceStatusDraft,
		},
	}

	q.items[item.ID] = item
	return item
}

// MarkPendingReview transitions a correspondence item from DRAFT to PENDING_REVIEW.
func (q *CorrespondenceQueue) MarkPendingReview(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, exists := q.items[id]
	if !exists {
		return fmt.Errorf("correspondence item %s not found", id)
	}

	if item.Status != CorrespondenceStatusDraft {
		return fmt.Errorf("cannot transition from %s to PENDING_REVIEW: item must be in DRAFT status", item.Status)
	}

	item.Status = CorrespondenceStatusPendingReview
	return nil
}

// ApproveAndSend transitions a correspondence item to SENT, records the reviewer,
// and optionally applies a modified body.
func (q *CorrespondenceQueue) ApproveAndSend(id, reviewedBy string, modifiedBody string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, exists := q.items[id]
	if !exists {
		return fmt.Errorf("correspondence item %s not found", id)
	}

	if item.Status != CorrespondenceStatusPendingReview {
		return fmt.Errorf("cannot transition from %s to SENT: item must be in PENDING_REVIEW status", item.Status)
	}

	now := time.Now()
	item.Status = CorrespondenceStatusSent
	item.SentAt = &now
	item.ReviewedBy = reviewedBy
	if modifiedBody != "" {
		item.ModifiedBody = modifiedBody
	}

	return nil
}

// GetPending returns all correspondence items currently in PENDING_REVIEW status.
func (q *CorrespondenceQueue) GetPending() []*CorrespondenceItem {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var pending []*CorrespondenceItem
	for _, item := range q.items {
		if item.Status == CorrespondenceStatusPendingReview {
			pending = append(pending, item)
		}
	}
	return pending
}

// GetByID retrieves a correspondence item by its ID.
func (q *CorrespondenceQueue) GetByID(id string) *CorrespondenceItem {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.items[id]
}
