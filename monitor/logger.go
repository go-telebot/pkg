package monitor

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
)

func (m *Monitor) Info(c tele.Context, msg string, payload ...M) {
	m.log(c, "info", msg, payload...)
}

func (m *Monitor) Debug(c tele.Context, msg string, payload ...M) {
	m.log(c, "debug", msg, payload...)
}

func (m *Monitor) Warn(c tele.Context, msg string, payload ...M) {
	m.log(c, "warn", msg, payload...)
}

func (m *Monitor) Error(c tele.Context, msg string, payload ...M) {
	m.log(c, "error", msg, payload...)
}

func (m *Monitor) log(c tele.Context, level, msg string, payload ...M) {
	record := Record{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
	}
	if c != nil {
		record.UpdateID = uint(c.Update().ID)
	}

	var data []byte
	if len(payload) > 0 {
		data, _ = json.Marshal(payload[0])
		record.Payload = string(data)
	}

	if m.logger != nil {
		v := []interface{}{
			strings.ToUpper(level),
			msg,
		}
		if data != nil {
			var buf bytes.Buffer
			json.Indent(&buf, data, "", "  ")
			v = append(v, buf.String())
		}
		m.logger.Println(v...)
	}

	m.bus <- record
}
