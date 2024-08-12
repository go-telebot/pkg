package monitor

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/mailru/go-clickhouse"
)

type (
	// Config is used to configure a monitor instance.
	Config struct {
		// URL is a ClickHouse URL monitor will try to connect.
		URL string

		// BufferSize is the maximum buffer size after which the buffer will be flushed.
		BufferSize int

		// TickPeriod is the interval flushing will fire at.
		TickPeriod time.Duration

		// Logger is used on custom logging. As an output pass stdout, a file, or any other writer.
		Logger *log.Logger
	}

	// Monitor delivers general analytics of your telebot Bot to the ClickHouse storage.
	Monitor struct {
		wg     *sync.WaitGroup
		flush  chan struct{}
		halt   chan struct{}
		db     *sql.DB
		bus    chan interface{}
		ticker *time.Ticker
		logger *log.Logger
	}

	// Update represents a single update that will be recorded.
	// Stores the most important and representative data possible.
	Update struct {
		// Date is a receiving date of the update in YYYYMMDD format.
		Date time.Time `json:"date"`

		// Time is a receiving full time of the update.
		Time time.Time `json:"time"`

		// ID is the update's unique identifier.
		ID uint `json:"id"`

		// Type is the update's type.
		//
		// One of:
		// 		message
		// 		callback
		// 		inline_query
		//		inline_result
		// 		shipping_query
		//		pre_checkout_query
		//		chat_member
		//
		Type string `json:"type"`

		// UserID is the update sender's ID.
		UserID int64 `json:"user_id"`

		// ChatID is an ID of the conversation the update belongs to.
		ChatID int64 `json:"chat_id"`

		// MessageID is a message ID the update relates to.
		MessageID string `json:"message_id"`

		// Text is a textual payload of the update.
		//
		//		Type                Content
		//		message             text or caption
		//		callback            data
		//		inline_query        query
		//		inline_result       query
		//		shipping_query      payload
		//		pre_checkout_query  payload
		//
		Text string `json:"text"`

		// Media is a message's media type.
		//
		// One of:
		//		(empty)
		//		animation
		//		audio
		//		document
		//		photo
		//		sticker
		//		video
		//		video_note
		//		voice
		//		contact
		//		dice
		//		poll
		//		venue
		//		location
		//
		Media string `json:"media,omitempty"` // message type only

		// IsInline shows whether a message the update relates to is inline.
		IsInline bool `json:"is_inline,omitempty"`

		// IsForwarded shows whether the message is forwarded.
		IsForwarded bool `json:"is_forwarded,omitempty"` // message type only

		// IsReply shows whether the message is replied to another message.
		IsReply bool `json:"is_reply,omitempty"` // message type only

		// WasEdited shows whether the message was sent or edited.
		WasEdited bool `json:"was_edited,omitempty"` // message type only

		// ResultID is an ID for the inline result that was chosen.
		ResultID string `json:"result_id,omitempty"` // inline_Result type only
	}

	// Record represents a single log record, associated with the update ID.
	Record struct {
		// Date is a date of the log record in YYYYMMDD format.
		Date time.Time

		// Time is a full time of the log record.
		Time time.Time

		// UpdateID is an ID of the update, log record associates to.
		UpdateID uint

		// Level is a log record level.
		Level string

		// Message is a log record message.
		Message string

		// Payload is an additional JSON data comes along with the message.
		Payload string
	}

	// M is a map shortcut, used for a payload field type.
	M = map[string]interface{}
)

// DB returns the sql.DB instance.
func (m *Monitor) DB() *sql.DB {
	return m.db
}

// Flush forces the batching worker to flush collected data.
func (m *Monitor) Flush() {
	m.wg.Add(1)
	m.flush <- struct{}{}
	m.wg.Wait()
}

// Close closes the monitor and stops batching worker completely.
func (m *Monitor) Close() {
	m.halt <- struct{}{}
}

// New forms a new monitor instance, connecting to the ClickHouse,
// and creating two primary tables if they are not exist.
func New(config Config) (*Monitor, error) {
	db, err := sql.Open("clickhouse", config.URL)
	if err != nil {
		return nil, fmt.Errorf("monitor: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("monitor: %w", err)
	}

	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if config.TickPeriod == 0 {
		config.TickPeriod = 5 * time.Second
	}

	m := &Monitor{
		db:     db,
		bus:    make(chan interface{}, config.BufferSize),
		ticker: time.NewTicker(config.TickPeriod),
		logger: config.Logger,
	}

	if err := m.createTables(); err != nil {
		return nil, fmt.Errorf("monitor: %w", err)
	}

	go m.startBatchWorker()
	return m, nil
}

func (m *Monitor) createTables() (err error) {
	_, err = m.db.Exec(queryCreateTableUpdates)
	if err != nil {
		return
	}

	_, err = m.db.Exec(queryCreateTableLogs)
	if err != nil {
		return
	}

	return
}

func (m *Monitor) startBatchWorker() {
	var buffer []interface{}

	insertBatch := func() {
		if err := m.insertBatch(buffer); err != nil {
			log.Println("monitor:", err)
		}
		buffer = buffer[:0]
	}

	defer insertBatch()

	for {
		select {
		case v := <-m.bus:
			buffer = append(buffer, v)
			if len(buffer) >= cap(m.bus) {
				insertBatch()
			}
		case <-m.ticker.C:
			insertBatch()
		case <-m.flush:
			insertBatch()
			m.wg.Done()
		case <-m.halt:
			m.Flush()
			return
		}
	}
}

func (m *Monitor) insertBatch(values []interface{}) error {
	if len(values) == 0 {
		return nil
	}

	var (
		updates []Update
		records []Record
	)
	for _, v := range values {
		switch vv := v.(type) {
		case Update:
			updates = append(updates, vv)
		case Record:
			records = append(records, vv)
		}
	}

	if len(updates) > 0 {
		if err := m.insertUpdates(updates); err != nil {
			return err
		}
	}
	if len(records) > 0 {
		if err := m.insertRecords(records); err != nil {
			return err
		}
	}

	return nil
}
