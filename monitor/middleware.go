package monitor

import (
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

// Middleware returns telebot MiddlewareFunc ready to be used.
func (m *Monitor) Middleware() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			update, ok := newUpdate(c)
			if ok {
				m.bus <- update
			}
			return next(c)
		}
	}
}

// OnError returns telebot OnError function ready to be used.
func (m *Monitor) OnError() func(error, tele.Context) {
	return func(err error, c tele.Context) {
		m.Error(c, err.Error())
	}
}

func newUpdate(c tele.Context) (Update, bool) {
	utype := updateType(c)
	if utype == "" {
		return Update{}, false
	}

	sender := c.Sender()
	if sender == nil {
		return Update{}, false
	}

	msg := c.Message()
	if msg != nil && msg.IsService() {
		return Update{}, false
	}

	text := c.Text()
	if text == "" {
		text = c.Data()
	}

	update := Update{
		Time:   time.Now(),
		ID:     uint(c.Update().ID),
		Type:   utype,
		UserID: sender.ID,
		Text:   text,
	}

	if chat := c.Chat(); chat != nil {
		update.ChatID = chat.ID
	}

	if msg != nil {
		update.MessageID = strconv.Itoa(msg.ID)
		update.Media = updateMedia(msg)
		update.IsForwarded = msg.IsForwarded()
		update.IsReply = msg.IsReply()
	}

	if c.Update().EditedMessage != nil || c.Update().EditedChannelPost != nil {
		update.WasEdited = true
	}

	if clb := c.Callback(); clb != nil {
		update.Text = clb.Unique
		if clb.Data != "" {
			update.Text += "|" + clb.Data
		}

		update.IsInline = clb.IsInline()
		if update.IsInline {
			update.MessageID = clb.MessageID
		}
	}

	if ir := c.InlineResult(); ir != nil {
		update.IsInline = true
		update.MessageID = ir.MessageID
		update.ResultID = ir.ResultID
	}

	return update, true
}

func updateType(c tele.Context) string {
	switch {
	case c.Callback() != nil: // must be first
		return "callback"
	case c.Message() != nil:
		return "message"
	case c.Query() != nil:
		return "inline_query"
	case c.InlineResult() != nil:
		return "inline_result"
	case c.ShippingQuery() != nil:
		return "shipping_query"
	case c.PreCheckoutQuery() != nil:
		return "pre_checkout_query"
	case c.ChatMember() != nil:
		return "chat_member"
	}
	return ""
}

func updateMedia(msg *tele.Message) string {
	switch {
	case msg.Animation != nil:
		return "animation"
	case msg.Audio != nil:
		return "audio"
	case msg.Document != nil:
		return "document"
	case msg.Photo != nil:
		return "photo"
	case msg.Sticker != nil:
		return "sticker"
	case msg.Video != nil:
		return "video"
	case msg.VideoNote != nil:
		return "video_note"
	case msg.Voice != nil:
		return "voice"
	case msg.Contact != nil:
		return "contact"
	case msg.Dice != nil:
		return "dice"
	case msg.Poll != nil:
		return "poll"
	case msg.Venue != nil:
		return "venue"
	case msg.Location != nil:
		return "location"
	}
	return ""
}
