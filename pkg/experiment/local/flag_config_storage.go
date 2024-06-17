package local

import (
	"sync"
)

// FlagConfigStorage defines an interface for managing flag configurations.
type FlagConfigStorage interface {
	GetFlagConfig(key string) map[string]interface{}
	GetFlagConfigs() map[string]map[string]interface{}
	PutFlagConfig(flagConfig map[string]interface{})
	RemoveIf(condition func(map[string]interface{}) bool)
}

// InMemoryFlagConfigStorage is an in-memory implementation of FlagConfigStorage.
type InMemoryFlagConfigStorage struct {
	flagConfigs     map[string]map[string]interface{}
	flagConfigsLock sync.Mutex
}

// NewInMemoryFlagConfigStorage creates a new instance of InMemoryFlagConfigStorage.
func NewInMemoryFlagConfigStorage() *InMemoryFlagConfigStorage {
	return &InMemoryFlagConfigStorage{
		flagConfigs: make(map[string]map[string]interface{}),
	}
}

// GetFlagConfig retrieves a flag configuration by key.
func (storage *InMemoryFlagConfigStorage) GetFlagConfig(key string) map[string]interface{} {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()
	return storage.flagConfigs[key]
}

// GetFlagConfigs retrieves all flag configurations.
func (storage *InMemoryFlagConfigStorage) GetFlagConfigs() map[string]map[string]interface{} {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()
	copyFlagConfigs := make(map[string]map[string]interface{})
	for key, value := range storage.flagConfigs {
		copyFlagConfigs[key] = value
	}
	return copyFlagConfigs
}

// PutFlagConfig stores a flag configuration.
func (storage *InMemoryFlagConfigStorage) PutFlagConfig(flagConfig map[string]interface{}) {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()
	storage.flagConfigs[flagConfig["key"].(string)] = flagConfig
}

// RemoveIf removes flag configurations based on a condition.
func (storage *InMemoryFlagConfigStorage) RemoveIf(condition func(map[string]interface{}) bool) {
	storage.flagConfigsLock.Lock()
	defer storage.flagConfigsLock.Unlock()
	for key, value := range storage.flagConfigs {
		if condition(value) {
			delete(storage.flagConfigs, key)
		}
	}
}
