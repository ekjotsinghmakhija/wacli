package store

import (
	"testing"
	"time"
)

func TestSearchMessages_EdgeCases(t *testing.T) {
	db, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer db.Close()

	// 1. Seed
	_ = db.UpsertChat("user1@s.whatsapp.net", "individual", "Alice", time.Now())
	_ = db.UpsertMessage(UpsertMessageParams{
		ChatJID:     "user1@s.whatsapp.net",
		MsgID:       "MSG_1",
		Timestamp:   time.Now(),
		Text:        "Valid text message",
		DisplayText: "Valid text message",
	})

	// 2. Edge Case: Empty Query
	_, err = db.SearchMessages(SearchMessagesParams{Query: "  "})
	if err == nil {
		t.Error("expected error for empty query, got nil")
	}

	// 3. Edge Case: SQL Injection Attempt via Malformed FTS syntax
	// SQLite FTS5 panics on raw `NEAR(` without quotes. SanitizeFTSQuery must catch this.
	res, err := db.SearchMessages(SearchMessagesParams{Query: `NEAR(hello world)`})
	if err != nil {
		t.Errorf("expected no error on malformed FTS syntax (should be sanitized), got: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("expected 0 results for injection attempt, got %d", len(res))
	}

	// 4. Edge Case: Standard match
	res, err = db.SearchMessages(SearchMessagesParams{Query: "Valid"})
	if err != nil || len(res) != 1 {
		t.Errorf("expected 1 result for 'Valid', got %d (err: %v)", len(res), err)
	}
}
