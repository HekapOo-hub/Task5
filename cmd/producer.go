package main

import (
	"fmt"
	"github.com/HekapOo-hub/Task5/internal/service"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	producer := service.Producer{}
	err := producer.SetUp()
	if err != nil {
		return
	}
	defer producer.Close()
	err = producer.Publish()
	if err != nil {
		return
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	fmt.Println("received signal", <-c)
}
