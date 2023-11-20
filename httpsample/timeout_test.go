package httpsample_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSimple200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-length", "5")
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte("12345"))
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Error("HTTP Get", err)
	}
	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error("ReadAll", err)
	}
	fmt.Printf("Body: %s\n", bs)
}
