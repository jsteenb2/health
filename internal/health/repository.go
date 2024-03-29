package health

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

type fileRepository struct {
	filepath string

	mu     *sync.Mutex
	checks checks
}

var _ Repository = (*fileRepository)(nil)

func NewFileRepository(filepath string) (Repository, error) {
	existingChecks, err := checksFromPersistence(filepath)
	if err != nil {
		return nil, err
	}

	return &fileRepository{
		filepath: filepath,
		mu:       new(sync.Mutex),
		checks:   existingChecks,
	}, nil
}

func checksFromPersistence(filepath string) ([]Check, error) {
	f, err := os.Open(filepath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		f, err = os.Create(filepath)
		if err != nil {
			return nil, err
		}
	}
	defer f.Close()

	existing := make([]Check, 0)
	err = gob.NewDecoder(f).Decode(&existing)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return existing, nil
}

var errEndpointExists = errors.New("endpoint exists")

func (r *fileRepository) Create(check Check) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, found := r.checks.find(check.ID); found {
		return errEndpointExists
	}

	newChecks := append(r.checks, check)

	if err := r.toDisk(newChecks); err != nil {
		return err
	}

	r.checks = newChecks
	return nil
}

func (r *fileRepository) List(page, size int) (int, []Check) {
	r.mu.Lock()
	defer r.mu.Unlock()

	total := len(r.checks)
	if size == -1 {
		return total, r.checks
	}

	start := size * (page - 1)
	if start >= len(r.checks) {
		return total, []Check{}
	}

	end := size * (page)
	if end > len(r.checks) {
		end = len(r.checks)
	}

	return total, r.checks[start:end]
}

var errCheckNotFound = errors.New("check not found by the provided id")

func (r *fileRepository) Read(id string) (Check, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	check, found := r.checks.find(id)
	if !found {
		return Check{}, errCheckNotFound
	}
	return check, nil
}

func (r *fileRepository) Delete(id string) error {
	out := make([]Check, 0, len(r.checks)-1)
	for _, check := range r.checks {
		if id == check.ID {
			continue
		}
		out = append(out, check)
	}

	if err := r.toDisk(out); err != nil {
		return err
	}

	r.checks = out
	return nil
}

func (r *fileRepository) toDisk(c []Check) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(r.filepath, buf.Bytes(), 666)
}

type checks []Check

func (c checks) find(id string) (Check, bool) {
	for _, check := range c {
		if check.ID == id {
			return check, true
		}
	}
	return Check{}, false
}
