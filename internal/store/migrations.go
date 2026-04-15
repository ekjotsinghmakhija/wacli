package store

import (
	"fmt"
	"time"
)

type migration struct {
	version int
	name    string
	up      func(*DB) error
}

var schemaMigrations = []migration{
	{version: 1, name: "core schema", up: migrateCoreSchema},
	{version: 2, name: "messages fts", up: migrateMessagesFTS},
}

func (d *DB) ensureSchema() error {
	if _, err := d.sql.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at INTEGER NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	applied := map[int]bool{}
	rows, err := d.sql.Query(`SELECT version FROM schema_migrations`)
	if err == nil {
		for rows.Next() {
			var version int
			if rows.Scan(&version) == nil {
				applied[version] = true
			}
		}
		rows.Close()
	}

	for _, m := range schemaMigrations {
		if applied[m.version] {
			continue
		}
		if err := m.up(d); err != nil {
			return fmt.Errorf("apply migration %03d: %w", m.version, err)
		}
		_, _ = d.sql.Exec(`INSERT INTO schema_migrations(version, name, applied_at) VALUES(?, ?, ?)`, m.version, m.name, time.Now().UTC().Unix())
	}
	return nil
}

func migrateCoreSchema(d *DB) error {
	_, err := d.sql.Exec(`
		CREATE TABLE IF NOT EXISTS chats (
			jid TEXT PRIMARY KEY,
			kind TEXT NOT NULL,
			name TEXT,
			last_message_ts INTEGER
		);

		CREATE TABLE IF NOT EXISTS messages (
			rowid INTEGER PRIMARY KEY AUTOINCREMENT,
			chat_jid TEXT NOT NULL,
			chat_name TEXT,
			msg_id TEXT NOT NULL,
			sender_jid TEXT,
			sender_name TEXT,
			ts INTEGER NOT NULL,
			from_me INTEGER NOT NULL,
			text TEXT,
			display_text TEXT,
			media_type TEXT,
			UNIQUE(chat_jid, msg_id),
			FOREIGN KEY (chat_jid) REFERENCES chats(jid) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_messages_chat_ts ON messages(chat_jid, ts);
		CREATE INDEX IF NOT EXISTS idx_messages_ts ON messages(ts);
	`)
	return err
}

func migrateMessagesFTS(d *DB) error {
	_, err := d.sql.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
			text,
			chat_name,
			sender_name,
			display_text
		);

		CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
			INSERT INTO messages_fts(rowid, text, chat_name, sender_name, display_text)
			VALUES (new.rowid, COALESCE(new.text,''), COALESCE(new.chat_name,''), COALESCE(new.sender_name,''), COALESCE(new.display_text,''));
		END;

		CREATE TRIGGER IF NOT EXISTS messages_ad AFTER DELETE ON messages BEGIN
			DELETE FROM messages_fts WHERE rowid = old.rowid;
		END;

		CREATE TRIGGER IF NOT EXISTS messages_au AFTER UPDATE ON messages BEGIN
			DELETE FROM messages_fts WHERE rowid = old.rowid;
			INSERT INTO messages_fts(rowid, text, chat_name, sender_name, display_text)
			VALUES (new.rowid, COALESCE(new.text,''), COALESCE(new.chat_name,''), COALESCE(new.sender_name,''), COALESCE(new.display_text,''));
		END;
	`)
	if err == nil {
		d.ftsEnabled = true
	}
	return nil
}
