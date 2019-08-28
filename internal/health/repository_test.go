package health_test

import (
	"encoding/gob"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jsteenb2/health/internal/health"
)

func TestFileRepository(t *testing.T) {
	newTempDir := func(t *testing.T) string {
		t.Helper()

		tmpDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal(err)
		}
		return tmpDir
	}

	readChecksFromFile := func(t *testing.T, filepath string) []health.Check {
		t.Helper()

		f, err := os.Open(filepath)
		mustNoError(t, err)
		defer f.Close()

		var checks []health.Check
		mustNoError(t, gob.NewDecoder(f).Decode(&checks))
		return checks
	}

	newFileWithChecks := func(t *testing.T, filepath string, checks ...health.Check) {
		t.Helper()

		f, err := os.Create(filepath)
		mustNoError(t, err)
		defer f.Close()

		err = gob.NewEncoder(f).Encode(checks)
		mustNoError(t, err)
		mustNoError(t, f.Close())
	}

	t.Run("NewFileRepository", func(t *testing.T) {
		t.Run("creates the persistence file if one doesn't exist", func(t *testing.T) {
			tmpDir := newTempDir(t)
			defer os.RemoveAll(tmpDir)

			expectedFile := filepath.Join(tmpDir, "file_repo")
			_, err := health.NewFileRepository(expectedFile)
			mustNoError(t, err)

			stats, err := os.Stat(expectedFile)
			mustNoError(t, err)

			equal(t, "file_repo", stats.Name(), "file name does not match")
		})

		t.Run("if file is persisted should load it into memory", func(t *testing.T) {
			tmpDir := newTempDir(t)
			defer os.RemoveAll(tmpDir)

			filePath := filepath.Join(tmpDir, "tmp_file")

			existingCheck := health.Check{ID: "id", Endpoint: "endpoint"}
			newFileWithChecks(t, filePath, existingCheck)

			repo, err := health.NewFileRepository(filePath)
			mustNoError(t, err)

			readChecks := repo.List(0, -1)
			mustEqual(t, 1, len(readChecks), "unexpected number of checks")
			mustEqual(t, existingCheck, readChecks[0], "invalid check received")
		})
	})

	t.Run("create", func(t *testing.T) {
		t.Run("adds new check to the checks", func(t *testing.T) {
			tmpDir := newTempDir(t)
			defer os.RemoveAll(tmpDir)

			file := filepath.Join(tmpDir, "file_repo")
			repo, err := health.NewFileRepository(file)
			mustNoError(t, err)

			newCheck := health.Check{
				ID:       "id-1",
				Endpoint: "http://example.com",
			}
			err = repo.Create(newCheck)
			mustNoError(t, err)

			checks := readChecksFromFile(t, file)
			mustEqual(t, 1, len(checks), "wrong number of checks found")
			equal(t, newCheck, checks[0], "check bounced")
		})

		t.Run("fails to write when check already exists", func(t *testing.T) {
			tmpDir := newTempDir(t)
			defer os.RemoveAll(tmpDir)

			filePath := filepath.Join(tmpDir, "tmp_file")

			existingCheck := health.Check{ID: "id", Endpoint: "endpoint"}
			newFileWithChecks(t, filePath, existingCheck)

			repo, err := health.NewFileRepository(filePath)
			mustNoError(t, err)

			err = repo.Create(existingCheck)
			mustError(t, err)

			err = repo.Create(health.Check{ID: "new-id", Endpoint: "new endpoint"})
			mustNoError(t, err)
		})
	})
}

func mustNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal("unexpected error: ", err.Error())
	}
}

func mustError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected an error: got=<nil>")
	}
}
