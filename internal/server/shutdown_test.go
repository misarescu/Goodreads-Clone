package server

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServeGracefulShutdown(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	s := &Server{mux: http.NewServeMux()}
	s.mux.HandleFunc("/slow", func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusOK)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serveDone := make(chan error, 1)
	go func() { serveDone <- s.Serve(ctx, ln, 5*time.Second) }()

	respCh := make(chan *http.Response, 1)
	errCh := make(chan error, 1)
	go func() {
		resp, err := http.Get("http://" + addr + "/slow")
		if err != nil {
			errCh <- err
			return
		}
		respCh <- resp
	}()
	<-started

	cancel() // begin shutdown while the request is in flight

	waitRefused(t, addr)

	select {
	case <-serveDone:
		t.Fatal("Serve returned before in-flight request finished")
	default:
	}

	close(release)

	awaitResponse(t, respCh, errCh)

	select {
	case err := <-serveDone:
		if err != nil {
			t.Errorf("Serve returned %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return")
	}
}

// waitRefused fails unless new connections are refused once the listener is closed.
func waitRefused(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err != nil {
			return
		}
		_ = c.Close()
		if time.Now().After(deadline) {
			t.Fatal("new connections still accepted after shutdown started")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func awaitResponse(t *testing.T, respCh <-chan *http.Response, errCh <-chan error) {
	t.Helper()
	select {
	case resp := <-respCh:
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
	case err := <-errCh:
		t.Fatalf("in-flight request failed: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("in-flight request did not complete")
	}
}
