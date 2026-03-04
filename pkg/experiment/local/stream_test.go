package local

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/r3labs/sse/v2"
	"github.com/stretchr/testify/assert"
)

type mockEventSource struct {
	httpClient *http.Client
	url        string
	headers    map[string]string

	subscribeChanError error
	chConnected        chan bool

	ctx         context.Context
	messageChan chan *sse.Event
	onDisCb     sse.ConnCallback
	onConnCb    sse.ConnCallback
}

func (s *mockEventSource) OnDisconnect(fn sse.ConnCallback) {
	s.onDisCb = fn
}

func (s *mockEventSource) OnConnect(fn sse.ConnCallback) {
	s.onConnCb = fn
}

func (s *mockEventSource) SubscribeChanRawWithContext(ctx context.Context, ch chan *sse.Event) error {
	s.ctx = ctx
	s.messageChan = ch
	s.chConnected <- true
	return s.subscribeChanError
}

func (s *mockEventSource) mockEventSourceFactory(httpClient *http.Client, url string, headers map[string]string) eventSource {
	s.httpClient = httpClient
	s.url = url
	s.headers = headers
	return s
}

func TestStream(t *testing.T) {
	var s = mockEventSource{chConnected: make(chan bool)}
	client := newSseStream("authToken", "url", 2*time.Second, 4*time.Second, 6*time.Second, 1*time.Second)
	client.setNewESFactory(s.mockEventSourceFactory)
	messageCh := make(chan streamEvent)
	errorCh := make(chan error)

	// Make connection.
	client.Connect(messageCh, errorCh)
	// Wait for connection "establish".
	<-s.chConnected

	// Check for all variables.
	assert.Equal(t, "url", s.url)
	assert.Equal(t, "authToken", s.headers["Authorization"])
	assert.NotNil(t, s.headers["X-Amp-Exp-Library"])

	// Signal connected.
	s.onConnCb(nil)

	// Send update 1, ensure received.
	go func() { s.messageChan <- &sse.Event{Data: []byte("data1")} }()
	assert.Equal(t, []byte("data1"), (<-messageCh).data)

	// Send keep alive, not passed down, checked later along with updates 2 and 3.
	go func() { s.messageChan <- &sse.Event{Data: []byte(" ")} }()

	// Send update 2 and 3, ensure received in order.
	go func() {
		s.messageChan <- &sse.Event{Data: []byte("data2")}
		s.messageChan <- &sse.Event{Data: []byte("data3")}
	}()
	assert.Equal(t, []byte("data2"), (<-messageCh).data)
	assert.Equal(t, []byte("data3"), (<-messageCh).data)

	// Stop client, ensure context cancelled.
	client.Cancel()
	assert.True(t, errors.Is(s.ctx.Err(), context.Canceled))

	// No message is passed through after cancel even it's received.
	go func() { s.messageChan <- &sse.Event{Data: []byte("data4")} }()

	// Ensure no message after cancel.
	select {
	case msg, ok := <-messageCh:
		if ok {
			assert.Fail(t, "Unexpected data message received", string(msg.data))
		}
	case err, ok := <-errorCh:
		if ok {
			assert.Fail(t, "Unexpected error message received", err)
		}
	case <-time.After(1 * time.Second):
		// No message received within the timeout, as expected
	}
}

func TestStreamConnTimeout(t *testing.T) {
	var s = mockEventSource{chConnected: make(chan bool)}
	client := newSseStream("", "", 2*time.Second, 4*time.Second, 6*time.Second, 1*time.Second)
	client.setNewESFactory(s.mockEventSourceFactory)
	messageCh := make(chan streamEvent)
	errorCh := make(chan error)

	// Make connection.
	client.Connect(messageCh, errorCh)
	<-s.chConnected
	// Wait for timeout to reach.
	time.Sleep(2*time.Second + 10*time.Millisecond)
	// Check that context cancelled and error received.
	assert.True(t, errors.Is(s.ctx.Err(), context.Canceled))
	assert.Equal(t, errors.New("stream connection timeout"), <-errorCh)
}

func TestStreamKeepAliveTimeout(t *testing.T) {
	var s = mockEventSource{chConnected: make(chan bool)}
	client := newSseStream("", "", 2*time.Second, 1*time.Second, 6*time.Second, 1*time.Second)
	client.setNewESFactory(s.mockEventSourceFactory)
	messageCh := make(chan streamEvent)
	errorCh := make(chan error)

	// Make connection.
	client.Connect(messageCh, errorCh)
	<-s.chConnected
	s.onConnCb(nil)

	// Send keepalive 1 and wait.
	go func() { s.messageChan <- &sse.Event{Data: []byte(" ")} }()
	time.Sleep(1*time.Second - 10*time.Millisecond)
	assert.False(t, errors.Is(s.ctx.Err(), context.Canceled))
	// Send keepalive 2 and wait.
	go func() { s.messageChan <- &sse.Event{Data: []byte(" ")} }()
	time.Sleep(1*time.Second - 10*time.Millisecond)
	assert.False(t, errors.Is(s.ctx.Err(), context.Canceled))
	// Send data and wait, data should reset keepalive.
	go func() { s.messageChan <- &sse.Event{Data: []byte("data1")} }()
	assert.Equal(t, []byte("data1"), (<-messageCh).data)
	time.Sleep(1*time.Second - 10*time.Millisecond)
	assert.False(t, errors.Is(s.ctx.Err(), context.Canceled))
	// Send data ensure stream is open.
	go func() { s.messageChan <- &sse.Event{Data: []byte("data1")} }()
	assert.Equal(t, []byte("data1"), (<-messageCh).data)
	assert.False(t, errors.Is(s.ctx.Err(), context.Canceled))
	// Wait for keepalive to timeout, stream should close.
	time.Sleep(1*time.Second + 10*time.Millisecond)
	assert.Equal(t, errors.New("stream keepalive timed out"), <-errorCh)
	assert.True(t, errors.Is(s.ctx.Err(), context.Canceled))
}

func TestStreamReconnectsTimeout(t *testing.T) {
	var s = mockEventSource{chConnected: make(chan bool)}
	client := newSseStream("", "", 2*time.Second, 3*time.Second, 2*time.Second, 0*time.Second)
	client.setNewESFactory(s.mockEventSourceFactory)
	messageCh := make(chan streamEvent)
	errorCh := make(chan error)

	// Make connection.
	client.Connect(messageCh, errorCh)
	<-s.chConnected
	s.onConnCb(nil)

	go func() { s.messageChan <- &sse.Event{Data: []byte("data1")} }()
	assert.Equal(t, []byte("data1"), (<-messageCh).data)
	// Sleep for reconnect to timeout, data should pass through.
	time.Sleep(2*time.Second + 100*time.Millisecond)
	<-s.chConnected
	s.onConnCb(nil)
	go func() { s.messageChan <- &sse.Event{Data: []byte(" ")} }()
	go func() { s.messageChan <- &sse.Event{Data: []byte("data2")} }()
	assert.Equal(t, []byte("data2"), (<-messageCh).data)
	assert.False(t, errors.Is(s.ctx.Err(), context.Canceled))
	// Cancel stream, should cancel context.
	client.Cancel()
	assert.True(t, errors.Is(s.ctx.Err(), context.Canceled))
	select {
	case msg, ok := <-errorCh:
		if ok {
			assert.Fail(t, "Unexpected message received after disconnect", msg)
		}
	case <-time.After(3 * time.Second):
		// No message received within the timeout, as expected
	}
}

func TestStreamConnectAndCancelImmediately(t *testing.T) {
	var s = mockEventSource{chConnected: make(chan bool)}
	client := newSseStream("", "", 2*time.Second, 3*time.Second, 2*time.Second, 0*time.Second)
	client.setNewESFactory(s.mockEventSourceFactory)
	messageCh := make(chan streamEvent)
	errorCh := make(chan error)

	// Make connection and cancel immediately.
	client.Connect(messageCh, errorCh)
	client.Cancel()
	// Make sure no error for all timeouts.
	select {
	case msg, ok := <-errorCh:
		if ok {
			assert.Fail(t, "Unexpected message received after disconnect", msg)
		}
	case <-time.After(4 * time.Second):
		// No message received within the timeout, as expected
	}
}

func TestStreamChannelCloseOk(t *testing.T) {
	var s = mockEventSource{chConnected: make(chan bool)}
	client := newSseStream("", "", 1*time.Second, 1*time.Second, 1*time.Second, 0*time.Second)
	client.setNewESFactory(s.mockEventSourceFactory)
	messageCh := make(chan streamEvent)
	errorCh := make(chan error)

	// Close channels.
	close(messageCh)
	close(errorCh)

	// Connect and send message, the client should cancel right away.
	client.Connect(messageCh, errorCh)
	<-s.chConnected
	s.onConnCb(nil)

	// Test no message received for closed channel.
	s.messageChan <- &sse.Event{Data: []byte("data1")}
	assert.True(t, errors.Is(s.ctx.Err(), context.Canceled))

	select {
	case msg, ok := <-messageCh:
		if ok {
			assert.Fail(t, "Unexpected message received after close", msg)
		}
	case msg, ok := <-errorCh:
		if ok {
			assert.Fail(t, "Unexpected message received after close", msg)
		}
	case <-time.After(2 * time.Second):
		// No message received within the timeout, as expected
	}
}

func TestStreamDisconnectErrorPasses(t *testing.T) {
	var s = mockEventSource{chConnected: make(chan bool)}
	client := newSseStream("", "", 1*time.Second, 1*time.Second, 1*time.Second, 0*time.Second)
	client.setNewESFactory(s.mockEventSourceFactory)
	messageCh := make(chan streamEvent)
	errorCh := make(chan error)

	// Make connection.
	client.Connect(messageCh, errorCh)
	<-s.chConnected
	s.onConnCb(nil)

	// Disconnect error goes through.
	s.onDisCb(nil)
	assert.Equal(t, errors.New("stream disconnected error"), <-errorCh)

	select {
	case msg, ok := <-errorCh:
		if ok {
			assert.Fail(t, "Unexpected message received after disconnect", msg)
		}
	case <-time.After(2 * time.Second):
		// No message received within the timeout, as expected
	}
}

func TestStreamConnectErrorPasses(t *testing.T) {
	var s = mockEventSource{chConnected: make(chan bool)}
	client := newSseStream("", "", 1*time.Second, 1*time.Second, 1*time.Second, 0*time.Second)
	client.setNewESFactory(s.mockEventSourceFactory)
	messageCh := make(chan streamEvent)
	errorCh := make(chan error)

	// Make connection.
	s.subscribeChanError = errors.New("some error occurred")
	client.Connect(messageCh, errorCh)
	<-s.chConnected
	s.onConnCb(nil)

	// Connect error goes through.
	assert.Equal(t, errors.New("some error occurred"), <-errorCh)

	select {
	case msg, ok := <-errorCh:
		if ok {
			assert.Fail(t, "Unexpected message received after disconnect", msg)
		}
	case <-time.After(2 * time.Second):
		// No message received within the timeout, as expected
	}
}

// mockEventSourceBlockingSubscribe is a variant of mockEventSource where
// SubscribeChanRawWithContext blocks until the context is cancelled, then
// returns an error. This reproduces the scenario where the subscribe goroutine
// tries to send to esConnectErrCh after the main event loop has already exited
// due to a connection timeout (Leak 2).
type mockEventSourceBlockingSubscribe struct {
	ctx         context.Context
	messageChan chan *sse.Event
	chConnected chan bool
	onDisCb     sse.ConnCallback
	onConnCb    sse.ConnCallback
}

func (s *mockEventSourceBlockingSubscribe) OnDisconnect(fn sse.ConnCallback) { s.onDisCb = fn }
func (s *mockEventSourceBlockingSubscribe) OnConnect(fn sse.ConnCallback)    { s.onConnCb = fn }
func (s *mockEventSourceBlockingSubscribe) SubscribeChanRawWithContext(ctx context.Context, ch chan *sse.Event) error {
	s.ctx = ctx
	s.messageChan = ch
	s.chConnected <- true
	<-ctx.Done() // Block until cancelled — simulates a long-lived SSE connection.
	return errors.New("context cancelled")
}

// goroutinesStabilize polls runtime.NumGoroutine() until it falls to <= limit
// or the timeout expires, returning true when the limit is reached.
func goroutinesStabilize(limit int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= limit {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return runtime.NumGoroutine() <= limit
}

// TestStreamNoGoroutineLeakOnConnectCancelRace verifies that concurrently
// firing the OnConnect callback and cancelling the context does not leak
// goroutines (Leak 1).
//
// Before the fix, OnConnect spawned `go func() { connectCh <- true }()` after
// a ctx.Done() TOCTOU check. If the main event loop exited on ctx.Done() before
// that goroutine could send, it would block forever on the unbuffered channel.
func TestStreamNoGoroutineLeakOnConnectCancelRace(t *testing.T) {
	baseline := runtime.NumGoroutine()
	const iterations = 50
	for i := 0; i < iterations; i++ {
		s := mockEventSource{chConnected: make(chan bool)}
		client := newSseStream("", "", 2*time.Second, 3*time.Second, 10*time.Second, 0*time.Second)
		client.setNewESFactory(s.mockEventSourceFactory)
		messageCh := make(chan streamEvent)
		errorCh := make(chan error, 1)
		client.Connect(messageCh, errorCh)
		<-s.chConnected

		// Fire OnConnect and Cancel concurrently to maximise the chance of
		// hitting the TOCTOU window.
		done := make(chan struct{})
		go func() {
			defer close(done)
			s.onConnCb(nil)
		}()
		client.Cancel()
		<-done
		select {
		case <-errorCh:
		default:
		}
	}
	assert.True(t, goroutinesStabilize(baseline+2, 200*time.Millisecond),
		"goroutine count %d exceeds baseline %d: goroutine leak detected", runtime.NumGoroutine(), baseline)
}

// TestStreamNoGoroutineLeakSubscribeErrorAfterCancel verifies that a subscribe
// error returned after the context is already cancelled does not leave the
// subscribe goroutine permanently blocked (Leak 2).
//
// Before the fix, the subscribe goroutine called `esConnectErrCh <- err` on an
// unbuffered channel after the main event loop had already exited (due to a
// connection timeout), causing the goroutine to block forever.
func TestStreamNoGoroutineLeakSubscribeErrorAfterCancel(t *testing.T) {
	baseline := runtime.NumGoroutine()
	const iterations = 10
	for i := 0; i < iterations; i++ {
		s := &mockEventSourceBlockingSubscribe{chConnected: make(chan bool)}
		// Short connection timeout so the main event loop exits quickly, leaving
		// the blocking subscribe goroutine to return an error afterwards.
		client := newSseStream("", "", 100*time.Millisecond, 3*time.Second, 10*time.Second, 0*time.Second)
		client.setNewESFactory(func(_ *http.Client, _ string, _ map[string]string) eventSource {
			return s
		})
		messageCh := make(chan streamEvent)
		errorCh := make(chan error, 1)
		client.Connect(messageCh, errorCh)
		<-s.chConnected

		// The 100ms connection timeout fires, cancels the context, and sends a
		// timeout error. The blocking subscribe goroutine then unblocks via
		// ctx.Done() and returns an error — the old code would leak here.
		select {
		case <-errorCh:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("expected timeout error not received")
		}
		client.Cancel()
	}
	assert.True(t, goroutinesStabilize(baseline+2, 200*time.Millisecond),
		"goroutine count %d exceeds baseline %d: goroutine leak detected", runtime.NumGoroutine(), baseline)
}

// TestStreamNoGoroutineLeakOnDisconnectCancelRace verifies that concurrently
// firing the OnDisconnect callback and cancelling the context does not leave
// the SSE library's callback goroutine permanently blocked (Leak 3).
//
// Before the fix, OnDisconnect did a blocking send to an unbuffered channel
// inside a `default` branch, creating the same TOCTOU race as OnConnect: the
// context could be cancelled between the ctx.Done() check and the send.
func TestStreamNoGoroutineLeakOnDisconnectCancelRace(t *testing.T) {
	baseline := runtime.NumGoroutine()
	const iterations = 50
	for i := 0; i < iterations; i++ {
		s := mockEventSource{chConnected: make(chan bool)}
		client := newSseStream("", "", 2*time.Second, 3*time.Second, 10*time.Second, 0*time.Second)
		client.setNewESFactory(s.mockEventSourceFactory)
		messageCh := make(chan streamEvent)
		errorCh := make(chan error, 1)
		client.Connect(messageCh, errorCh)
		<-s.chConnected
		s.onConnCb(nil)

		// Fire OnDisconnect and Cancel concurrently to hit the TOCTOU window.
		done := make(chan struct{})
		go func() {
			defer close(done)
			s.onDisCb(nil)
		}()
		client.Cancel()
		<-done
		select {
		case <-errorCh:
		default:
		}
	}
	assert.True(t, goroutinesStabilize(baseline+2, 200*time.Millisecond),
		"goroutine count %d exceeds baseline %d: goroutine leak detected", runtime.NumGoroutine(), baseline)
}
