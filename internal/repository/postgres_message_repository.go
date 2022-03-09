package repository

import (
	"context"
	"fmt"
	"github.com/HekapOo-hub/Task5/internal/config"
	"github.com/HekapOo-hub/Task5/internal/model"
	"github.com/jackc/pgx"
	log "github.com/sirupsen/logrus"
)

type PostgresRepository struct {
	pool  *pgx.ConnPool
	batch *pgx.Batch
	tx    *pgx.Tx
}

func NewPostgresRepository() (*PostgresRepository, error) {
	cfg, err := config.GetPostgresConfig()
	if err != nil {
		log.Warnf("postgres config error: %v", err)
		return nil, fmt.Errorf("new postgres repository func %v", err)
	}

	log.Infof("p cfg %v", cfg)
	connConfig := pgx.ConnConfig{
		Host:     cfg.Host,
		Database: cfg.DBName,
		User:     cfg.UserName,
		Password: cfg.Password,
	}
	poolConfig := pgx.ConnPoolConfig{
		ConnConfig:     connConfig,
		MaxConnections: 3,
	}

	pool, err := pgx.NewConnPool(poolConfig)
	if err != nil {
		log.Warnf("creating postgres connection pool error %v", err)
		return nil, fmt.Errorf("creating postgres connection pool error %w", err)
	}
	_, err = pool.Prepare("create", "INSERT INTO messages (count) VALUES ($1)")
	if err != nil {
		return nil, fmt.Errorf("new postgres repository %w", err)
	}
	_, err = pool.Prepare("delete", "DELETE FROM messages WHERE count=$1")
	if err != nil {
		return nil, fmt.Errorf("new postgres repository %w", err)
	}

	return &PostgresRepository{pool: pool}, nil
}

func (r *PostgresRepository) Add(message model.Message) error {
	if r.batch == nil {
		var err error
		r.tx, err = r.pool.BeginEx(context.Background(), nil)
		if err != nil {
			return fmt.Errorf("postgres transaction %w", err)
		}

		r.batch = r.tx.BeginBatch()
	}

	r.batch.Queue(message.Command, []interface{}{message.Value}, nil, nil)

	return nil
}

func (r *PostgresRepository) SendBatch() error {
	err := r.batch.Send(context.Background(), nil)
	if err != nil {
		if e := r.tx.Rollback(); e != nil {
			log.Warnf("send batch %v", e)
		}

		if e := r.batch.Close(); e != nil {
			return fmt.Errorf("send batch %w", err)
		}
		return fmt.Errorf("send batch %w", err)
	}

	err = r.batch.Close()
	if err != nil {
		if e := r.tx.Rollback(); e != nil {
			return fmt.Errorf("send batch")
		}
		return fmt.Errorf("send batch %w", err)
	}
	err = r.tx.Commit()
	if err != nil {
		return fmt.Errorf("close tx %w", err)
	}
	return nil
}
func (r *PostgresRepository) OpenTx() error {
	var err error
	r.tx, err = r.pool.BeginEx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("open tx %w", err)
	}
	return nil
}
