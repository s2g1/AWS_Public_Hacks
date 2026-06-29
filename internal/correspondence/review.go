package correspondence

import (
	"fmt"
	"sync"
	"time"
)

// ReviewStatus represents the review state of a correspondence review item.
type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "PENDING"
	ReviewStatusApproved ReviewStatus = "APPROVED"
	ReviewStatusRejected ReviewStatus = "REJECTED"
)

// ReviewItem represents a correspondence item in the review queue with review metadata.
type ReviewItem struct {
	ID              string                   `json:"id"`
	Correspondence  *GeneratedCorrespondence `json:"correspondence"`
	CreatedAt       time.Time                `json:"createdAt"`
	ReviewedAt      *time.Time               `json:"reviewedAt,omitempty"`
	ReviewedBy      string                   `json:"reviewedBy,omitempty"`
	Status          ReviewStatus             `json:"status"`
	RejectionReason string                   `json:"rejectionReason,omitempty"`
}

// ReviewQueue manages an in-memory queue of correspondence items pending review.
type ReviewQueue struct {
	mu    sync.Mutex
	items []ReviewItem
}

// NewReviewQueue creates a new empty ReviewQueue.
func NewReviewQueue() *ReviewQueue {
	return &ReviewQueue{
		items: []ReviewItem{},
	}
}

// Add adds a correspondence item to the review queue with PENDING status.
func (q *ReviewQueue) Add(item ReviewItem) {
	q.mu.Lock()
	defer q.mu.Unlock()

	item.Status = ReviewStatusPending
	q.items = append(q.items, item)
}

// GetPendingReviews returns all correspondence items with PENDING status.
func (q *ReviewQueue) GetPendingReviews() []ReviewItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	var pending []ReviewItem
	for _, item := range q.items {
		if item.Status == ReviewStatusPending {
			pending = append(pending, item)
		}
	}
	return pending
}

// Approve marks a correspondence item as approved by the given reviewer.
func (q *ReviewQueue) Approve(id, reviewedBy string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, item := range q.items {
		if item.ID == id {
			if item.Status != ReviewStatusPending {
				return fmt.Errorf("item %s is not in PENDING status (current: %s)", id, item.Status)
			}
			now := time.Now()
			q.items[i].Status = ReviewStatusApproved
			q.items[i].ReviewedAt = &now
			q.items[i].ReviewedBy = reviewedBy
			return nil
		}
	}
	return fmt.Errorf("item %s not found", id)
}

// Reject marks a correspondence item as rejected with a reason.
func (q *ReviewQueue) Reject(id, reason string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, item := range q.items {
		if item.ID == id {
			if item.Status != ReviewStatusPending {
				return fmt.Errorf("item %s is not in PENDING status (current: %s)", id, item.Status)
			}
			now := time.Now()
			q.items[i].Status = ReviewStatusRejected
			q.items[i].ReviewedAt = &now
			q.items[i].RejectionReason = reason
			return nil
		}
	}
	return fmt.Errorf("item %s not found", id)
}
