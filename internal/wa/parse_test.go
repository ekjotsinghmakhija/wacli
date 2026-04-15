package wa

import (
	"testing"
	"time"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/events"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func TestParseMessage_EdgeCases(t *testing.T) {
	senderJID, _ := types.ParseJID("user@s.whatsapp.net")
	chatJID, _ := types.ParseJID("group@g.us")

	baseInfo := events.MessageInfo{
		ID:        "TEST_ID",
		Sender:    senderJID,
		Chat:      chatJID,
		IsFromMe:  false,
		Timestamp: time.Now(),
	}

	tests := []struct {
		name     string
		msg      *waProto.Message
		wantText string
		wantType string
	}{
		{
			name:     "Nil Message Payload",
			msg:      &waProto.Message{},
			wantText: "",
			wantType: "",
		},
		{
			name: "Text Message",
			msg: &waProto.Message{
				Conversation: proto.String("hello"),
			},
			wantText: "hello",
			wantType: "text",
		},
		{
			name: "Image without Caption",
			msg: &waProto.Message{
				ImageMessage: &waProto.ImageMessage{},
			},
			wantText: "",
			wantType: "image",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			evt := &events.Message{
				Info:    baseInfo,
				Message: tc.msg,
			}
			parsed := ParseMessage(evt)

			if parsed.Text != tc.wantText {
				t.Errorf("Text: got %q, want %q", parsed.Text, tc.wantText)
			}
			if parsed.MediaType != tc.wantType {
				t.Errorf("MediaType: got %q, want %q", parsed.MediaType, tc.wantType)
			}
		})
	}
}
