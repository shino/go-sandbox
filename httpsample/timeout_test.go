package httpsample_test

import (
	"fmt"
	"io"
	"net"
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
		t.Fatal("HTTP Get", err)
	}
	t.Logf("resp.Status: %#v\n", resp.Status)

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("ReadAll", err)
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
	_, err := c.Get(ts.URL)
	if err != nil {
		// Go 1.12.3
		// Output: c.Get err: &url.Error{Op:"Get", URL:"http://127.0.0.1:39731", Err:(*http.httpError)(0xc0001340f0)}
		t.Logf("c.Get err")
		t.Logf("err: %#v", err)
		switch v := err.(type) {
		case net.Error:
			t.Logf("err is net.Error")
			// Get \"http://127.0.0.1:36335\": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
			t.Logf("err.Error(): %#v", v.Error())
			// Output: true
			t.Logf("err.Timeout(): %#v", v.Timeout())
			// Output: true
			// !!Deprecated!!
			t.Logf("err.Temporary(): %#v", v.Temporary())
		default:
			t.Fatalf("err is not net.Error: %#v", v)
		}
		return
	}
	t.Fatal("Must not reach here")
}
