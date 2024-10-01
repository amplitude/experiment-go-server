package local

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"github.com/r3labs/sse/v2"
	"gopkg.in/cenkalti/backoff.v1"
)

const STREAM_KEEP_ALIVE_BYTE = byte(' ')

// Mute panics caused by writing to a closed channel.
func mutePanic(f func()) {
	if err := recover(); err != nil && f != nil {
		f()
	}
}

type EventSource struct {
	connect func() error
}

type NewEventSourceFactory = func (
	ctx context.Context,
	url string,
	connTimeout time.Duration,
	maxTime time.Duration,
	headers map[string]string,
	onConnect func(es *EventSource),
	onDisconnect func(es *EventSource),
	onUpdate func(es *EventSource, msg []byte),
) *EventSource

func NewEventSource(
	ctx context.Context,
	url string,
	connTimeout time.Duration,
	maxTime time.Duration,
	headers map[string]string,
	onConnect func(es *EventSource),
	onDisconnect func(es *EventSource),
	onUpdate func(es *EventSource, msg []byte),
) *EventSource {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   connTimeout,
		}).Dial,
		TLSHandshakeTimeout: connTimeout,
		ResponseHeaderTimeout: connTimeout,
	}

	// The http client timeout includes reading body, which is the entire SSE lifecycle until SSE is closed.
	httpClient := &http.Client{Transport: transport, Timeout: maxTime} // Max time for this connection.
	
	client := sse.NewClient(url)
	client.Connection = httpClient
	client.Headers = headers

	sse.ClientMaxBufferSize(1 << 32)(client)
	client.ReconnectStrategy = &backoff.StopBackOff{};

	es := &EventSource{}
	es.connect = func() error {
		return client.SubscribeRawWithContext(ctx, func(msg *sse.Event) {
			onUpdate(es, msg.Data)
		})
	}

	// On connect callback.
	client.OnConnect(func(c *sse.Client) {onConnect(es)})
	// On disconnect callback.
	client.OnDisconnect(func(c *sse.Client) {onDisconnect(es)})

	return es
}

type StreamEvent struct {
	data []byte
}

type SseStream struct {
	AuthToken                        string
	url                            string
    connectionTimeout time.Duration
    keepaliveTimeout time.Duration
    reconnInterval time.Duration
    maxJitter time.Duration
	lock                sync.Mutex
	cancelClientContext context.CancelFunc
	connTimeoutTimer *time.Timer
	keepaliveTimer *time.Timer
	reconnTimer *time.Timer
	NewEventSource NewEventSourceFactory
	es *EventSource
}

func NewSseStream(
	authToken, 
	url string,
    connectionTimeout time.Duration,
    keepaliveTimeout time.Duration,
    reconnInterval time.Duration,
    maxJitter time.Duration,
	newES NewEventSourceFactory,
) *SseStream {
	return &SseStream{
		AuthToken:                        authToken,
		url:                            url,
		connectionTimeout: connectionTimeout,
		keepaliveTimeout: keepaliveTimeout,
		reconnInterval: reconnInterval,
		maxJitter: maxJitter,
		NewEventSource: newES,
	}
}

func (s *SseStream) Connect(
	messageCh chan StreamEvent,
	errorCh chan error,
) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cancelInternal()

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelClientContext = cancel

	es := s.NewEventSource(
		ctx, 
		s.url, 
		s.connectionTimeout, 
		s.reconnInterval + s.maxJitter,
		map[string]string{
			"Authorization": s.AuthToken,
			"X-Amp-Exp-Library": fmt.Sprintf("experiment-go-server/%v", experiment.VERSION),
		},
		func (es *EventSource) {
			s.lock.Lock()
			defer s.lock.Unlock()
			if (es != s.es) {
				return
			}
			if (s.connTimeoutTimer != nil) {
				s.connTimeoutTimer.Stop()
			}
			s.resetKeepAliveTimeout(errorCh)
		},
		func (es *EventSource) {
			s.lock.Lock()
			if (es != s.es) {
				s.lock.Unlock()
				return
			}
			// Disconnected.
			s.cancelInternal()
			s.lock.Unlock()

			// Possible write to closed channel
			defer mutePanic(nil)
			errorCh <- errors.New("disconnected error")
		},
		func (es *EventSource, msg []byte) {
			s.lock.Lock()
			if (es != s.es) {
				s.lock.Unlock()
				return
			}
			// Reset keep alive.
			s.resetKeepAliveTimeout(errorCh)
			s.lock.Unlock()
			if (len(msg) == 1 && msg[0] == STREAM_KEEP_ALIVE_BYTE) {
				// Keep alive. 
				return
			}

			// Possible write to closed channel
			defer mutePanic(s.Cancel)
			messageCh <- StreamEvent{msg}
		},
	)
	s.es = es

	go func() {
		err := es.connect()
		if (err != nil) {
			s.lock.Lock()
			if (es != s.es) {
				s.lock.Unlock()
				return
			}
			s.cancelInternal()
			s.lock.Unlock()

			// Possible write to closed channel
			defer mutePanic(nil)
			errorCh <- err
		}
	}()
	s.reconnTimer = time.AfterFunc(
		randTimeDuration(s.reconnInterval, s.maxJitter),
	 	func() {
			s.lock.Lock()
			if (es != s.es) {
				s.lock.Unlock()
				return
			}
			s.reconnTimer = nil
			s.lock.Unlock()
			// Connect performs cancelInternal()
			s.Connect(messageCh, errorCh)
		},
	)
	s.connTimeoutTimer = time.AfterFunc(
		s.connectionTimeout,
		func() {
			s.lock.Lock()
			if (es != s.es) {
				s.lock.Unlock()
				return
			}
			s.connTimeoutTimer = nil
			s.cancelInternal()
			s.lock.Unlock()
			errorCh <- errors.New("timedout error")
		},
	)

	return nil
}

func (s *SseStream) cancelInternal() {
	s.es = nil
	if (s.connTimeoutTimer != nil) {
		s.connTimeoutTimer.Stop()
		s.connTimeoutTimer = nil
	}
	if (s.keepaliveTimer != nil) {
		s.keepaliveTimer.Stop()
		s.keepaliveTimer = nil
	}
	if (s.reconnTimer != nil) {
		s.reconnTimer.Stop()
		s.reconnTimer = nil
	}
	if (s.cancelClientContext != nil) {
		s.cancelClientContext()
		s.cancelClientContext = nil
	}
}

func (s *SseStream) Cancel() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelInternal()
}

func (s *SseStream) resetKeepAliveTimeout(errorCh chan error) {
	if (s.keepaliveTimer != nil) {
		s.keepaliveTimer.Stop()
	}
	s.keepaliveTimer = time.AfterFunc(s.keepaliveTimeout, func() {
		s.lock.Lock()
		s.keepaliveTimer = nil
		// Timed out, raise error.
		s.cancelInternal()
		s.lock.Unlock()

		errorCh <- errors.New("keep alive failed")
	})
}
