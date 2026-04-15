package store

import (
	"fmt"
	"strings"
	"time"
)

type SearchMessagesParams struct {
	Query   string
	ChatJID string
	Limit   int
	After   *time.Time
}

func (d *DB) SearchMessages(p SearchMessagesParams) ([]Message, error) {
	if strings.TrimSpace(p.Query) == "" {
		return nil, fmt.Errorf("query is required")
	}
	if p.Limit <= 0 {
		p.Limit = 50
	}

	if d.ftsEnabled {
		return d.searchFTS(p)
	}
	return d.searchLIKE(p)
}

func sanitizeFTSQuery(q string) string {
	tokens := strings.Fields(q)
	if len(tokens) == 0 {
		return `""`
	}
	quoted := make([]string, len(tokens))
	for i, tok := range tokens {
		quoted[i] = `"` + strings.ReplaceAll(tok, `"`, `""`) + `"`
	}
	return strings.Join(quoted, " ")
}

func (d *DB) searchFTS(p SearchMessagesParams) ([]Message, error) {
	query := `
		SELECT m.chat_jid, COALESCE(c.name,''), m.msg_id, COALESCE(m.sender_jid,''), m.ts, m.from_me, COALESCE(m.text,''), COALESCE(m.display_text,''), COALESCE(m.media_type,''),
		       snippet(messages_fts, 0, '[', ']', '…', 12)
		FROM messages_fts
		JOIN messages m ON messages_fts.rowid = m.rowid
		LEFT JOIN chats c ON c.jid = m.chat_jid
		WHERE messages_fts MATCH ?`

	args := []interface{}{sanitizeFTSQuery(p.Query)}

	if strings.TrimSpace(p.ChatJID) != "" {
		query += " AND m.chat_jid = ?"
		args = append(args, p.ChatJID)
	}
	if p.After != nil {
		query += " AND m.ts > ?"
		args = append(args, unix(*p.After))
	}

	query += " ORDER BY bm25(messages_fts) LIMIT ?"
	args = append(args, p.Limit)

	return d.scanMessages(query, args...)
}

func (d *DB) searchLIKE(p SearchMessagesParams) ([]Message, error) {
	query := `
		SELECT m.chat_jid, COALESCE(c.name,''), m.msg_id, COALESCE(m.sender_jid,''), m.ts, m.from_me, COALESCE(m.text,''), COALESCE(m.display_text,''), COALESCE(m.media_type,''), ''
		FROM messages m
		LEFT JOIN chats c ON c.jid = m.chat_jid
		WHERE (LOWER(m.text) LIKE LOWER(?) OR LOWER(m.display_text) LIKE LOWER(?))`

	needle := "%" + strings.ReplaceAll(p.Query, "%", "\\%") + "%"
	args := []interface{}{needle, needle}

	if strings.TrimSpace(p.ChatJID) != "" {
		query += " AND m.chat_jid = ?"
		args = append(args, p.ChatJID)
	}

	query += " ORDER BY m.ts DESC LIMIT ?"
	args = append(args, p.Limit)

	return d.scanMessages(query, args...)
}

func (d *DB) scanMessages(query string, args ...interface{}) ([]Message, error) {
	rows, err := d.sql.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Message
	for rows.Next() {
		var m Message
		var ts int64
		var fromMe int
		if err := rows.Scan(&m.ChatJID, &m.ChatName, &m.MsgID, &m.SenderJID, &ts, &fromMe, &m.Text, &m.DisplayText, &m.MediaType, &m.Snippet); err == nil {
			m.Timestamp = fromUnix(ts)
			m.FromMe = fromMe != 0
			out = append(out, m)
		}
	}
	return out, nil
}
