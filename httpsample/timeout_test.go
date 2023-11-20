package httpsample_test

import (
	"context"
	"errors"
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
		// Outputs are at Go 1.12.3
		t.Logf("c.Get err")
		// Output: c.Get err: &url.Error{Op:"Get", URL:"http://127.0.0.1:39731", Err:(*http.httpError)(0xc0001340f0)}
		t.Logf("err: %#v", err)
		// Output: Get "http://127.0.0.1:41883": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
		t.Logf("err: %+v", err)
		// Output: Get "http://127.0.0.1:32919": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
		t.Logf("err: %s", err)
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

func TestTimeoutAtBodyRead(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Header().Set("content-length", "5")
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("123"))
		go func() {
			time.Sleep(2000 * time.Millisecond)
			w.Write([]byte("45"))
		}()
	}))
	defer ts.Close()

	c := &http.Client{
		Timeout: 1000 * time.Millisecond,
	}
	t.Log("About to Get")
	resp, err := c.Get(ts.URL)
	if err != nil {
		t.Logf("err: %v", err)
		t.Logf("err: %#v", err)
		t.Fatal("HTTP Get", err)
	}
	t.Logf("resp.Status: %#v\n", resp.Status)

	defer resp.Body.Close()

	if _, err := io.ReadAll(resp.Body); err != nil {
		// Outputs are at Go 1.12.3
		t.Logf("c.Get err")
		// Output: &errors.errorString{s:"unexpected EOF"}
		t.Logf("err: %#v", err)
		// Output: unexpected EOF
		t.Logf("err: %v", err)
		switch v := err.(type) {
		case net.Error:
			t.Logf("err is net.Error")
			t.Logf("err.Error(): %#v", v.Error())
			t.Logf("err.Timeout(): %#v", v.Timeout())
			t.Logf("err.Temporary(): %#v", v.Temporary())
		default:
			// This clause is executed, err is not net.Error instance
		}
		return
	}

	// var b bytes.Buffer
	// count, err := io.CopyN(&b, resp.Body, 5)
	// fmt.Printf("count: %#v\n", count) // count = 3
	// fmt.Printf("err: %#v\n", err)
	// count, err = io.CopyN(&b, resp.Body, 2)
	// fmt.Printf("count: %#v\n", count)
	// fmt.Printf("err: %#v\n", err)     // unexpected EOF

	t.Fatal("Must not reach here")
}

func TestTimeoutAtBodyReadWithContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Header().Set("content-length", "5")
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("123"))
		go func() {
			time.Sleep(2000 * time.Millisecond)
			w.Write([]byte("45"))
		}()
	}))
	defer ts.Close()

	c := &http.Client{
		Timeout: 2000 * time.Millisecond,
	}
	ctx, _ := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Logf("err: %v", err)
		t.Logf("err: %#v", err)
		t.Fatal("HTTP NewRequestWithContext", err)
	}
	t.Log("About to Get")
	resp, err := c.Do(req)
	if err != nil {
		t.Logf("err: %v", err)
		t.Logf("err: %#v", err)
		t.Fatal("HTTP Get", err)
	}
	t.Logf("resp.Status: %#v\n", resp.Status)

	defer resp.Body.Close()

	if _, err := io.ReadAll(resp.Body); err != nil {
		// Outputs are at Go 1.12.3
		t.Logf("c.Get err")
		// Output: unexpected EOF
		t.Logf("err: %v", err)
		// Check the context
		select {
		case <-ctx.Done():
			t.Log("context done")
			// Output: context deadline exceeded
			t.Logf("context err: %+v", ctx.Err())
			// Output: true
			t.Logf("ctx.Err() is context.DeadlineExceeded: %v", errors.Is(ctx.Err(), context.DeadlineExceeded))
			switch v := ctx.Err().(type) {
			case net.Error:
				// Output: true
				t.Logf("timeout:   %v", v.Timeout())
				// Output: true     *Deprecated*
				t.Logf("temporary: %v", v.Temporary())
			}
		}
		return
	}
	t.Fatal("Must not reach here")
}
