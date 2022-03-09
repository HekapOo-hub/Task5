package service

import (
	"fmt"
	"github.com/HekapOo-hub/Task5/internal/config"
	"github.com/HekapOo-hub/Task5/internal/model"
	"github.com/HekapOo-hub/Task5/internal/repository"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type Consumer struct {
	conn       *amqp.Connection
	ch         *amqp.Channel
	repository *repository.PostgresRepository
}

func (c *Consumer) Close() error {
	err := c.repository.SendBatch()
	if err != nil {
		log.Warnf("consumer close: %v", err)
		return fmt.Errorf("consumer close: %w", err)
	}
	err = c.ch.Close()
	if err != nil {
		log.Warnf("consumer close: %v", err)
	}
	err = c.conn.Close()
	if err != nil {
		log.Warnf("consumer close: %v", err)
		return fmt.Errorf("consumer close %w", err)
	}
	return nil
}

func (c *Consumer) SetUp() error {
	var err error
	c.repository, err = repository.NewPostgresRepository()
	if err != nil {
		log.Warnf("consumer set up: %v", err)
		return fmt.Errorf("consumer set up: %w", err)
	}
	cfg, err := config.GetRabbitConfig()
	if err != nil {
		log.Warnf("producer set up: %v", err)
		return fmt.Errorf("producer set up: %w", err)
	}
	c.conn, err = amqp.Dial(cfg.GetURL())
	if err != nil {
		log.Warnf("producer set up: %v", err)
		return fmt.Errorf("producer set up: %w", err)
	}
	c.ch, err = c.conn.Channel()
	if err != nil {
		log.Warnf("comsumer set up: %v", err)
		return fmt.Errorf("consumer set up: %w", err)
	}
	return nil
}

func (c *Consumer) Consume() error {
	q, err := c.ch.QueueDeclare(
		"task_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Warnf("consume: %v", err)
		return fmt.Errorf("consume %w", err)
	}

	err = c.ch.Qos(
		2000,  // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Warnf("consume: %v", err)
		return fmt.Errorf("consume %w", err)
	}
	messages, err := c.ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Warnf("consume: %v", err)
		return fmt.Errorf("consume %w", err)
	}
	go func() {
		log.Info("start consuming...")
		var count int
		for d := range messages {
			log.Printf("Received a message: %s", d.Body)
			modelMsg, err := model.DecodeMessage(d.Body)
			if err != nil {
				log.Warnf("consume: %v", err)
				continue
			}
			err = c.repository.Add(*modelMsg)
			if err != nil {
				log.Warnf("consume: %v", err)
				continue
			}
			count++
			if count%2000 == 0 {
				count = 0
				err := c.repository.SendBatch()
				if err != nil {
					log.Warnf("consume claim handler %v", err)
				}
				err = d.Ack(true)
				if err != nil {
					log.Warnf("consume: %v", err)
					return
				}
				/*err = c.repository.OpenTx()
				if err != nil {
					log.Warnf("consume claim %v", err)
				}*/
			}
		}
	}()
	return nil
}
