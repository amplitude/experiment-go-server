package local

import (
	"errors"
	"testing"
	"time"

	"github.com/amplitude/experiment-go-server/internal/evaluation"
	"github.com/stretchr/testify/assert"
)

func createTestPollerObjs() (mockFlagConfigApi, flagConfigStorage, cohortStorage, *cohortLoader) {
	api := mockFlagConfigApi{}
	cohortDownloadAPI := &mockCohortDownloadApi{}
	flagConfigStorage := newInMemoryFlagConfigStorage()
	cohortStorage := newInMemoryCohortStorage()
	cohortLoader := newCohortLoader(cohortDownloadAPI, cohortStorage)
	return api, flagConfigStorage, cohortStorage, cohortLoader
}

func TestFlagConfigPoller(t *testing.T) {
	api, flagConfigStorage, cohortStorage, cohortLoader := createTestPollerObjs()

	poller := NewFlagConfigPoller(&api, &Config{FlagConfigPollerInterval: 1 * time.Second}, flagConfigStorage, cohortStorage, cohortLoader)
	errorCh := make(chan error)

	// Poller start normal.
	api.getFlagConfigsFunc = func() (map[string]*evaluation.Flag, error) {
		return FLAG_1, nil
	}
	poller.Start(func (e error) {
		errorCh <- e
	}) // Start should block for first poll.
	assert.Equal(t, FLAG_1, flagConfigStorage.getFlagConfigs()) // Test flags in storage.

	// Change up flags to empty.
	api.getFlagConfigsFunc = func() (map[string]*evaluation.Flag, error) {
		return map[string]*evaluation.Flag{}, nil
	}
	time.Sleep(1100 * time.Millisecond) // Sleep for poller to poll.
	assert.Equal(t, map[string]*evaluation.Flag{}, flagConfigStorage.getFlagConfigs()) // Test flags empty in storage.

	// Stop poller, make sure there's no more poll.
	poller.Stop()
	api.getFlagConfigsFunc = func() (map[string]*evaluation.Flag, error) {
		assert.Fail(t, "Unexpected poll")
		return nil, nil
	}
	time.Sleep(1100 * time.Millisecond) // Sleep for poller to poll.
}


func TestFlagConfigPollerStartFail(t *testing.T) {
	api, flagConfigStorage, cohortStorage, cohortLoader := createTestPollerObjs()

	poller := NewFlagConfigPoller(&api, &Config{FlagConfigPollerInterval: 1 * time.Second}, flagConfigStorage, cohortStorage, cohortLoader)
	errorCh := make(chan error)

	// Poller start normal.
	api.getFlagConfigsFunc = func() (map[string]*evaluation.Flag, error) {
		return nil, errors.New("start error")
	}
	err := poller.Start(func (e error) {
		errorCh <- e
	}) // Start should block for first poll.
	assert.Equal(t, errors.New("start error"), err) // Test flags in storage.
}

func TestFlagConfigPollerPollingFail(t *testing.T) {
	api, flagConfigStorage, cohortStorage, cohortLoader := createTestPollerObjs()

	poller := NewFlagConfigPoller(&api, &Config{FlagConfigPollerInterval: 1 * time.Second}, flagConfigStorage, cohortStorage, cohortLoader)
	errorCh := make(chan error)

	// Poller start normal.
	api.getFlagConfigsFunc = func() (map[string]*evaluation.Flag, error) {
		return FLAG_1, nil
	}
	poller.Start(func (e error) {
		errorCh <- e
	}) // Start should block for first poll.
	assert.Equal(t, FLAG_1, flagConfigStorage.getFlagConfigs()) // Test flags in storage.

	// Return error on poll.
	api.getFlagConfigsFunc = func() (map[string]*evaluation.Flag, error) {
		return nil, errors.New("flag error")
	}
	time.Sleep(1100 * time.Millisecond) // Sleep for poller to poll.
	assert.Equal(t, errors.New("flag error"), <- errorCh) // Error callback called.

	// Make sure there's no more poll.
	api.getFlagConfigsFunc = func() (map[string]*evaluation.Flag, error) {
		assert.Fail(t, "Unexpected poll")
		return nil, nil
	}
	time.Sleep(1100 * time.Millisecond) // Wait for a poll which should never happen.

	// Can start again.
	api.getFlagConfigsFunc = func() (map[string]*evaluation.Flag, error) {
		return map[string]*evaluation.Flag{}, nil
	}
	poller.Start(func (e error) {
		errorCh <- e
	})
	assert.Equal(t, map[string]*evaluation.Flag{}, flagConfigStorage.getFlagConfigs()) // Test flags in storage.
}


type mockFlagConfigStreamApi struct {
	connectFunc func(
		func (map[string]*evaluation.Flag) error,
		func (map[string]*evaluation.Flag) error,
		func (error),
	) error
	closeFunc func()
}
func (api *mockFlagConfigStreamApi) Connect(
	onInitUpdate func (map[string]*evaluation.Flag) error,
	onUpdate func (map[string]*evaluation.Flag) error,
	onError func (error),
) error {
	return api.connectFunc(onInitUpdate, onUpdate, onError)
}
func (api *mockFlagConfigStreamApi) Close() { api.closeFunc() }

func createTestStreamerObjs() (mockFlagConfigStreamApi, flagConfigStorage, cohortStorage, *cohortLoader) {
	api := mockFlagConfigStreamApi{}
	cohortDownloadAPI := &mockCohortDownloadApi{}
	flagConfigStorage := newInMemoryFlagConfigStorage()
	cohortStorage := newInMemoryCohortStorage()
	cohortLoader := newCohortLoader(cohortDownloadAPI, cohortStorage)
	return api, flagConfigStorage, cohortStorage, cohortLoader
}

func TestFlagConfigStreamer(t *testing.T) {
	api, flagConfigStorage, cohortStorage, cohortLoader := createTestStreamerObjs()

	streamer := NewFlagConfigStreamer(&api, &Config{FlagConfigPollerInterval: 1 * time.Second}, flagConfigStorage, cohortStorage, cohortLoader)
	errorCh := make(chan error)

	var updateCb func (map[string]*evaluation.Flag) error
	api.connectFunc = func(
		onInitUpdate func (map[string]*evaluation.Flag) error,
		onUpdate func (map[string]*evaluation.Flag) error,
		onError func (error),
	) error {
		onInitUpdate(FLAG_1)
		updateCb = onUpdate
		return nil
	}
	api.closeFunc = func() {
		updateCb = nil
	}

	// Streamer start normal.
	streamer.Start(func (e error) {
		errorCh <- e
	}) // Start should block for first set of flags.
	assert.Equal(t, FLAG_1, flagConfigStorage.getFlagConfigs()) // Test flags in storage.

	// Update flags with empty set.
	updateCb(map[string]*evaluation.Flag{})
	assert.Equal(t, map[string]*evaluation.Flag{}, flagConfigStorage.getFlagConfigs()) // Empty flags are updated.

	// Stop streamer.
	streamer.Stop()
	assert.Nil(t, updateCb) // Make sure stream Close is called.

	// Streamer start again.
	streamer.Start(func (e error) {
		errorCh <- e
	}) // Start should block for first set of flags.
	assert.Equal(t, FLAG_1, flagConfigStorage.getFlagConfigs()) // Test flags in storage.

	streamer.Stop()
}

func TestFlagConfigStreamerStartFail(t *testing.T) {
	api, flagConfigStorage, cohortStorage, cohortLoader := createTestStreamerObjs()

	streamer := NewFlagConfigStreamer(&api, &Config{FlagConfigPollerInterval: 1 * time.Second}, flagConfigStorage, cohortStorage, cohortLoader)
	errorCh := make(chan error)

	api.connectFunc = func(
		onInitUpdate func (map[string]*evaluation.Flag) error,
		onUpdate func (map[string]*evaluation.Flag) error,
		onError func (error),
	) error {
		return errors.New("api connect error")
	}
	api.closeFunc = func() {
	}

	// Streamer start.
	err := streamer.Start(func (e error) {
		errorCh <- e
	}) // Start should block for first set of flags, which is error.
	assert.Equal(t, errors.New("api connect error"), err)
}

func TestFlagConfigStreamerStreamingFail(t *testing.T) {
	api, flagConfigStorage, cohortStorage, cohortLoader := createTestStreamerObjs()

	streamer := NewFlagConfigStreamer(&api, &Config{FlagConfigPollerInterval: 1 * time.Second}, flagConfigStorage, cohortStorage, cohortLoader)
	errorCh := make(chan error)

	var updateCb func (map[string]*evaluation.Flag) error
	var errorCb func (error)
	api.connectFunc = func(
		onInitUpdate func (map[string]*evaluation.Flag) error,
		onUpdate func (map[string]*evaluation.Flag) error,
		onError func (error),
	) error {
		onInitUpdate(FLAG_1)
		updateCb = onUpdate
		errorCb = onError
		return nil
	}
	api.closeFunc = func() {
		updateCb = nil
		errorCb = nil
	}

	// Streamer start normal.
	streamer.Start(func (e error) {
		errorCh <- e
	}) // Start should block for first set of flags.
	assert.Equal(t, FLAG_1, flagConfigStorage.getFlagConfigs()) // Test flags in storage.

	// Stream error.
	go func() {errorCb(errors.New("stream error"))} ()
	assert.Equal(t, errors.New("stream error"), <-errorCh) // Error callback is called.
	assert.Nil(t, updateCb) // Make sure stream Close is called.
	assert.Nil(t, errorCb)

	// Streamer start again.
	flagConfigStorage.removeIf(func (f *evaluation.Flag) bool {return true})
	streamer.Start(func (e error) {
		errorCh <- e
	}) // Start should block for first set of flags.
	assert.Equal(t, FLAG_1, flagConfigStorage.getFlagConfigs()) // Test flags in storage.

	streamer.Stop()
}

type mockFlagConfigUpdater struct {
	startFunc func (func (error)) error
	stopFunc func ()
}

func (u *mockFlagConfigUpdater) Start(f func (error)) error { return u.startFunc(f) }
func (u *mockFlagConfigUpdater) Stop() { u.stopFunc() }

func TestFlagConfigFallbackRetryWrapper(t *testing.T) {
	main := mockFlagConfigUpdater{}
	var mainOnError func (error)
	main.startFunc = func (onError func (error)) error {
		mainOnError = onError
		return nil
	}
	main.stopFunc = func () { 
		mainOnError = nil 
	}
	fallback := mockFlagConfigUpdater{}	
	fallback.startFunc = func (onError func (error)) error {
		return nil
	}
	fallback.stopFunc = func () { 
	}
	w := NewFlagConfigFallbackRetryWrapper(&main, &fallback, 1 * time.Second, 0, true)
	err := w.Start(nil)
	assert.Nil(t, err)
	assert.NotNil(t, mainOnError)

	w.Stop()
	assert.Nil(t, mainOnError)
}

func TestFlagConfigFallbackRetryWrapperBothStartFail(t *testing.T) {
	main := mockFlagConfigUpdater{}
	var mainOnError func (error)
	main.startFunc = func (onError func (error)) error {
		mainOnError = onError
		return errors.New("main start error")
	}
	main.stopFunc = func () { 
		mainOnError = nil 
	}
	fallback := mockFlagConfigUpdater{}	
	fallback.startFunc = func (onError func (error)) error {
		return errors.New("fallback start error")
	}
	fallback.stopFunc = func () { 
	}
	w := NewFlagConfigFallbackRetryWrapper(&main, &fallback, 1 * time.Second, 0, true)
	err := w.Start(nil)
	assert.Equal(t, errors.New("fallback start error"), err)
	assert.NotNil(t, mainOnError)
	mainOnError = nil

	// Test no retry if start fail.
	time.Sleep(2000 * time.Millisecond)
	assert.Nil(t, mainOnError)
}

func TestFlagConfigFallbackRetryWrapperMainStartFailFallbackSuccess(t *testing.T) {
	main := mockFlagConfigUpdater{}
	var mainOnError func (error)
	main.startFunc = func (onError func (error)) error {
		mainOnError = onError
		return errors.New("main start error")
	}
	main.stopFunc = func () { 
		mainOnError = nil 
	}
	fallback := mockFlagConfigUpdater{}
	fallbackStopCh := make(chan bool)
	fallback.startFunc = func (onError func (error)) error {
		return nil
	}
	fallback.stopFunc = func () { 
		go func() {fallbackStopCh <- true} ()
	}
	w := NewFlagConfigFallbackRetryWrapper(&main, &fallback, 1 * time.Second, 0, true)
	err := w.Start(nil)
	assert.Nil(t, err)
	assert.NotNil(t, mainOnError)
	mainOnError = nil

	// Test retry if main start fail and fallback start success.
	time.Sleep(1100 * time.Millisecond)
	assert.NotNil(t, mainOnError) // Main started called.
	mainOnError = nil
	select {
	case <-fallbackStopCh: assert.Fail(t, "Unexpected fallback stopped")
	default:
	}

	// Test next retry success.
	main.startFunc = func (onError func (error)) error {
		mainOnError = onError
		return nil
	}
	time.Sleep(1100 * time.Millisecond)
	assert.NotNil(t, mainOnError) // Main errored.
	<-fallbackStopCh // Fallback stopped.

	w.Stop()
}

func TestFlagConfigFallbackRetryWrapperMainUpdatingFail(t *testing.T) {
	main := mockFlagConfigUpdater{}
	var mainOnError func (error)
	main.startFunc = func (onError func (error)) error {
		mainOnError = onError
		return nil
	}
	main.stopFunc = func () { 
		mainOnError = nil 
	}
	fallback := mockFlagConfigUpdater{}
	fallbackStartCh := make(chan bool)
	fallbackStopCh := make(chan bool)
	fallback.startFunc = func (onError func (error)) error {
		go func() {fallbackStartCh <- true} ()
		return nil
	}
	fallback.stopFunc = func () {}
	w := NewFlagConfigFallbackRetryWrapper(&main, &fallback, 1 * time.Second, 0, true)
	// Start success
	err := w.Start(nil)
	assert.Nil(t, err)
	assert.NotNil(t, mainOnError)
	select {
	case <-fallbackStartCh: assert.Fail(t, "Unexpected fallback started")
	default:
	}

	// Test main updating failed, fallback.
	fallback.stopFunc = func () { // Start tracking fallback stops (Start() may call stops).
		go func() {fallbackStopCh <- true} ()
	}
	mainOnError(errors.New("main updating error"))
	mainOnError = nil
	<-fallbackStartCh // Fallbacks started.
	select {
	case <-fallbackStopCh: assert.Fail(t, "Unexpected fallback stopped")
	default:
	}

	// Test retry start fail as main updating fail.
	main.startFunc = func (onError func (error)) error {
		mainOnError = onError
		return errors.New("main start error")
	}
	time.Sleep(1100 * time.Millisecond)
	assert.NotNil(t, mainOnError) // Main started called.
	mainOnError = nil
	select { // Test no changes made to fallback updater.
	case <-fallbackStartCh: assert.Fail(t, "Unexpected fallback started")
	case <-fallbackStopCh: assert.Fail(t, "Unexpected fallback stopped")
	default:
	}

	// Test next retry success.
	main.startFunc = func (onError func (error)) error {
		mainOnError = onError
		return nil
	}
	time.Sleep(1100 * time.Millisecond)
	assert.NotNil(t, mainOnError) // Main errored.
	select {
	case <-fallbackStartCh: assert.Fail(t, "Unexpected fallback stopped")
	default:
	}
	<-fallbackStopCh // Fallback stopped.

	w.Stop()

}

func TestFlagConfigFallbackRetryWrapperMainOnly(t *testing.T) {
	main := mockFlagConfigUpdater{}
	var mainOnError func (error)
	main.startFunc = func (onError func (error)) error {
		mainOnError = onError
		return nil
	}
	main.stopFunc = func () { 
		mainOnError = nil 
	}
	w := NewFlagConfigFallbackRetryWrapper(&main, nil, 1 * time.Second, 0, true)
	err := w.Start(nil)
	assert.Nil(t, err)
	assert.NotNil(t, mainOnError)

	// Signal updating error.
	mainOnError(errors.New("main error"))
	mainOnError = nil

	// Wait for retry and check.
	time.Sleep(1100 * time.Millisecond)
	assert.NotNil(t, mainOnError)
	mainOnError_2 := mainOnError
	mainOnError = nil

	// Check no more retrys after start success.
	time.Sleep(1100 * time.Millisecond)
	assert.Nil(t, mainOnError)

	// Again.
	// Signal updating error.
	mainOnError_2(errors.New("main error"))
	mainOnError = nil

	// Wait for retry and check.
	time.Sleep(1100 * time.Millisecond)
	assert.NotNil(t, mainOnError)
	mainOnError = nil

	// Check no more retrys after start success.
	time.Sleep(1100 * time.Millisecond)
	assert.Nil(t, mainOnError)

	w.Stop()
}