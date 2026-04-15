// internal/wa/parse.go
package wa

import (
	"go.mau.fi/whatsmeow/events"
)

type ParsedMessage struct {
	ChatJID   string
	SenderJID string
	MsgID     string
	Text      string
	MediaType string
	FromMe    bool
}

func ParseMessage(evt *events.Message) ParsedMessage {
	msg := evt.Message
	var text, mediaType string

	if msg.GetConversation() != "" {
		text = msg.GetConversation()
		mediaType = "text"
	} else if msg.GetExtendedTextMessage() != nil {
		text = msg.GetExtendedTextMessage().GetText()
		mediaType = "text"
	} else if msg.GetImageMessage() != nil {
		text = msg.GetImageMessage().GetCaption()
		mediaType = "image"
	} else if msg.GetVideoMessage() != nil {
		text = msg.GetVideoMessage().GetCaption()
		mediaType = "video"
	} else if msg.GetDocumentMessage() != nil {
		text = msg.GetDocumentMessage().GetTitle()
		mediaType = "document"
	}

	return ParsedMessage{
		ChatJID:   evt.Info.Chat.ToNonAD().String(),
		SenderJID: evt.Info.Sender.ToNonAD().String(),
		MsgID:     evt.Info.ID,
		Text:      text,
		MediaType: mediaType,
		FromMe:    evt.Info.IsFromMe,
	}
}
