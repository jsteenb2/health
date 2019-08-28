package health_test

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"testing"

	"github.com/jsteenb2/health/internal/health"
)

func TestService(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			repo := &fakeRepo{
				createFn: func(ctx context.Context, check health.Check) (health.Check, error) {
					return check, nil
				},
			}
			svc := health.NewSVC(repo)

			endpoint := "http://www.example.com"
			c, err := svc.Create(context.TODO(), endpoint)
			if err != nil {
				t.Fatalf("unexpected err: got=%s", err.Error())
			}

			equal(t, endpoint, c.Endpoint, "invalid endpoint")
			t.Log(c.ID)
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
						createFn: func(ctx context.Context, check health.Check) (health.Check, error) {
							return check, nil
						},
					}
					svc := health.NewSVC(repo)

					_, err := svc.Create(context.TODO(), tt.endpoint)
					if err == nil {
						t.Fatal("did not receive expected error")
					}
				}

				t.Run(tt.name, fn)
			}
		})

		t.Run("repo throws an error on creation", func(t *testing.T) {
			expectedErr := errors.New("rando create error here")
			repo := &fakeRepo{
				createFn: func(ctx context.Context, check health.Check) (health.Check, error) {
					return check, expectedErr
				},
			}
			svc := health.NewSVC(repo)

			_, err := svc.Create(context.TODO(), "http://example.com")
			if expectedErr != err {
				t.Fatalf("did not receive expected repo error: expected=%q got=%q", expectedErr, err)
			}
		})
	})
}

func validateID(t *testing.T, endpoint string, got string) {
	t.Helper()

	// validates that IDs' area created in this fashion
	// test driven here why there is going to be duplicate
	// code in service.

	h := md5.New()
	_, err := h.Write([]byte(endpoint))
	if err != nil {
		t.Fatal("unexpected hash error: ", err.Error())
	}

	expected := fmt.Sprintf("%x", h.Sum([]byte("tyrael")))
	if expected != got {
		t.Errorf("unexpected hash value: expected=%q got %q", expected, got)
	}
}

type fakeRepo struct {
	createFn func(ctx context.Context, check health.Check) (health.Check, error)
}

func (f *fakeRepo) Create(ctx context.Context, check health.Check) (health.Check, error) {
	if f.createFn == nil {
		panic("not implemented yet")
	}
	return f.createFn(ctx, check)
}
