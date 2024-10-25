package main

import (
	"os"
	"os/signal"

	"github.com/Logiase/MiraiGo-Template/bot"

	_ "github.com/HBcao233/qbotGo/plugins/twitter"
)

func main() {
	bot.Init()
	bot.Login()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
