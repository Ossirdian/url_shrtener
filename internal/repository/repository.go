package repository

import (
	"context"

	"github.com/ossirdian/url_shortener/internal/model"
)

type URLRepository interface {
	Create(ctx context.Context, input model.URLCreate) (model.URL, error)
	Delete(ctx context.Context, alias string) error
	GetByAlias(ctx context.Context, alias string) (model.URL, error)
	Update(ctx context.Context, input model.URLUpdate) (model.URL, error)
}
