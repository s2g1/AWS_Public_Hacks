package correspondence

import (
	"testing"
	"time"
)

func TestReviewQueue_AddAndGetPending(t *testing.T) {
	q := NewReviewQueue()

	item := ReviewItem{
		ID: "item-1",
		Correspondence: &GeneratedCorrespondence{
			Subject:     "Payment Approved",
			Body:        "Your payment has been approved.",
			Format:      OutputFormatEmailHTML,
			GeneratedAt: time.Now(),
			Status:      CorrespondenceStatusDraft,
		},
		CreatedAt: time.Now(),
	}

	q.Add(item)

	pending := q.GetPendingReviews()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending item, got %d", len(pending))
	}
	if pending[0].ID != "item-1" {
		t.Errorf("expected item ID 'item-1', got %s", pending[0].ID)
	}
	if pending[0].Status != ReviewStatusPending {
		t.Errorf("expected PENDING status, got %s", pending[0].Status)
	}
}

func TestReviewQueue_Approve(t *testing.T) {
	q := NewReviewQueue()

	q.Add(ReviewItem{
		ID: "item-2",
		Correspondence: &GeneratedCorrespondence{
			Subject: "Test",
			Body:    "Body",
		},
		CreatedAt: time.Now(),
	})

	err := q.Approve("item-2", "reviewer-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After approval, GetPendingReviews should return no items
	pending := q.GetPendingReviews()
	if len(pending) != 0 {
		t.Errorf("expected 0 pending items after approval, got %d", len(pending))
	}
}

func TestReviewQueue_Reject(t *testing.T) {
	q := NewReviewQueue()

	q.Add(ReviewItem{
		ID: "item-3",
		Correspondence: &GeneratedCorrespondence{
			Subject: "Test",
			Body:    "Body",
		},
		CreatedAt: time.Now(),
	})

	err := q.Reject("item-3", "Content inappropriate")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After rejection, GetPendingReviews should return no items
	pending := q.GetPendingReviews()
	if len(pending) != 0 {
		t.Errorf("expected 0 pending items after rejection, got %d", len(pending))
	}
}

func TestReviewQueue_ApproveNotFound(t *testing.T) {
	q := NewReviewQueue()

	err := q.Approve("nonexistent", "reviewer")
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestReviewQueue_RejectNotFound(t *testing.T) {
	q := NewReviewQueue()

	err := q.Reject("nonexistent", "reason")
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestReviewQueue_CannotApproveAlreadyApproved(t *testing.T) {
	q := NewReviewQueue()

	q.Add(ReviewItem{
		ID: "item-4",
		Correspondence: &GeneratedCorrespondence{
			Subject: "Test",
			Body:    "Body",
		},
		CreatedAt: time.Now(),
	})

	_ = q.Approve("item-4", "reviewer")
	err := q.Approve("item-4", "another-reviewer")
	if err == nil {
		t.Error("expected error when approving already-approved item")
	}
}

func TestReviewQueue_CannotRejectAlreadyRejected(t *testing.T) {
	q := NewReviewQueue()

	q.Add(ReviewItem{
		ID: "item-5",
		Correspondence: &GeneratedCorrespondence{
			Subject: "Test",
			Body:    "Body",
		},
		CreatedAt: time.Now(),
	})

	_ = q.Reject("item-5", "reason1")
	err := q.Reject("item-5", "reason2")
	if err == nil {
		t.Error("expected error when rejecting already-rejected item")
	}
}

func TestReviewQueue_MultiplePendingItems(t *testing.T) {
	q := NewReviewQueue()

	for i := 0; i < 5; i++ {
		q.Add(ReviewItem{
			ID: "item-" + string(rune('A'+i)),
			Correspondence: &GeneratedCorrespondence{
				Subject: "Test",
				Body:    "Body",
			},
			CreatedAt: time.Now(),
		})
	}

	// Approve one
	_ = q.Approve("item-A", "reviewer")

	pending := q.GetPendingReviews()
	if len(pending) != 4 {
		t.Errorf("expected 4 pending items, got %d", len(pending))
	}
}
