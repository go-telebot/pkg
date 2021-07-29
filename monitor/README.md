# 📟 Monitor

```go
package main

import (
	"github.com/go-telebot/pkg/monitor"

	tele "gopkg.in/telebot.v3"
)

func main() {
	bot, _ := tele.NewBot(tele.Settings{...})
	mon, _ := monitor.New(monitor.Config{...})

	// Register middleware that sends every incoming update
	// to the ClickHouse instance configured in monitor.Config.
	bot.Use(mon.Middleware())

	// All the errors will be logged as well.
	bot.OnError = mon.OnError()

	// This is an example of custom logging, in case you want
	// to log and lately analyze any specific information,
	// which will be associated with the context.
	bot.Handle("/start", func(c tele.Context) error {
		mon.Info(c, "New user", monitor.M{
			"user_id":  c.Sender().ID,
			"user_ref": c.Message().Payload,
		})

		// You can also use:
		// mon.Debug
		// mon.Warn
		// mon.Error
	})
	
	// ...
}
```