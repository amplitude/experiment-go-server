package local

import (
	"sync"

	"github.com/amplitude/experiment-go-server/internal/evaluation"
)

type flagConfigStorage interface {
	getFlagConfig(key string) *evaluation.Flag
	getFlagConfigs() map[string]*evaluation.Flag
	getFlagConfigsArray() []*evaluation.Flag
	putFlagConfig(flagConfig *evaluation.Flag)
	removeIf(condition func(*evaluation.Flag) bool)
}

type inMemoryFlagConfigStorage struct {
	flagConfigs     map[string]*evaluation.Flag
	flagConfigsLock sync.Mutex
}

func newInMemoryFlagConfigStorage() *inMemoryFlagConfigStorage {
	return &inMemoryFlagConfigStorage{
		flagConfigs: make(map[string]*evaluation.Flag),
	}
}

func (storage *inMemoryFlagConfigStorage) GetFlagConfigs() map[string]*evaluation.Flag {
	return storage.getFlagConfigs()
}

func (storage *inMemoryFlagConfigStorage) getFlagConfig(key string) *evaluation.Flag {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()
	return storage.flagConfigs[key]
}

func (storage *inMemoryFlagConfigStorage) getFlagConfigs() map[string]*evaluation.Flag {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()
	copyFlagConfigs := make(map[string]*evaluation.Flag)
	for key, value := range storage.flagConfigs {
		copyFlagConfigs[key] = value
	}
	return copyFlagConfigs
}

func (storage *inMemoryFlagConfigStorage) getFlagConfigsArray() []*evaluation.Flag {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()

	var copyFlagConfigs []*evaluation.Flag
	for _, value := range storage.flagConfigs {
		copyFlagConfigs = append(copyFlagConfigs, value)
	}
	return copyFlagConfigs
}

func (storage *inMemoryFlagConfigStorage) putFlagConfig(flagConfig *evaluation.Flag) {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()
	storage.flagConfigs[flagConfig.Key] = flagConfig
}

func (storage *inMemoryFlagConfigStorage) removeIf(condition func(*evaluation.Flag) bool) {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()
	for key, value := range storage.flagConfigs {
		if condition(value) {
			delete(storage.flagConfigs, key)
		}
	}
}
