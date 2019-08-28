package health

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"net/url"
)

type Check struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Code     int32  `json:"code"`
	Endpoint string `json:"endpoint"`
	Checked  int64  `json:"checked"`
	Duration string `json:"duration"`
}

type SVC interface {
	Create(ctx context.Context, endpoint string) (Check, error)
}

type Repository interface {
	Create(ctx context.Context, check Check) (Check, error)
}

type service struct {
	repo Repository
}

var _ SVC = (*service)(nil)

func NewSVC(repo Repository) SVC {
	return &service{
		repo: repo,
	}
}

var (
	errInvalidEndpoint = errors.New("endpoint must be a valid absolute URL")
)

func (s *service) Create(ctx context.Context, endpoint string) (Check, error) {
	u, err := validateURL(endpoint)
	if err != nil {
		return Check{}, errInvalidEndpoint
	}

	id, err := newID(endpoint)
	if err != nil {
		return Check{}, errors.New("unexpected error")
	}

	return s.repo.Create(ctx, Check{
		ID:       id,
		Endpoint: u.String(),
	})
}

func validateURL(endpoint string) (*url.URL, error) {
	if endpoint == "" {
		return nil, errInvalidEndpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, errInvalidEndpoint
	}
	if u.Host == "" {
		return nil, errInvalidEndpoint
	}
	return u, nil
}

func newID(endpoint string) (string, error) {
	h := md5.New()
	_, err := h.Write([]byte(endpoint))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum([]byte("tyrael"))), nil
}
