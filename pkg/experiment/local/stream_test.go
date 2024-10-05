package local

import (
	"context"
	"errors"
	"net/http"
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
