package store

import (
	"fmt"
	"strings"
	"time"
)

func (d *DB) UpsertChat(jid, kind, name string, lastTS time.Time) error {
	if strings.TrimSpace(kind) == "" {
		kind = "unknown"
	}
	_, err := d.sql.Exec(`
		INSERT INTO chats(jid, kind, name, last_message_ts)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(jid) DO UPDATE SET
			kind=excluded.kind,
			name=CASE WHEN excluded.name IS NOT NULL AND excluded.name != '' THEN excluded.name ELSE chats.name END,
			last_message_ts=CASE WHEN excluded.last_message_ts > COALESCE(chats.last_message_ts, 0) THEN excluded.last_message_ts ELSE chats.last_message_ts END
	`, jid, kind, name, unix(lastTS))
	return err
}

func (d *DB) ListChats(query string, limit int) ([]Chat, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT jid, kind, COALESCE(name,''), COALESCE(last_message_ts,0) FROM chats WHERE 1=1`
	var args []interface{}

	if strings.TrimSpace(query) != "" {
		q += ` AND (LOWER(name) LIKE LOWER(?) OR LOWER(jid) LIKE LOWER(?))`
		needle := "%" + query + "%"
		args = append(args, needle, needle)
	}
	q += ` ORDER BY last_message_ts DESC LIMIT ?`
	args = append(args, limit)

	rows, err := d.sql.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Chat
	for rows.Next() {
		var c Chat
		var ts int64
		if err := rows.Scan(&c.JID, &c.Kind, &c.Name, &ts); err == nil {
			c.LastMessageTS = fromUnix(ts)
			out = append(out, c)
		}
	}
	return out, nil
}

func (d *DB) GetChat(jid string) (Chat, error) {
	row := d.sql.QueryRow(`SELECT jid, kind, COALESCE(name,''), COALESCE(last_message_ts,0) FROM chats WHERE jid = ?`, jid)
	var c Chat
	var ts int64
	if err := row.Scan(&c.JID, &c.Kind, &c.Name, &ts); err != nil {
		return Chat{}, err
	}
	c.LastMessageTS = fromUnix(ts)
	return c, nil
}

func (d *DB) UpsertContact(jid, phone, pushName, fullName, firstName, businessName string) error {
	now := time.Now().UTC().Unix()
	_, err := d.sql.Exec(`
		INSERT INTO contacts(jid, phone, push_name, full_name, first_name, business_name, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(jid) DO UPDATE SET
			phone=COALESCE(NULLIF(excluded.phone,''), contacts.phone),
			push_name=COALESCE(NULLIF(excluded.push_name,''), contacts.push_name),
			full_name=COALESCE(NULLIF(excluded.full_name,''), contacts.full_name),
			first_name=COALESCE(NULLIF(excluded.first_name,''), contacts.first_name),
			business_name=COALESCE(NULLIF(excluded.business_name,''), contacts.business_name),
			updated_at=excluded.updated_at
	`, jid, phone, pushName, fullName, firstName, businessName, now)
	return err
}

func (d *DB) UpsertGroup(jid, name, ownerJID string, created time.Time) error {
	now := time.Now().UTC().Unix()
	_, err := d.sql.Exec(`
		INSERT INTO groups(jid, name, owner_jid, created_ts, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(jid) DO UPDATE SET
			name=COALESCE(NULLIF(excluded.name,''), groups.name),
			owner_jid=COALESCE(NULLIF(excluded.owner_jid,''), groups.owner_jid),
			created_ts=COALESCE(NULLIF(excluded.created_ts,0), groups.created_ts),
			updated_at=excluded.updated_at
	`, jid, name, ownerJID, unix(created), now)
	return err
}
