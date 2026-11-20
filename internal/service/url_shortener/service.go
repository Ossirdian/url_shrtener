package urlshortener

import (
	"context"
	"errors"
	"math/rand"
	"net/url"
	"regexp"

	"github.com/ossirdian/url_shortener/internal/model"
	"github.com/ossirdian/url_shortener/internal/repository"
	"github.com/ossirdian/url_shortener/internal/service"
)

const (
	generatedAliasLength = 6
	maxAliasLength       = 32
	letters              = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var aliasPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type Service struct {
	repo repository.URLRepository
}

func NewService(repo repository.URLRepository) service.UrlShortenerService {
	return &Service{repo: repo}
}

func NewRandomString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func validateAlias(alias string) error {
	if alias == "" {
		return service.ErrInvalidArgument
	}
	if len(alias) > maxAliasLength {
		return service.ErrInvalidArgument
	}
	if !aliasPattern.MatchString(alias) {
		return service.ErrInvalidArgument
	}
	return nil
}

func (s *Service) Create(ctx context.Context, input model.URLCreate) (model.URL, error) {
	if !isValidURL(input.OriginalUrl) {
		return model.URL{}, service.ErrInvalidURL
	}
	if input.Alias != "" {
		if err := validateAlias(input.Alias); err != nil {
			return model.URL{}, err
		}
		return s.repo.Create(ctx, input)
	}

	aliasCreated := false
	for !aliasCreated {
		alias := NewRandomString(generatedAliasLength)

		_, err := s.repo.GetByAlias(ctx, alias)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return model.URL{}, err
		}
		if err == nil {
			continue
		}

		input.Alias = alias
		created, err := s.repo.Create(ctx, input)
		if errors.Is(err, repository.ErrConflict) {
			continue
		}
		if err != nil {
			return model.URL{}, err
		}

		aliasCreated = true
		return created, nil
	}

	return model.URL{}, repository.ErrConflict
}

func isValidURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

func (s *Service) Resolve(ctx context.Context, alias string) (model.URL, error) {
	if err := validateAlias(alias); err != nil {
		return model.URL{}, err
	}
	return s.repo.GetByAlias(ctx, alias)
}

func (s *Service) Update(ctx context.Context, input model.URLUpdate) (model.URL, error) {
	if err := validateAlias(input.OldAlias); err != nil {
		return model.URL{}, err
	}
	if err := validateAlias(input.NewAlias); err != nil {
		return model.URL{}, err
	}
	return s.repo.Update(ctx, input)
}

func (s *Service) Delete(ctx context.Context, alias string) error {
	if err := validateAlias(alias); err != nil {
		return err
	}
	return s.repo.Delete(ctx, alias)
}
