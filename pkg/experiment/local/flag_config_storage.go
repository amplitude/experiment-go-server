package local

import (
	"github.com/amplitude/experiment-go-server/internal/evaluation"
	"sync"
)

type FlagConfigStorage interface {
	GetFlagConfig(key string) *evaluation.Flag
	GetFlagConfigs() map[string]*evaluation.Flag
	PutFlagConfig(flagConfig *evaluation.Flag)
	RemoveIf(condition func(*evaluation.Flag) bool)
}

type InMemoryFlagConfigStorage struct {
	flagConfigs     map[string]*evaluation.Flag
	flagConfigsLock sync.Mutex
}

func NewInMemoryFlagConfigStorage() *InMemoryFlagConfigStorage {
	return &InMemoryFlagConfigStorage{
		flagConfigs: make(map[string]*evaluation.Flag),
	}
}

func (storage *InMemoryFlagConfigStorage) GetFlagConfig(key string) *evaluation.Flag {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()
	return storage.flagConfigs[key]
}

func (storage *InMemoryFlagConfigStorage) GetFlagConfigs() map[string]*evaluation.Flag {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()
	copyFlagConfigs := make(map[string]*evaluation.Flag)
	for key, value := range storage.flagConfigs {
		copyFlagConfigs[key] = value
	}
	return copyFlagConfigs
}

func (storage *InMemoryFlagConfigStorage) PutFlagConfig(flagConfig *evaluation.Flag) {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()
	storage.flagConfigs[flagConfig.Key] = flagConfig
}

func (storage *InMemoryFlagConfigStorage) RemoveIf(condition func(*evaluation.Flag) bool) {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()
	for key, value := range storage.flagConfigs {
		if condition(value) {
			delete(storage.flagConfigs, key)
		}
	}
}
