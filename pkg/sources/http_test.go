package sources_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sources2 "github.com/MontFerret/lab/v2/pkg/sources"
)

func TestHTTPSourceReadRejectsNonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer srv.Close()

	src, err := sources2.Create(srv.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	onNext, onError := src.Read(context.Background())
	nextClosed := false
	errorClosed := false
	var files []sources2.File
	var errs []sources2.Error

	for !nextClosed || !errorClosed {
		select {
		case file, ok := <-onNext:
			if !ok {
				nextClosed = true
				continue
			}

			files = append(files, file)
		case err, ok := <-onError:
			if !ok {
				errorClosed = true
				continue
			}

			errs = append(errs, err)
		}
	}

	if len(files) != 0 {
		t.Fatalf("expected no files, got %+v", files)
	}

	if len(errs) != 1 {
		t.Fatalf("expected one error, got %+v", errs)
	}

	if errs[0].Filename != srv.URL {
		t.Fatalf("expected filename %q, got %q", srv.URL, errs[0].Filename)
	}

	if errs[0].Message != "404 Not Found" {
		t.Fatalf("expected status error, got %q", errs[0].Message)
	}
}
