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

func mutePanic() {
	if err := recover(); err != nil {
		// log.Println("panic occurred:", err)
	}
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
	keepaliveTimer *time.Timer
	reconnTimer *time.Timer
}

func NewSseStream(authToken, url string,
    connectionTimeout time.Duration,
    keepaliveTimeout time.Duration,
    reconnInterval time.Duration,
    maxJitter time.Duration) *SseStream {
	return &SseStream{
		AuthToken:                        authToken,
		url:                            url,
		connectionTimeout: connectionTimeout,
		keepaliveTimeout: keepaliveTimeout,
		reconnInterval: reconnInterval,
		maxJitter: maxJitter,
	}
}

func (s *SseStream) Connect(
	messageCh chan StreamEvent,
	errorCh chan error,
) error {
	s.Cancel()

	s.lock.Lock()
	defer s.lock.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelClientContext = cancel
	
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   s.connectionTimeout,
		}).Dial,
		TLSHandshakeTimeout: s.connectionTimeout,
		ResponseHeaderTimeout: s.connectionTimeout,
	}

	// The http client timeout includes reading body, which is the entire SSE lifecycle until SSE is closed.
	httpClient := &http.Client{Transport: transport, Timeout: s.reconnInterval + s.maxJitter} // Max time for this connection.
	
	client := sse.NewClient(s.url)
	client.Connection = httpClient
	client.Headers = map[string]string{
		"Authorization": s.AuthToken,
		"X-Amp-Exp-Library": fmt.Sprintf("experiment-go-server/%v", experiment.VERSION),
	}
	
	sse.ClientMaxBufferSize(1 << 32)(client)
	client.ReconnectStrategy = &backoff.StopBackOff{};
	client.OnConnect(func(c *sse.Client) {
		fmt.Println("connected")
		// Connected.
	})
	client.OnDisconnect(func(c *sse.Client) {
		// Disconnected.
		errorCh <- errors.New("disconnected error")
	})

	go func() {
		s.resetKeepAliveTimeout(errorCh)
		err := client.SubscribeRawWithContext(ctx, func(msg *sse.Event) {
			// Reset keep alive.
			s.resetKeepAliveTimeout(errorCh)
			data := string(msg.Data)
			if (data == " ") {
				// Keep alive. 
				return
			}

			// Possible write to closed channel
			defer mutePanic()
			messageCh <- StreamEvent{msg.Data}
		})
		if (err != nil) {
			s.Cancel()

			// Possible write to closed channel
			defer mutePanic()
			errorCh <- err
		}
	}()
	s.reconnTimer = time.AfterFunc(
		randTimeDuration(s.reconnInterval, s.maxJitter),
	 	func() {
			s.reconnTimer = nil
			s.Cancel()
			s.Connect(messageCh, errorCh)
		},
	)

	return nil
}

func (s *SseStream) cancelInternal() {
	if (s.keepaliveTimer != nil) {
		s.keepaliveTimer.Stop()
	}
	if (s.reconnTimer != nil) {
		s.reconnTimer.Stop()
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
	s.lock.Lock()
	defer s.lock.Unlock()

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
