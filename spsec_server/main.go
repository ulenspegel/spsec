package main

import (
	"log"
    
	"spsec/app"
	"spsec/bot"
	"spsec/config"
	"spsec/logger"
	"spsec/mode"
	"spsec/serv"
)

func main() {
    logRing := logger.NewLogger(5, config.LogFileName, 5*1024*1024)
    srv := serv.NewServer()

    tg, err := bot.NewBot(config.ApiKey)
    if err != nil {
        log.Fatal(err)
    }

    sm := mode.NewSilent()

    a := app.New(logRing, srv, tg, sm)
    a.Run()
}
