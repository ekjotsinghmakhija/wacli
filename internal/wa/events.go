package wa

import (
	"go.mau.fi/whatsmeow/types/events"
)

// EventRouter defines how the application core will consume whatsmeow events.
// This interface allows the application to remain decoupled from whatsmeow internals.
type EventRouter interface {
	OnMessage(msg *events.Message)
	OnHistorySync(evt *events.HistorySync)
	OnReceipt(evt *events.Receipt)
	OnPresence(evt *events.Presence)
}

// BindRouter attaches an EventRouter to the WAClient.
// WARNING: Event handlers run synchronously within the whatsmeow socket loop.
// The EventRouter implementation MUST NOT block, or it will sever the connection.
func BindRouter(c WAClient, router EventRouter) uint32 {
	return c.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			router.OnMessage(v)
		case *events.HistorySync:
			router.OnHistorySync(v)
		case *events.Receipt:
			router.OnReceipt(v)
		case *events.Presence:
			router.OnPresence(v)
		}
	})
}
