package repository

import (
	"errors"

	"github.com/jackc/pgconn"
)

var (
	ErrNotFound = errors.New("url not found")
	ErrConflict = errors.New("url already exists")
)

func NormalizeError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505", "23503", "23514":
			return ErrConflict
		}
	}

	return err
}
