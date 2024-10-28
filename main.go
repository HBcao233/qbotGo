package main

import (
	"os"
	"os/signal"

	"github.com/Logiase/MiraiGo-Template/client"

	_ "github.com/HBcao233/qbotGo/plugins/html"
	_ "github.com/HBcao233/qbotGo/plugins/twitter"
)

func main() {
	client.Init()
	client.Login()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
