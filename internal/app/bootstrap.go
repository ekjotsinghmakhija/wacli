// internal/app/bootstrap.go
package app

import (
	"log"

	"github.com/ekjotsinghmakhija/wacli/internal/store"
	"github.com/ekjotsinghmakhija/wacli/internal/wa"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (a *App) ProcessHistorySync(evt *events.HistorySync) {
	history := evt.Data
	for _, conv := range history.GetConversations() {
		chatJID := conv.GetID()
		name := conv.GetName()

		_ = a.db.UpsertChat(chatJID, "history", name, wa.ConvertTime(conv.GetConversationTimestamp()))

		for _, historyMsg := range conv.GetMessages() {
			msg := historyMsg.GetMessage()
			if msg == nil || msg.Message == nil {
				continue
			}

			liveEvt := &events.Message{
				Info: types.MessageInfo{
					MessageSource: types.MessageSource{
						Chat:     wa.ParseJID(chatJID),
						Sender:   wa.ParseJID(msg.GetKey().GetParticipant()),
						IsFromMe: msg.GetKey().GetFromMe(),
					},
					ID:        msg.GetKey().GetID(),
					Timestamp: wa.ConvertTime(msg.GetMessageTimestamp()),
				},
				Message: msg.Message,
			}

			p := wa.ParseMessage(liveEvt)

			err := a.db.UpsertMessage(store.UpsertMessageParams{
				ChatJID:     p.ChatJID,
				ChatName:    name,
				MsgID:       p.MsgID,
				SenderJID:   p.SenderJID,
				Timestamp:   liveEvt.Info.Timestamp,
				FromMe:      p.FromMe,
				Text:        p.Text,
				DisplayText: p.Text,
				MediaType:   p.MediaType,
			})

			if err != nil {
				log.Printf("history sync upsert error: %v", err)
			}
		}
	}
}
