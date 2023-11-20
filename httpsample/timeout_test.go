package httpsample_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSimple200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-length", "5")
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte("12345"))
	}))
	defer ts.Close()

	c := &http.Client{
		Timeout: 3000 * time.Microsecond,
	}
	resp, err := c.Get(ts.URL)
	if err != nil {
		t.Error("HTTP Get", err)
	}
	fmt.Printf("resp.Status: %#v\n", resp.Status)
	fmt.Printf("resp.Body: %#v\n", resp.Body)

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error("ReadAll", err)
	}
	fmt.Printf("Body: %s\n", bs)
}

func TestTimeoutBeforeResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Header().Set("content-length", "5")
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte("12345"))
	}))
	defer ts.Close()

	c := &http.Client{
		Timeout: 300 * time.Microsecond,
	}
	resp, err := c.Get(ts.URL)
	if err != nil {
		t.Error("HTTP Get", err)
	}
	fmt.Printf("resp.Status: %#v\n", resp.Status)
	fmt.Printf("resp.Body: %#v\n", resp.Body)

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error("ReadAll", err)
	}
	fmt.Printf("Body: %s\n", bs)
}
