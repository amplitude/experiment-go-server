package local

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStream(t *testing.T) {
	var streamOnUpdate func(es *EventSource, msg []byte)
	var streamCtx context.Context
	streamConnectCalled := make(chan bool)
	es := EventSource{}
	newMockES := func (ctx context.Context, url string, connTimeout time.Duration, maxTime time.Duration, headers map[string]string, onConnect func(es *EventSource), onDisconnect func(es *EventSource), onUpdate func(es *EventSource, msg []byte)) *EventSource {
		// Check for all variables.
		assert.Equal(t, "url", url)
		assert.Equal(t, 2 * time.Second, connTimeout)
		assert.Equal(t, 7 * time.Second, maxTime)
		assert.Equal(t, "authToken", headers["Authorization"])
		assert.NotNil(t, headers["X-Amp-Exp-Library"])

		es.connect = func() error {
			streamCtx = ctx
			streamOnUpdate = onUpdate
			onConnect(&es)
			streamConnectCalled <- true

			<-ctx.Done()
			return nil
		}
		return &es
	}
	client := NewSseStream("authToken", "url", 2 * time.Second, 4 * time.Second, 6 * time.Second, 1 * time.Second, newMockES)
	messageCh := make(chan StreamEvent)
	errorCh := make(chan error)
	
	// Make connection.
	client.Connect(messageCh, errorCh)
	// Wait for connection "establish".
	<-streamConnectCalled
	
	// Send update 1, ensure received.
	go func() {streamOnUpdate(&es, []byte("data1"))}()
	assert.Equal(t, []byte("data1"), (<-messageCh).data)

	// Send update 2 and 3, ensure received in order.
	go func() {
		streamOnUpdate(&es, []byte("data2"))
		streamOnUpdate(&es, []byte("data3"))
	}()
	assert.Equal(t, []byte("data2"), (<-messageCh).data)
	assert.Equal(t, []byte("data3"), (<-messageCh).data)

	// Stop client, ensure context cancelled.
	client.Cancel()
	assert.True(t, errors.Is(streamCtx.Err(), context.Canceled))

	// No message is passed through even it's received.
	go func() {streamOnUpdate(&es, []byte("data4"))}()

	// Ensure no message after cancel.
	select {
	case msg, ok := <-messageCh:
		if ok {
			assert.Fail(t, "Unexpected data message received", msg)
		}
	case msg, ok := <-errorCh:
		if ok {
			assert.Fail(t, "Unexpected error message received", msg)
		}
	case <-time.After(1 * time.Second):
		// No message received within the timeout, as expected
	}
}

func TestStreamConnTimeout(t *testing.T) {
	var streamCtx context.Context
	es := EventSource{}
	newMockES := func (ctx context.Context, url string, connTimeout time.Duration, maxTime time.Duration, headers map[string]string, onConnect func(es *EventSource), onDisconnect func(es *EventSource), onUpdate func(es *EventSource, msg []byte)) *EventSource {
		es.connect = func() error {
			streamCtx = ctx
			<-ctx.Done()
			return nil
		}
		return &es
	}
	client := NewSseStream("", "", 2 * time.Second, 4 * time.Second, 6 * time.Second, 1 * time.Second, newMockES)
	messageCh := make(chan StreamEvent)
	errorCh := make(chan error)
	
	// Make connection. 
	client.Connect(messageCh, errorCh)
	// Wait for timeout to reach. 
	time.Sleep(2 * time.Second + 10 * time.Millisecond)
	// Check that context cancelled and error received.
 	assert.True(t, errors.Is(streamCtx.Err(), context.Canceled))
	assert.Equal(t, errors.New("timedout error"), <-errorCh)
}

func TestStreamKeepAliveTimeout(t *testing.T) {
	var streamOnUpdate func(es *EventSource, msg []byte)
	var streamCtx context.Context
	streamConnectCalled := make(chan bool)
	es := EventSource{}
	newMockES := func (ctx context.Context, url string, connTimeout time.Duration, maxTime time.Duration, headers map[string]string, onConnect func(es *EventSource), onDisconnect func(es *EventSource), onUpdate func(es *EventSource, msg []byte)) *EventSource {
		es.connect = func() error {
			streamCtx = ctx
			streamOnUpdate = onUpdate
			onConnect(&es)
			streamConnectCalled <- true
			<-ctx.Done()
			return nil
		}
		return &es
	}
	client := NewSseStream("", "", 2 * time.Second, 1 * time.Second, 6 * time.Second, 1 * time.Second, newMockES)
	messageCh := make(chan StreamEvent)
	errorCh := make(chan error)
	
	// Make connection.
	client.Connect(messageCh, errorCh)
	<-streamConnectCalled

	// Send keepalive 1 and wait.
	go func() {streamOnUpdate(&es, []byte(" "))}()
	time.Sleep(1 * time.Second - 10 * time.Millisecond)
	assert.False(t, errors.Is(streamCtx.Err(), context.Canceled))
	// Send keepalive 2 and wait.
	go func() {streamOnUpdate(&es, []byte(" "))}()
	time.Sleep(1 * time.Second - 10 * time.Millisecond)
	assert.False(t, errors.Is(streamCtx.Err(), context.Canceled))
	// Send data and wait, data should reset keepalive.
	go func() {streamOnUpdate(&es, []byte("data1"))}()
	assert.Equal(t, []byte("data1"), (<-messageCh).data)
	time.Sleep(1 * time.Second - 10 * time.Millisecond)
	assert.False(t, errors.Is(streamCtx.Err(), context.Canceled))
	// Send data ensure stream is open.
	go func() {streamOnUpdate(&es, []byte("data1"))}()
	assert.Equal(t, []byte("data1"), (<-messageCh).data)
	assert.False(t, errors.Is(streamCtx.Err(), context.Canceled))
	// Wait for keepalive to timeout, stream should close.
	time.Sleep(1 * time.Second + 10 * time.Millisecond)
	assert.Equal(t, errors.New("keep alive failed"), <-errorCh)
 	assert.True(t, errors.Is(streamCtx.Err(), context.Canceled))
}

func TestStreamReconnectsTimeout(t *testing.T) {
	var streamOnUpdate func(es *EventSource, msg []byte)
	var streamCtx context.Context
	streamConnectCalled := make(chan bool)
	es := EventSource{}
	newMockES := func (ctx context.Context, url string, connTimeout time.Duration, maxTime time.Duration, headers map[string]string, onConnect func(es *EventSource), onDisconnect func(es *EventSource), onUpdate func(es *EventSource, msg []byte)) *EventSource {
		es.connect = func() error {
			streamCtx = ctx
			streamOnUpdate = onUpdate
			onConnect(&es)
			streamConnectCalled <- true
			<-ctx.Done()
			return nil
		}
		return &es
	}
	client := NewSseStream("", "", 2 * time.Second, 3 * time.Second, 2 * time.Second, 0 * time.Second, newMockES)
	messageCh := make(chan StreamEvent)
	errorCh := make(chan error)
	
	// Make connection.
	client.Connect(messageCh, errorCh)
	<-streamConnectCalled
	// Sleep for reconnect to timeout, data should pass through.
	time.Sleep(2 * time.Second * 3 + 100 * time.Millisecond)
	go func() {streamOnUpdate(&es, []byte(" "))}()
	go func() {streamOnUpdate(&es, []byte("data1"))}()
	assert.Equal(t, []byte("data1"), (<-messageCh).data)
	assert.False(t, errors.Is(streamCtx.Err(), context.Canceled))
	// Cancel stream, should cancel context.
	client.Cancel()
	assert.True(t, errors.Is(streamCtx.Err(), context.Canceled))
	select {
	case msg, ok := <-errorCh:
		if ok {
			assert.Fail(t, "Unexpected message received after disconnect", msg)
		}
	case <-time.After(6 * time.Second):
		// No message received within the timeout, as expected
	}
}

func TestStreamConnectAndCancelImmediately(t *testing.T) {
	es := EventSource{}
	newMockES := func (ctx context.Context, url string, connTimeout time.Duration, maxTime time.Duration, headers map[string]string, onConnect func(es *EventSource), onDisconnect func(es *EventSource), onUpdate func(es *EventSource, msg []byte)) *EventSource {
		es.connect = func() error {
			onConnect(&es)
			<-ctx.Done()
			return nil
		}
		return &es
	}
	client := NewSseStream("", "", 2 * time.Second, 3 * time.Second, 2 * time.Second, 0 * time.Second, newMockES)
	messageCh := make(chan StreamEvent)
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
	var streamOnUpdate func(es *EventSource, msg []byte)
	streamConnectCalled := make(chan bool)
	var streamCtx context.Context
	es := EventSource{}
	newMockES := func (ctx context.Context, url string, connTimeout time.Duration, maxTime time.Duration, headers map[string]string, onConnect func(es *EventSource), onDisconnect func(es *EventSource), onUpdate func(es *EventSource, msg []byte)) *EventSource {
		es.connect = func() error {
			streamOnUpdate = onUpdate
			streamCtx = ctx
			onConnect(&es)
			streamConnectCalled <- true
			<-ctx.Done();
			return nil
		}
		return &es
	}
	client := NewSseStream("", "", 1 * time.Second, 1 * time.Second, 1 * time.Second, 0 * time.Second, newMockES)
	messageCh := make(chan StreamEvent)
	errorCh := make(chan error)

	// Close channels.
	close(messageCh)
	close(errorCh)

	// Connect and send message, the client should cancel right away.
	client.Connect(messageCh, errorCh)
	<-streamConnectCalled
	streamOnUpdate(&es, []byte("data1"))
	assert.True(t, errors.Is(streamCtx.Err(), context.Canceled))

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
	var streamOnDisconnect func(es *EventSource)
	streamConnectCalled := make(chan bool)
	es := EventSource{}
	newMockES := func (ctx context.Context, url string, connTimeout time.Duration, maxTime time.Duration, headers map[string]string, onConnect func(es *EventSource), onDisconnect func(es *EventSource), onUpdate func(es *EventSource, msg []byte)) *EventSource {
		es.connect = func() error {
			streamOnDisconnect = onDisconnect
			onConnect(&es)
			streamConnectCalled <- true
			<-ctx.Done();
			return nil
		}
		return &es
	}
	client := NewSseStream("", "", 1 * time.Second, 1 * time.Second, 1 * time.Second, 0 * time.Second, newMockES)
	messageCh := make(chan StreamEvent)
	errorCh := make(chan error)
	
	// Make connection.
	client.Connect(messageCh, errorCh)
	<-streamConnectCalled

	// Disconnect error goes through.
	go func() {streamOnDisconnect(&es)}()
	assert.Equal(t, errors.New("disconnected error"), <-errorCh)

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
	streamConnectCalled := make(chan bool)
	es := EventSource{}
	newMockES := func (ctx context.Context, url string, connTimeout time.Duration, maxTime time.Duration, headers map[string]string, onConnect func(es *EventSource), onDisconnect func(es *EventSource), onUpdate func(es *EventSource, msg []byte)) *EventSource {
		es.connect = func() error {
			streamConnectCalled <- true
			return errors.New("some error occurred")
		}
		return &es
	}
	client := NewSseStream("", "", 1 * time.Second, 1 * time.Second, 1 * time.Second, 0 * time.Second, newMockES)
	messageCh := make(chan StreamEvent)
	errorCh := make(chan error)
	
	// Make connection.
	client.Connect(messageCh, errorCh)
	<-streamConnectCalled

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

