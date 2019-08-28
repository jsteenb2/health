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

	var existing []Check
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

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(newChecks); err != nil {
		return err
	}

	err := ioutil.WriteFile(r.filepath, buf.Bytes(), 666)
	if err != nil {
		return err
	}

	r.checks = newChecks
	return nil
}

func (r *fileRepository) List(page, size int) []Check {
	r.mu.Lock()
	defer r.mu.Unlock()
	if size == -1 {
		return r.checks
	}
	return r.checks[size*page : size*(page+1)]
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
