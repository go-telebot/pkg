# 🤖 Telebot Tools

> `go get github.com/go-telebot/pkg`

This repository holds a number of useful tools for [Telebot V3](https://github.com/tucnak/telebot/tree/v3) framework.

## [Monitor](https://github.com/go-telebot/pkg/tree/main/monitor)

`Monitor` is a middleware, which collects and stores important and representative analytics of the bot, records almost every incoming update in the highly efficient and fast ClickHouse storage. It also allows a developer to log its own custom data with JSON-formatted payload.

### Integration
```go
mon, err := monitor.New(monitor.Config{...})

b.OnError = mon.OnError()
b.Use(mon.Middleware())

// And you're ready to go!
```

## [Telegraph](https://github.com/go-telebot/pkg/tree/main/telegraph)

`Telegraph` is a simple package to upload files to Telegraph.

### Usage
```go
telegraph.UploadFile("image.png")
telegraph.Upload(reader)
```