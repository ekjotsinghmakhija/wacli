// internal/app/sync.go
package app

import (
	"log"

	"github.com/ekjotsinghmakhija/wacli/internal/store"
	"github.com/ekjotsinghmakhija/wacli/internal/wa"
	"go.mau.fi/whatsmeow/events"
)

func (a *App) OnMessage(evt *events.Message) {
	p := wa.ParseMessage(evt)

	err := a.db.UpsertMessage(store.UpsertMessageParams{
		ChatJID:     p.ChatJID,
		MsgID:       p.MsgID,
		SenderJID:   p.SenderJID,
		Timestamp:   evt.Info.Timestamp,
		FromMe:      p.FromMe,
		Text:        p.Text,
		DisplayText: p.Text,
		MediaType:   p.MediaType,
	})

	if err != nil {
		log.Printf("sync error: upsert message: %v", err)
	}

	_ = a.db.UpsertChat(p.ChatJID, "unknown", "", evt.Info.Timestamp)
}

func (a *App) OnHistorySync(evt *events.HistorySync) {
	a.ProcessHistorySync(evt)
}

func (a *App) OnReceipt(evt *events.Receipt) {
	// Action -> Update message status
}

func (a *App) OnPresence(evt *events.Presence) {
	// Action -> Update contact state
}
