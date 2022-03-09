package main

import (
	"fmt"
	"github.com/HekapOo-hub/Task5/internal/service"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Info("start main")
	consumer := service.Consumer{}
	err := consumer.SetUp()
	if err != nil {
		return
	}
	log.Info("set up ended")
	defer consumer.Close()
	err = consumer.Consume()
	if err != nil {
		return
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	fmt.Println("received signal", <-c)
}
