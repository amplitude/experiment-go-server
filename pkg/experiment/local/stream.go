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
	tmsse "github.com/tmaxmax/go-sse"
)

// Keep alive data.
const STREAM_KEEP_ALIVE_BYTE = byte(' ')

// Mute panics caused by writing to a closed channel.
func mutePanic(f func()) {
	if err := recover(); err != nil && f != nil {
		f()
	}
}

type connCallback func()

type sseEvent struct {
	Data []byte
}

// eventSource is a boiled down version of an SSE client.
type eventSource interface {
	OnDisconnect(fn connCallback)
	OnConnect(fn connCallback)
	SubscribeChanRawWithContext(ctx context.Context, ch chan *sseEvent) error
}

type tmaxmaxEventSource struct {
	httpClient   *http.Client
	url          string
	headers      map[string]string
	onConnect    connCallback
	onDisconnect connCallback
}

func newEventSource(httpClient *http.Client, url string, headers map[string]string) eventSource {
	return &tmaxmaxEventSource{
		httpClient: httpClient,
		url:        url,
		headers:    headers,
	}
}

func (e *tmaxmaxEventSource) OnDisconnect(fn connCallback) {
	e.onDisconnect = fn
}

func (e *tmaxmaxEventSource) OnConnect(fn connCallback) {
	e.onConnect = fn
}

func (e *tmaxmaxEventSource) SubscribeChanRawWithContext(ctx context.Context, ch chan *sseEvent) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.url, http.NoBody)
	if err != nil {
		return err
	}
	for key, value := range e.headers {
		req.Header.Set(key, value)
	}

	connected := false
	signalConnect := func() {
		if connected {
			return
		}
		connected = true
		if e.onConnect != nil {
			e.onConnect()
		}
	}

	client := &tmsse.Client{
		HTTPClient: e.httpClient,
		ResponseValidator: func(res *http.Response) error {
			if err := tmsse.DefaultValidator(res); err != nil {
				return err
			}
			signalConnect()
			return nil
		},
		Backoff: tmsse.Backoff{
			MaxRetries: -1,
		},
	}
	conn := client.NewConnection(req)
	conn.Buffer(nil, 1<<32)

	conn.SubscribeToAll(func(ev tmsse.Event) {
		select {
		case <-ctx.Done():
			return
		case ch <- &sseEvent{Data: []byte(ev.Data)}:
		}
	})

	err = conn.Connect()
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return err
	}
	if connected && e.onDisconnect != nil {
		e.onDisconnect()
	}
	return err
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

	// Buffered size 1 to avoid goroutine leaks from TOCTOU races: the context
	// may be cancelled between the ctx.Done() check and the channel send, which
	// would leave a goroutine permanently blocked on an unbuffered channel.
	connectCh := make(chan bool, 1)
	esMsgCh := make(chan *sseEvent)
	esConnectErrCh := make(chan error, 1)
	esDisconnectCh := make(chan bool, 1)
	// Redirect on disconnect to a channel.
	client.OnDisconnect(func() {
		select {
		case <-ctx.Done(): // Cancelled.
			return
		case esDisconnectCh <- true: // Non-blocking due to buffer.
		}
	})
	// Redirect on connect to a channel.
	// No goroutine spawn needed: the buffered channel accepts the send without
	// blocking the SSE callback.
	client.OnConnect(func() {
		select {
		case <-ctx.Done(): // Cancelled.
			return
		case connectCh <- true: // Non-blocking due to buffer.
		}
	})
	go func() {
		// Subscribe to messages using channel.
		// This should be a non blocking call, but unsure how long it takes.
		err := client.SubscribeChanRawWithContext(ctx, esMsgCh)
		if err != nil {
			// Use select so this goroutine is never permanently blocked: if the
			// context was already cancelled (e.g. connection timed out and the
			// main event loop already exited), drop the error instead of leaking.
			select {
			case <-ctx.Done():
			case esConnectErrCh <- err:
			}
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
