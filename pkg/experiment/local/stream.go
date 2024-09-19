package local

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"github.com/r3labs/sse/v2"
	"gopkg.in/cenkalti/backoff.v1"
)

// type sseStream interface {
// 	getFlagConfigs() (map[string]*evaluation.Flag, error)
// }

func recoverPanic() {
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

func (a *SseStream) Connect(
	messageCh chan StreamEvent,
	errorCh chan error,
) error {
	a.Cancel()

	a.lock.Lock()
	defer a.lock.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	a.cancelClientContext = cancel
	
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   a.connectionTimeout,
		}).Dial,
		TLSHandshakeTimeout: a.connectionTimeout,
		ResponseHeaderTimeout: a.connectionTimeout,
	}

	// The http client timeout includes reading body, which is the entire SSE lifecycle until SSE is closed.
	httpClient := &http.Client{Transport: transport, Timeout: a.reconnInterval + a.maxJitter} // Max time for this connection.
	
	client := sse.NewClient(a.url)
	client.Connection = httpClient
	client.Headers = map[string]string{
		"Authorization": a.AuthToken,
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
		a.resetKeepAliveTimeout(errorCh)
		err := client.SubscribeRawWithContext(ctx, func(msg *sse.Event) {
			// Reset keep alive.
			a.resetKeepAliveTimeout(errorCh)
			data := string(msg.Data)
			if (data == " ") {
				// Keep alive. 
				return
			}

			// Possible write to closed channel
			defer recoverPanic()
			messageCh <- StreamEvent{msg.Data}
		})
		if (err != nil) {
			a.Cancel()

			// Possible write to closed channel
			defer recoverPanic()
			errorCh <- err
		}
	}()
	a.reconnTimer = time.NewTimer(a.reconnInterval - a.maxJitter + time.Duration(rand.Int64N(a.maxJitter.Nanoseconds() * 2)))
	go func() {
		<- a.reconnTimer.C
		a.Cancel()
		a.Connect(messageCh, errorCh)
	}()

	return nil
}

func (a *SseStream) Cancel() {
	a.lock.Lock()
	defer a.lock.Unlock()

	if (a.keepaliveTimer != nil) {
		if (!a.keepaliveTimer.Stop()) {
			select{
			case <- a.keepaliveTimer.C:
			default:
			}
		}
	}
	if (a.reconnTimer != nil) {
		if (!a.reconnTimer.Stop()) {
			select{
			case <- a.reconnTimer.C:
			default:
			}
		}
	}
	if (a.cancelClientContext != nil) {
		a.cancelClientContext()
		a.cancelClientContext = nil
	}
}

func (a *SseStream) resetKeepAliveTimeout(errorCh chan error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if (a.keepaliveTimer != nil) {
		if (!a.keepaliveTimer.Stop()) {
			select{
			case <- a.keepaliveTimer.C:
			default:
			}
		}
	}
	a.keepaliveTimer = time.NewTimer(a.keepaliveTimeout)
	go func() {
		<- a.keepaliveTimer.C
		// Timed out, raise error.
		a.Cancel()
		errorCh <- errors.New("keep alive failed")
	}()
}
