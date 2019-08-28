package health_test

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jsteenb2/health/internal/health"
)

func TestService(t *testing.T) {
	validateID := func(t *testing.T, endpoint string, got string) {
		t.Helper()

		// validates that IDs' area created in this fashion
		// test driven here why there is going to be duplicate
		// code in service.

		h := md5.New()
		_, err := h.Write([]byte(endpoint))
		mustNoError(t, err)

		expected := fmt.Sprintf("%x", h.Sum([]byte("tyrael")))
		if expected != got {
			t.Errorf("unexpected hash value: expected=%q got %q", expected, got)
		}
	}

	t.Run("create", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			repo := &fakeRepo{
				createFn: func(check health.Check) error { return nil },
			}
			svc := health.NewSVC(repo)

			endpoint := "http://www.example.com"
			c, err := svc.Create(endpoint)
			mustNoError(t, err)

			equal(t, endpoint, c.Endpoint, "invalid endpoint")
			equal(t, "Created", c.Status, "invalid status")
			validateID(t, endpoint, c.ID)
		})

		t.Run("invalid urls ", func(t *testing.T) {
			tests := []struct {
				name     string
				endpoint string
			}{
				{name: "empty", endpoint: ""},
				{name: "rando", endpoint: "$%2w.olo3threeve"},
				{name: "relative", endpoint: "/en-us"},
				{name: "missing host domain", endpoint: "http:///threeve"},
			}

			for _, tt := range tests {
				fn := func(t *testing.T) {
					repo := &fakeRepo{
						createFn: func(check health.Check) error { return nil },
					}
					svc := health.NewSVC(repo)

					_, err := svc.Create(tt.endpoint)
					mustError(t, err)
				}

				t.Run(tt.name, fn)
			}
		})

		t.Run("repo throws an error on creation", func(t *testing.T) {
			expectedErr := errors.New("rando create error here")
			repo := &fakeRepo{
				createFn: func(check health.Check) error {
					return expectedErr
				},
			}
			svc := health.NewSVC(repo)

			_, err := svc.Create("http://example.com")
			equal(t, expectedErr, err, "did not receive expected repo error")
		})
	})

	t.Run("list", func(t *testing.T) {
		repo := &fakeRepo{
			listFn: func(page, size int) (int, []health.Check) {
				return page, []health.Check{{ID: "id", Checked: int64(size)}}
			},
		}

		svc := health.NewSVC(repo)

		total, currentPage, checks := svc.List(1)

		equal(t, 1, total, "unexpected total")
		equal(t, 1, currentPage, "unexpected page number")
		mustEqual(t, 1, len(checks), "unexpected num of checks")
		equal(t, "id", checks[0].ID, "unexpected id")
		equal(t, int64(10), checks[0].Checked, "unexpected checked")
	})

	t.Run("read", func(t *testing.T) {
		t.Run("when a valid id for an existing item is provided should return the endpoint check", func(t *testing.T) {
			repo := &fakeRepo{
				readFn: func(id string) (health.Check, error) {
					return health.Check{ID: id}, nil
				},
			}

			svc := health.NewSVC(repo)

			// len 44 string required as that is the length of the md5 sum
			// of the given endpoints
			id := strings.Repeat("a", 44)
			check, err := svc.Read(id)
			mustNoError(t, err)

			equal(t, id, check.ID, "unexpected id")
		})

		t.Run("when an invalid id is provided should return an error", func(t *testing.T) {
			repo := &fakeRepo{
				readFn: func(id string) (health.Check, error) {
					return health.Check{ID: id}, nil
				},
			}

			svc := health.NewSVC(repo)

			id := "invalid id must be length 44"
			_, err := svc.Read(id)
			mustError(t, err)
		})
	})
}

type fakeRepo struct {
	createFn func(check health.Check) error
	listFn   func(page, size int) (int, []health.Check)
	readFn   func(id string) (health.Check, error)
}

func (f *fakeRepo) Create(check health.Check) error {
	if f.createFn == nil {
		panic("no createFn set")
	}
	return f.createFn(check)
}

func (f *fakeRepo) List(page, size int) (int, []health.Check) {
	if f.listFn == nil {
		panic("not implemented")
	}
	return f.listFn(page, size)
}

func (f *fakeRepo) Read(id string) (health.Check, error) {
	if f.readFn == nil {
		panic("not implemented yet")
	}
	return f.readFn(id)
}
