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

// Keep alive data.
const STREAM_KEEP_ALIVE_BYTE = byte(' ')

// Mute panics caused by writing to a closed channel.
func mutePanic(f func()) {
	if err := recover(); err != nil && f != nil {
		f()
	}
}

// This is a boiled down version of sse.Client.
type eventSource interface {
	OnDisconnect(fn sse.ConnCallback)
	OnConnect(fn sse.ConnCallback)
	SubscribeChanRawWithContext(ctx context.Context, ch chan *sse.Event) error
}

func newEventSource(httpClient *http.Client, url string, headers map[string]string) eventSource {
	client := sse.NewClient(url)
	client.Connection = httpClient
	client.Headers = headers
	sse.ClientMaxBufferSize(1 << 32)(client)
	client.ReconnectStrategy = &backoff.StopBackOff{}
	return client
}

type streamEvent struct {
	data []byte
}

type stream interface {
	Connect(messageCh chan streamEvent, errorCh chan error)
	Cancel()
	// For testing.
	setNewESFactory(f func(httpClient *http.Client, url string, headers map[string]string) eventSource)
}

type sseStream struct {
	AuthToken           string
	url                 string
	connectionTimeout   time.Duration
	keepaliveTimeout    time.Duration
	reconnInterval      time.Duration
	maxJitter           time.Duration
	lock                sync.Mutex
	cancelClientContext *context.CancelFunc
	newESFactory        func(httpClient *http.Client, url string, headers map[string]string) eventSource
}

func newSseStream(
	authToken,
	url string,
	connectionTimeout time.Duration,
	keepaliveTimeout time.Duration,
	reconnInterval time.Duration,
	maxJitter time.Duration,
) stream {
	return &sseStream{
		AuthToken:         authToken,
		url:               url,
		connectionTimeout: connectionTimeout,
		keepaliveTimeout:  keepaliveTimeout,
		reconnInterval:    reconnInterval,
		maxJitter:         maxJitter,
		newESFactory:      newEventSource,
	}
}

func (s *sseStream) setNewESFactory(f func(httpClient *http.Client, url string, headers map[string]string) eventSource) {
	s.newESFactory = f
}

func (s *sseStream) Connect(
	messageCh chan streamEvent,
	errorCh chan error,
) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.connectInternal(messageCh, errorCh)
}

func (s *sseStream) connectInternal(
	messageCh chan streamEvent,
	errorCh chan error,
) {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelClientContext = &cancel

	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: s.connectionTimeout,
		}).Dial,
		TLSHandshakeTimeout:   s.connectionTimeout,
		ResponseHeaderTimeout: s.connectionTimeout,
	}

	// The http client timeout includes reading body, which is the entire SSE lifecycle until SSE is closed.
	httpClient := &http.Client{Transport: transport, Timeout: s.reconnInterval + s.maxJitter} // Max time for this connection.

	client := s.newESFactory(httpClient, s.url, map[string]string{
		"Authorization":     s.AuthToken,
		"X-Amp-Exp-Library": fmt.Sprintf("experiment-go-server/%v", experiment.VERSION),
	})

	connectCh := make(chan bool)
	esMsgCh := make(chan *sse.Event)
	esConnectErrCh := make(chan error)
	esDisconnectCh := make(chan bool)
	// Redirect on disconnect to a channel.
	client.OnDisconnect(func(s *sse.Client) {
		select {
		case <-ctx.Done(): // Cancelled.
			return
		default:
			esDisconnectCh <- true
		}
	})
	// Redirect on connect to a channel.
	client.OnConnect(func(s *sse.Client) {
		select {
		case <-ctx.Done(): // Cancelled.
			return
		default:
			go func() { connectCh <- true }()
		}
	})
	go func() {
		// Subscribe to messages using channel.
		// This should be a non blocking call, but unsure how long it takes.
		err := client.SubscribeChanRawWithContext(ctx, esMsgCh)
		if err != nil {
			esConnectErrCh <- err
		}
	}()

	cancelWithLock := func() {
		s.lock.Lock()
		defer s.lock.Unlock()
		cancel()
		if s.cancelClientContext == &cancel {
			s.cancelClientContext = nil
		}
	}
	go func() {
		// First wait for connect.
		select {
		case <-ctx.Done(): // Cancelled.
			return
		case err := <-esConnectErrCh: // Channel subscribe error.
			cancelWithLock()
			defer mutePanic(nil)
			errorCh <- err
			return
		case <-time.After(s.connectionTimeout): // Timeout.
			cancelWithLock()
			defer mutePanic(nil)
			errorCh <- errors.New("stream connection timeout")
			return
		case <-connectCh: // Connected callbacked.
		}
		for {
			select { // Forced priority on context done.
			case <-ctx.Done(): // Cancelled.
				return
			default:
			}
			select {
			case <-ctx.Done(): // Cancelled.
				return
			case <-esDisconnectCh: // Disconnected.
				cancelWithLock()
				defer mutePanic(nil)
				errorCh <- errors.New("stream disconnected error")
				return
			case event := <-esMsgCh: // Message received.
				if len(event.Data) == 1 && event.Data[0] == STREAM_KEEP_ALIVE_BYTE {
					// Keep alive.
					continue
				}
				// Possible write to closed channel
				// If channel closed, cancel.
				defer mutePanic(cancelWithLock)
				messageCh <- streamEvent{event.Data}
			case <-time.After(s.keepaliveTimeout): // Keep alive timeout.
				cancelWithLock()
				defer mutePanic(nil)
				errorCh <- errors.New("stream keepalive timed out")
			}
		}
	}()

	// Reconnect after interval.
	time.AfterFunc(randTimeDuration(s.reconnInterval, s.maxJitter), func() {
		select {
		case <-ctx.Done(): // Cancelled.
			return
		default: // Reconnect.
			cancelWithLock()
			s.connectInternal(messageCh, errorCh)
			return
		}
	})
}

func (s *sseStream) Cancel() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.cancelClientContext != nil {
		(*(s.cancelClientContext))()
		s.cancelClientContext = nil
	}
}
