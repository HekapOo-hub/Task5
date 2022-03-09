package service

import (
	"encoding/json"
	"fmt"
	"github.com/HekapOo-hub/Task5/internal/config"
	"github.com/HekapOo-hub/Task5/internal/model"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"time"
)

type Producer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func (p *Producer) SetUp() error {
	cfg, err := config.GetRabbitConfig()
	if err != nil {
		log.Warnf("producer set up: %v", err)
		return fmt.Errorf("producer set up: %w", err)
	}
	p.conn, err = amqp.Dial(cfg.GetURL())
	if err != nil {
		log.Warnf("producer set up: %v", err)
		return fmt.Errorf("producer set up: %w", err)
	}

	p.ch, err = p.conn.Channel()
	if err != nil {
		log.Warnf("producer set up: %v", err)
		return fmt.Errorf("producer set up: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	err := p.ch.Close()
	log.Warnf("producer close: %v", err)
	err = p.conn.Close()
	if err != nil {
		log.Warnf("producer close: %v", err)
		return fmt.Errorf("producer close %w", err)
	}
	return nil
}
func (p *Producer) Publish() error {
	q, err := p.ch.QueueDeclare(
		"task_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Warnf("producer's publish: %v", err)
		return fmt.Errorf("producer's publish: %w", err)
	}
	start := time.Now()
	for i := 0; ; i++ {
		offset := i % 2
		var command string
		if offset == 0 {
			command = "create"
		} else {
			command = "delete"
		}
		msg := model.Message{Value: i - offset, Command: command}
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			log.Warnf("producer's publish %v", err)
			continue
		}
		err = p.ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         msgBytes,
			})
		if err != nil {
			log.Warnf("producer's publish: %v", err)
			return fmt.Errorf("producer's publish: %w", err)
		}
		if i%2000 == 0 {
			log.Infof("produced %d messages with speed %.2f/s\n", i, float64(i)/time.Since(start).Seconds())
			time.Sleep(time.Second)
		}
	}
}
