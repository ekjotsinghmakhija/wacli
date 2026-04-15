// internal/store/messages.go
package store

import (
	"time"
)

type UpsertMessageParams struct {
	ChatJID     string
	ChatName    string
	MsgID       string
	SenderJID   string
	SenderName  string
	Timestamp   time.Time
	FromMe      bool
	Text        string
	DisplayText string
	MediaType   string
}

func (d *DB) UpsertMessage(p UpsertMessageParams) error {
	_, err := d.sql.Exec(`
		INSERT INTO messages(
			chat_jid, chat_name, msg_id, sender_jid, sender_name, ts, from_me, text, display_text, media_type
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(chat_jid, msg_id) DO UPDATE SET
			chat_name=COALESCE(NULLIF(excluded.chat_name,''), messages.chat_name),
			sender_name=COALESCE(NULLIF(excluded.sender_name,''), messages.sender_name),
			display_text=CASE WHEN excluded.display_text IS NOT NULL AND excluded.display_text != '' THEN excluded.display_text ELSE messages.display_text END
	`, p.ChatJID, nullIfEmpty(p.ChatName), p.MsgID, nullIfEmpty(p.SenderJID), nullIfEmpty(p.SenderName),
		unix(p.Timestamp), boolToInt(p.FromMe), nullIfEmpty(p.Text), nullIfEmpty(p.DisplayText), nullIfEmpty(p.MediaType))
	return err
}

func (d *DB) GetMessage(chatJID, msgID string) (Message, error) {
	row := d.sql.QueryRow(`
		SELECT m.chat_jid, COALESCE(c.name,''), m.msg_id, COALESCE(m.sender_jid,''), m.ts, m.from_me, COALESCE(m.text,''), COALESCE(m.display_text,''), COALESCE(m.media_type,''), ''
		FROM messages m
		LEFT JOIN chats c ON c.jid = m.chat_jid
		WHERE m.chat_jid = ? AND m.msg_id = ?
	`, chatJID, msgID)

	var m Message
	var ts int64
	var fromMe int
	if err := row.Scan(&m.ChatJID, &m.ChatName, &m.MsgID, &m.SenderJID, &ts, &fromMe, &m.Text, &m.DisplayText, &m.MediaType, &m.Snippet); err != nil {
		return Message{}, err
	}
	m.Timestamp = fromUnix(ts)
	m.FromMe = fromMe != 0
	return m, nil
}
