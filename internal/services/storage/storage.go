package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Storage struct {
	DB *sqlx.DB
}

func New(storageDSN string) (*Storage, error) {
	db, err := sqlx.Connect("postgres", storageDSN)
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to database: %w", err)
	}

	return &Storage{
		DB: db,
	}, nil
}

func (s *Storage) SaveText(ctx context.Context, name, hashedUrl string) error {
	const op = "storage.SaveText"

	conn, err := s.DB.Connx(ctx)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(
		ctx,
		`INSERT INTO text_urls (name, hashed_url) VALUES($1, $2)`,
		name,
		hashedUrl,
	); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

func (s *Storage) ProvideHashedUrl(ctx context.Context, name string) (string, error) {
	const op = "storage.ProvideUrl"

	conn, err := s.DB.Connx(ctx)
	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}

	var hashedUrl string

	if err := conn.GetContext(ctx, &hashedUrl, `SELECT hashed_url FROM text_urls WHERE name=$1`, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("not found")
		}

		return "", fmt.Errorf("%s:%w", op, err)
	}

	return hashedUrl, nil
}

func (s *Storage) CheckNamePresence(ctx context.Context, name string) (bool, error) {
	const op = "storage.CheckHashExist"

	conn, err := s.DB.Connx(ctx)
	if err != nil {
		return true, fmt.Errorf("%s:%w", op, err)
	}
	defer conn.Close()

	var id int

	if err := conn.GetContext(ctx, &id, `SELECT id FROM text_urls WHERE name=$1`, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("%s:%w", op, err)
	}

	return true, nil
}

func (s *Storage) RemoveRecord(ctx context.Context, name string) {
	const op = "storage.RemoveRecord"

	conn, err := s.DB.Connx(ctx)
	if err != nil {
		return
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `DELETE FROM text_urls WHERE name=$1`, name); err != nil {
		return
	}

	return
}
