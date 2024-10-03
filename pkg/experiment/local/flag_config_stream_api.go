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

type flagConfigStreamApiV2 struct {
	DeploymentKey                        string
	ServerURL                            string
    connectionTimeout time.Duration
	stopCh chan bool
	lock sync.Mutex
	newSseStreamFactory func (
		authToken, 
		url string,
		connectionTimeout time.Duration,
		keepaliveTimeout time.Duration,
		reconnInterval time.Duration,
		maxJitter time.Duration,
	) Stream
}

func NewFlagConfigStreamApiV2(
	deploymentKey                        string,
	serverURL                            string,
    connectionTimeout time.Duration,
) *flagConfigStreamApiV2 {
	return &flagConfigStreamApiV2{
		DeploymentKey:                        deploymentKey,
		ServerURL:                            serverURL,
		connectionTimeout: connectionTimeout,
		stopCh: nil,
		lock: sync.Mutex{},
		newSseStreamFactory: NewSseStream,
	}
}

func (api *flagConfigStreamApiV2) Connect(
	onInitUpdate func (map[string]*evaluation.Flag) error,
    onUpdate func (map[string]*evaluation.Flag) error,
    onError func (error),
) error {
	api.lock.Lock()
	defer api.lock.Unlock()

	err := api.closeInternal()
	if (err != nil) {
		return err
	}

	// Create URL.
	endpoint, err := url.Parse(api.ServerURL)
	if err != nil {
		return err
	}
	endpoint.Path = "sdk/stream/v1/flags"

	// Create Stream.
	stream := api.newSseStreamFactory("Api-Key " + api.DeploymentKey, endpoint.String(), api.connectionTimeout, streamApiKeepaliveTimeout, streamApiReconnInterval, streamApiMaxJitter)

	streamMsgCh := make(chan StreamEvent)
	streamErrCh := make(chan error)

	closeStream := func () {
		stream.Cancel()
		close(streamMsgCh)
		close(streamErrCh)
	}

	// Connect.
	err = stream.Connect(streamMsgCh, streamErrCh)
	if (err != nil) {
		return err
	}

	// Retrieve first flag configs and parse it.
	// If any error here means init error.
	select{
	case msg := <-streamMsgCh:
		// Parse message and verify data correct.
		flags, err := parseData(msg.data)
		if (err != nil) {
			closeStream()
			return errors.New("flag config stream api corrupt data, cause: " + err.Error())
		}
		if (onInitUpdate != nil) {
			err = onInitUpdate(flags)
		} else if (onUpdate != nil) {
			err = onUpdate(flags)
		}
		if (err != nil) {
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

	closeAllAndNotify := func(err error) {
		api.lock.Lock()
		defer api.lock.Unlock()
		closeStream()
		if (api.stopCh == stopCh) {
			api.stopCh = nil
		}
		close(stopCh)
		if (onError != nil) {
			onError(err)
		}
	}

	// Retrieve and pass on message forever until stopCh closes.
	go func() {
		for {
			select{
			case <-stopCh: // Channel returns immediately when closed. Note the local channel is referred here, so it's guaranteed to not be nil.
				closeStream()
				return
			case msg := <-streamMsgCh:
				// Parse message and verify data correct.
				flags, err := parseData(msg.data)
				if (err != nil) {
					// Error, close everything.
					closeAllAndNotify(errors.New("stream corrupt data, cause: " + err.Error()))
					return
				}
				if (onUpdate != nil) {
					// Deliver async. Don't care about any errors.
					go func() {onUpdate(flags)}()
				}
			case err := <-streamErrCh:
				// Error, close everything.
				closeAllAndNotify(err)
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

func (api *flagConfigStreamApiV2) closeInternal() error {
	if (api.stopCh != nil) {
		close(api.stopCh)
		api.stopCh = nil
	}
	return nil
}
func (api *flagConfigStreamApiV2) Close() error {
	api.lock.Lock()
	defer api.lock.Unlock()

	return api.closeInternal()
}