package local

import (
	"encoding/json"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/amplitude/experiment-go-server/internal/evaluation"
)

const streamApiMaxJitter = 5 * time.Second
const streamApiKeepaliveTimeout = 17 * time.Second
const streamApiReconnInterval = 15 * time.Minute

type flagConfigStreamApi interface {
	Connect(
		onInitUpdate func(map[string]*evaluation.Flag) error,
		onUpdate func(map[string]*evaluation.Flag) error,
		onError func(error),
	) error
	Close()
}

type flagConfigStreamApiV2 struct {
	DeploymentKey       string
	ServerURL           string
	connectionTimeout   time.Duration
	stopCh              chan bool
	lock                sync.Mutex
	newSseStreamFactory func(
		authToken,
		url string,
		connectionTimeout time.Duration,
		keepaliveTimeout time.Duration,
		reconnInterval time.Duration,
		maxJitter time.Duration,
	) stream
}

func newFlagConfigStreamApiV2(
	deploymentKey string,
	serverURL string,
	connectionTimeout time.Duration,
) *flagConfigStreamApiV2 {
	return &flagConfigStreamApiV2{
		DeploymentKey:       deploymentKey,
		ServerURL:           serverURL,
		connectionTimeout:   connectionTimeout,
		stopCh:              nil,
		lock:                sync.Mutex{},
		newSseStreamFactory: newSseStream,
	}
}

func (api *flagConfigStreamApiV2) Connect(
	onInitUpdate func(map[string]*evaluation.Flag) error,
	onUpdate func(map[string]*evaluation.Flag) error,
	onError func(error),
) error {
	api.lock.Lock()
	defer api.lock.Unlock()

	api.closeInternal()

	// Create URL.
	endpoint, err := url.Parse(api.ServerURL)
	if err != nil {
		return err
	}
	endpoint.Path = "sdk/stream/v1/flags"

	// Create Stream.
	stream := api.newSseStreamFactory("Api-Key "+api.DeploymentKey, endpoint.String(), api.connectionTimeout, streamApiKeepaliveTimeout, streamApiReconnInterval, streamApiMaxJitter)

	streamMsgCh := make(chan streamEvent)
	streamErrCh := make(chan error)

	closeStream := func() {
		stream.Cancel()
		close(streamMsgCh)
		close(streamErrCh)
	}

	// Connect.
	stream.Connect(streamMsgCh, streamErrCh)

	// Retrieve first flag configs and parse it.
	// If any error here means init error.
	select {
	case msg := <-streamMsgCh:
		// Parse message and verify data correct.
		flags, err := parseData(msg.data)
		if err != nil {
			closeStream()
			return errors.New("flag config stream api corrupt data, cause: " + err.Error())
		}
		if onInitUpdate != nil {
			err = onInitUpdate(flags)
		} else if onUpdate != nil {
			err = onUpdate(flags)
		}
		if err != nil {
			closeStream()
			return err
		}
	case err := <-streamErrCh:
		// Error when creating the stream.
		closeStream()
		return err
	case <-time.After(api.connectionTimeout):
		// Timed out.
		closeStream()
		return errors.New("flag config stream api connect timeout")
	}

	// Prep procedures for stopping.
	stopCh := make(chan bool)
	api.stopCh = stopCh

	closeAll := func() {
		api.lock.Lock()
		defer api.lock.Unlock()
		closeStream()
		if api.stopCh == stopCh {
			api.stopCh = nil
		}
		close(stopCh)
	}

	// Retrieve and pass on message forever until stopCh closes.
	go func() {
		for {
			select {
			case <-stopCh: // Channel returns immediately when closed. Note the local channel is referred here, so it's guaranteed to not be nil.
				closeStream()
				return
			case msg := <-streamMsgCh:
				// Parse message and verify data correct.
				flags, err := parseData(msg.data)
				if err != nil {
					// Error, close everything.
					closeAll()
					if onError != nil {
						onError(errors.New("stream corrupt data, cause: " + err.Error()))
					}
					return
				}
				if onUpdate != nil {
					// Deliver async. Don't care about any errors.
					//nolint:errcheck
					go func() { onUpdate(flags) }()
				}
			case err := <-streamErrCh:
				// Error, close everything.
				closeAll()
				if onError != nil {
					onError(err)
				}
				return
			}
		}
	}()

	return nil
}

func parseData(data []byte) (map[string]*evaluation.Flag, error) {

	var flagsArray []*evaluation.Flag
	err := json.Unmarshal(data, &flagsArray)
	if err != nil {
		return nil, err
	}
	flags := make(map[string]*evaluation.Flag)
	for _, flag := range flagsArray {
		flags[flag.Key] = flag
	}

	return flags, nil
}

func (api *flagConfigStreamApiV2) closeInternal() {
	if api.stopCh != nil {
		close(api.stopCh)
		api.stopCh = nil
	}
}
func (api *flagConfigStreamApiV2) Close() {
	api.lock.Lock()
	defer api.lock.Unlock()

	api.closeInternal()
}
