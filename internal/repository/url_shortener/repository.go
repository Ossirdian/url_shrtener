package urlshortener

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ossirdian/url_shortener/internal/model"
	def "github.com/ossirdian/url_shortener/internal/repository"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{pool: pool}
}

func (r *repository) Create(ctx context.Context, input model.URLCreate) (model.URL, error) {
	query, args, err := def.Builder().
		Insert("URL").
		Columns("original_url", "alias").
		Values(input.OriginalUrl, input.Alias).
		ToSql()
	if err != nil {
		return model.URL{}, err
	}

	if _, err := r.pool.Exec(ctx, query, args...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.URL{}, def.ErrConflict
		}
		return model.URL{}, err
	}

	return model.URL{
		Alias:       input.Alias,
		OriginalUrl: input.OriginalUrl,
	}, nil
}

func (r *repository) GetByAlias(ctx context.Context, alias string) (model.URL, error) {
	query, args, err := def.Builder().
		Select("alias", "original_url").
		From("URL").
		Where("alias = ?", alias).
		ToSql()
	if err != nil {
		return model.URL{}, err
	}

	var result model.URL
	if err := r.pool.QueryRow(ctx, query, args...).Scan(
		&result.Alias,
		&result.OriginalUrl,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.URL{}, def.ErrNotFound
		}
		return model.URL{}, err
	}

	return result, nil
}

func (r *repository) Delete(ctx context.Context, alias string) error {
	query, args, err := def.Builder().
		Delete("URL").
		Where("alias = ?", alias).
		ToSql()
	if err != nil {
		return err
	}

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return def.ErrNotFound
	}

	return nil
}

func (r *repository) Update(ctx context.Context, input model.URLUpdate) (model.URL, error) {
	query, args, err := def.Builder().
		Update("URL").
		Set("alias", input.NewAlias).
		Where("alias = ?", input.OldAlias).
		Suffix("RETURNING alias, original_url").
		ToSql()
	if err != nil {
		return model.URL{}, err
	}

	var result model.URL
	if err := r.pool.QueryRow(ctx, query, args...).Scan(
		&result.Alias,
		&result.OriginalUrl,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.URL{}, def.ErrConflict
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return model.URL{}, def.ErrNotFound
		}
		return model.URL{}, err
	}

	return result, nil
}
