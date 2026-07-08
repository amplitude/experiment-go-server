package local

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// adapterHarness drives the real tmaxmaxEventSource (not the mock) against an
// httptest server and exposes its callbacks and message channel for assertions.
type adapterHarness struct {
	events       chan *sseEvent
	connected    chan struct{}
	disconnected chan struct{}
	subscribeErr chan error
}

func startAdapter(ctx context.Context, url string, headers map[string]string) *adapterHarness {
	h := &adapterHarness{
		events:       make(chan *sseEvent, 8),
		connected:    make(chan struct{}, 1),
		disconnected: make(chan struct{}, 1),
		subscribeErr: make(chan error, 1),
	}
	es := newEventSource(&http.Client{}, url, headers)
	es.OnConnect(func() { h.connected <- struct{}{} })
	es.OnDisconnect(func() { h.disconnected <- struct{}{} })
	go func() { h.subscribeErr <- es.SubscribeChanRawWithContext(ctx, h.events) }()
	return h
}

func (h *adapterHarness) awaitConnect(t *testing.T) {
	t.Helper()
	select {
	case <-h.connected:
	case <-time.After(2 * time.Second):
		t.Fatal("expected OnConnect to fire")
	}
}

func (h *adapterHarness) awaitEvent(t *testing.T) *sseEvent {
	t.Helper()
	select {
	case ev := <-h.events:
		return ev
	case <-time.After(2 * time.Second):
		t.Fatal("expected an event")
		return nil
	}
}

// sseHandler flushes the given raw SSE frames, then returns (closing the stream).
func sseHandler(frames ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		for _, f := range frames {
			fmt.Fprint(w, f)
			flusher.Flush()
		}
	}
}

// A validated 200 signals connect, keepalive frames survive as a single space
// byte, data events pass through, and a lost connection fires OnDisconnect.
func TestAdapterConnectDeliversEventsAndDisconnects(t *testing.T) {
	srv := httptest.NewServer(sseHandler("data:  \n\n", "data: hello\n\n"))
	defer srv.Close()

	h := startAdapter(context.Background(), srv.URL, map[string]string{"Authorization": "Api-Key test"})
	h.awaitConnect(t)

	keepalive := h.awaitEvent(t)
	assert.Equal(t, []byte{STREAM_KEEP_ALIVE_BYTE}, keepalive.Data, "keepalive should arrive as a single space byte")

	data := h.awaitEvent(t)
	assert.Equal(t, []byte("hello"), data.Data)

	select {
	case <-h.disconnected:
	case <-time.After(2 * time.Second):
		t.Fatal("expected OnDisconnect after the server closed the stream")
	}
	assert.Error(t, <-h.subscribeErr, "a lost connection should surface an error")
}

func TestAdapterSendsAuthHeader(t *testing.T) {
	gotAuth := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth <- r.Header.Get("Authorization")
		sseHandler("data: x\n\n")(w, r)
	}))
	defer srv.Close()

	h := startAdapter(context.Background(), srv.URL, map[string]string{"Authorization": "Api-Key secret"})
	h.awaitConnect(t)
	assert.Equal(t, "Api-Key secret", <-gotAuth)
}

func TestAdapterNon200IsPermanentError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	h := startAdapter(context.Background(), srv.URL, nil)
	require.Error(t, <-h.subscribeErr)
	assertNoSignal(t, h.connected, "OnConnect must not fire on a non-200 response")
	assertNoSignal(t, h.disconnected, "OnDisconnect must not fire when a connection was never established")
}

// go-sse validates Content-Type: text/event-stream; r3labs did not. This guards
// that behavior change so a proxy that mangles the content type fails fast
// rather than hanging on a silently-dead stream.
func TestAdapterRejectsWrongContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: hi\n\n")
	}))
	defer srv.Close()

	h := startAdapter(context.Background(), srv.URL, nil)
	require.Error(t, <-h.subscribeErr)
	assertNoSignal(t, h.connected, "a non-SSE content type must not be treated as connected")
}

func TestAdapterContextCancelReturnsCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "data:  \n\n")
		flusher.Flush()
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	h := startAdapter(ctx, srv.URL, nil)
	h.awaitConnect(t)
	h.awaitEvent(t)

	cancel()
	assert.ErrorIs(t, <-h.subscribeErr, context.Canceled)
	assertNoSignal(t, h.disconnected, "cancellation is not a disconnect")
}

func assertNoSignal(t *testing.T, ch <-chan struct{}, msg string) {
	t.Helper()
	select {
	case <-ch:
		t.Fatal(msg)
	default:
	}
}
