package health

import (
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
	Create(endpoint string) (Check, error)
	List(page int) (int, []Check)
}

type Repository interface {
	Create(check Check) error
	List(page, size int) (total int, checks []Check)
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

func (s *service) Create(endpoint string) (Check, error) {
	u, err := validateURL(endpoint)
	if err != nil {
		return Check{}, errInvalidEndpoint
	}

	id, err := newID(endpoint)
	if err != nil {
		return Check{}, errors.New("unexpected error")
	}

	newCheck := Check{
		ID:       id,
		Endpoint: u.String(),
	}
	if err := s.repo.Create(newCheck); err != nil {
		return Check{}, err
	}
	return newCheck, nil
}

func (s *service) List(page int) (int, []Check) {
	return s.repo.List(page, 10)
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
