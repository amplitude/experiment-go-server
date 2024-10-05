package local

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/amplitude/experiment-go-server/internal/evaluation"
	"github.com/stretchr/testify/assert"
)

type mockSseStream struct {
	// Params
	authToken         string
	url               string
	connectionTimeout time.Duration
	keepaliveTimeout  time.Duration
	reconnInterval    time.Duration
	maxJitter         time.Duration

	// Channels to emit messages to simulate new events received through stream.
	messageCh chan (streamEvent)
	errorCh   chan (error)

	// Channel to tell there's a connection call.
	chConnected chan bool
}

func (s *mockSseStream) Connect(messageCh chan (streamEvent), errorCh chan (error)) error {
	s.messageCh = messageCh
	s.errorCh = errorCh

	s.chConnected <- true
	return nil
}

func (s *mockSseStream) Cancel() {
}

func (s *mockSseStream) setNewESFactory(f func(httpClient *http.Client, url string, headers map[string]string) eventSource) {
}

func (s *mockSseStream) newSseStreamFactory(
	authToken,
	url string,
	connectionTimeout time.Duration,
	keepaliveTimeout time.Duration,
	reconnInterval time.Duration,
	maxJitter time.Duration,
) stream {
	s.authToken = authToken
	s.url = url
	s.connectionTimeout = connectionTimeout
	s.keepaliveTimeout = keepaliveTimeout
	s.reconnInterval = reconnInterval
	s.maxJitter = maxJitter
	return s
}

var FLAG_1_STR = []byte("[{\"key\":\"flagkey\",\"variants\":{},\"segments\":[]}]")
var FLAG_1, _ = parseData(FLAG_1_STR)

func TestFlagConfigStreamApi(t *testing.T) {
	sse := mockSseStream{chConnected: make(chan bool)}
	api := newFlagConfigStreamApiV2("deploymentkey", "serverurl", 1*time.Second)
	api.newSseStreamFactory = sse.newSseStreamFactory
	receivedMsgCh := make(chan map[string]*evaluation.Flag)
	receivedErrCh := make(chan error)

	go func() {
		// On connect.
		<-sse.chConnected
		sse.messageCh <- streamEvent{data: FLAG_1_STR}
		assert.Equal(t, FLAG_1, <-receivedMsgCh)
	}()
	err := api.Connect(
		func(m map[string]*evaluation.Flag) error {
			receivedMsgCh <- m
			return nil
		},
		func(m map[string]*evaluation.Flag) error {
			receivedMsgCh <- m
			return nil
		},
		func(err error) { receivedErrCh <- err },
	)
	assert.Nil(t, err)

	go func() { sse.messageCh <- streamEvent{data: FLAG_1_STR} }()
	assert.Equal(t, FLAG_1, <-receivedMsgCh)
	go func() { sse.messageCh <- streamEvent{data: FLAG_1_STR} }()
	assert.Equal(t, FLAG_1, <-receivedMsgCh)

	api.Close()
}

func TestFlagConfigStreamApiErrorNoInitialFlags(t *testing.T) {
	sse := mockSseStream{chConnected: make(chan bool)}
	api := newFlagConfigStreamApiV2("deploymentkey", "serverurl", 1*time.Second)
	api.newSseStreamFactory = sse.newSseStreamFactory

	go func() {
		// On connect.
		<-sse.chConnected
	}()
	err := api.Connect(nil, nil, nil)
	assert.Equal(t, errors.New("flag config stream api connect timeout"), err)
}

func TestFlagConfigStreamApiErrorCorruptInitialFlags(t *testing.T) {
	sse := mockSseStream{chConnected: make(chan bool)}
	api := newFlagConfigStreamApiV2("deploymentkey", "serverurl", 1*time.Second)
	api.newSseStreamFactory = sse.newSseStreamFactory
	receivedMsgCh := make(chan map[string]*evaluation.Flag)
	receivedErrCh := make(chan error)

	go func() {
		// On connect.
		<-sse.chConnected
		sse.messageCh <- streamEvent{data: []byte("bad data")}
		<-receivedMsgCh // Should hang as no good data was received.
		assert.Fail(t, "Bad message went through")
	}()
	err := api.Connect(
		func(m map[string]*evaluation.Flag) error { receivedMsgCh <- m; return nil },
		func(m map[string]*evaluation.Flag) error { receivedMsgCh <- m; return nil },
		func(err error) { receivedErrCh <- err },
	)
	assert.Equal(t, "flag config stream api corrupt data", strings.Split(err.Error(), ", cause: ")[0])
}

func TestFlagConfigStreamApiErrorInitialFlagsUpdateFailStopsApi(t *testing.T) {
	sse := mockSseStream{chConnected: make(chan bool)}
	api := newFlagConfigStreamApiV2("deploymentkey", "serverurl", 1*time.Second)
	api.newSseStreamFactory = sse.newSseStreamFactory
	receivedMsgCh := make(chan map[string]*evaluation.Flag)
	receivedErrCh := make(chan error)

	go func() {
		// On connect.
		<-sse.chConnected
		sse.messageCh <- streamEvent{data: FLAG_1_STR}
		<-receivedMsgCh // Should hang as no updates was received.
		assert.Fail(t, "Bad message went through")
	}()
	err := api.Connect(
		func(m map[string]*evaluation.Flag) error { return errors.New("bad update") },
		func(m map[string]*evaluation.Flag) error { receivedMsgCh <- m; return nil },
		func(err error) { receivedErrCh <- err },
	)
	assert.Equal(t, errors.New("bad update"), err)
}

func TestFlagConfigStreamApiErrorInitialFlagsFutureUpdateFailDoesntStopApi(t *testing.T) {
	sse := mockSseStream{chConnected: make(chan bool)}
	api := newFlagConfigStreamApiV2("deploymentkey", "serverurl", 1*time.Second)
	api.newSseStreamFactory = sse.newSseStreamFactory
	receivedMsgCh := make(chan map[string]*evaluation.Flag)
	receivedErrCh := make(chan error)

	go func() {
		// On connect.
		<-sse.chConnected
		sse.messageCh <- streamEvent{data: FLAG_1_STR}
		assert.Equal(t, FLAG_1, <-receivedMsgCh) // Should hang as no updates was received.
	}()
	err := api.Connect(
		func(m map[string]*evaluation.Flag) error { receivedMsgCh <- m; return nil },
		func(m map[string]*evaluation.Flag) error { return errors.New("bad update") },
		func(err error) { receivedErrCh <- err },
	)
	assert.Nil(t, err)
	// Send an update, this should call onUpdate cb which fails.
	sse.messageCh <- streamEvent{data: FLAG_1_STR}
	// Make sure channel is not closed.
	sse.messageCh <- streamEvent{data: FLAG_1_STR}
}

func TestFlagConfigStreamApiErrorDuringStreaming(t *testing.T) {
	sse := mockSseStream{chConnected: make(chan bool)}
	api := newFlagConfigStreamApiV2("deploymentkey", "serverurl", 1*time.Second)
	api.newSseStreamFactory = sse.newSseStreamFactory
	receivedMsgCh := make(chan map[string]*evaluation.Flag)
	receivedErrCh := make(chan error)

	go func() {
		// On connect.
		<-sse.chConnected
		sse.messageCh <- streamEvent{data: FLAG_1_STR}
		assert.Equal(t, FLAG_1, <-receivedMsgCh)
	}()
	err := api.Connect(
		func(m map[string]*evaluation.Flag) error { receivedMsgCh <- m; return nil },
		func(m map[string]*evaluation.Flag) error { receivedMsgCh <- m; return nil },
		func(err error) { receivedErrCh <- err },
	)
	assert.Nil(t, err)

	go func() { sse.errorCh <- errors.New("error1") }()
	assert.Equal(t, errors.New("error1"), <-receivedErrCh)

	// The message channel should be closed.
	defer mutePanic(nil)
	sse.messageCh <- streamEvent{data: FLAG_1_STR}
	assert.Fail(t, "Unexpected message after error")
}
