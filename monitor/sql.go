package monitor

import (
	"strings"

	"github.com/mailru/go-clickhouse"
)

const (
	queryCreateTableUpdates = `
	CREATE TABLE IF NOT EXISTS updates (
		date         Date,
		time         DateTime,
		id           UInt32,
		type         String,
		user_id      Int64,
		chat_id	     Int64,
		message_id   String,
		text         String,
		media        String,
		is_inline    UInt8,
		is_forwarded UInt8,
		is_reply     UInt8,
		was_edited   UInt8,
		result_id    String
	)
	ENGINE = MergeTree()
	ORDER BY (date, id, type, user_id)
	PARTITION BY toYYYYMM(date)`

	queryCreateTableLogs = `
	CREATE TABLE IF NOT EXISTS log (
		date         Date,
		time         DateTime,
		update_id    UInt32,
		level        String,
		message      String,
		payload      String
	)
	ENGINE = MergeTree()
	ORDER BY (date, update_id, level)
	PARTITION BY toYYYYMM(date)`

	queryInsertIntoUpdates = `
	INSERT INTO updates (
		date,
		time,
		id,
		type,
		user_id,
		chat_id,
		message_id,
		text,
		media,
		is_inline,
		is_forwarded,
		is_reply,
		was_edited,
		result_id
	) VALUES `

	queryInsertIntoLog = `
	INSERT INTO log (
		date,
		time,
		update_id,
		level,
		message,
		payload
	) VALUES `
)

func (m *Monitor) insertUpdates(updates []Update) error {
	const values = "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	var (
		stmt []string
		args []interface{}
	)
	for _, update := range updates {
		stmt = append(stmt, values)
		args = append(
			args,
			clickhouse.Date(update.Time),
			update.Time,
			update.ID,
			update.Type,
			update.UserID,
			update.ChatID,
			update.MessageID,
			update.Text,
			update.Media,
			update.IsInline,
			update.IsForwarded,
			update.IsReply,
			update.WasEdited,
			update.ResultID,
		)
	}

	query := queryInsertIntoUpdates + strings.Join(stmt, ",")
	_, err := m.db.Exec(query, args...)
	return err
}

func (m *Monitor) insertRecords(records []Record) error {
	const values = "(?, ?, ?, ?, ?, ?)"

	var (
		stmt []string
		args []interface{}
	)
	for _, record := range records {
		stmt = append(stmt, values)
		args = append(
			args,
			clickhouse.Date(record.Time),
			record.Time,
			record.UpdateID,
			record.Level,
			record.Message,
			record.Payload,
		)
	}

	query := queryInsertIntoLog + strings.Join(stmt, ",")
	_, err := m.db.Exec(query, args...)
	return err
}
