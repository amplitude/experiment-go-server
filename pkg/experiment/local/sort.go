package local

import (
	"fmt"
	"github.com/amplitude/experiment-go-server/internal/evaluation"
)

func topologicalSort(flags map[string]*evaluation.Flag, flagKeys []string) ([]*evaluation.Flag, error) {
	result := make([]*evaluation.Flag, 0)
	// Extract keys and copy flags map
	keys := make([]string, 0)
	available := make(map[string]*evaluation.Flag)
	for k, v := range flags {
		keys = append(keys, k)
		available[k] = v
	}
	// Get the starting keys
	var startingKeys []string
	if len(flagKeys) > 0 {
		startingKeys = flagKeys
	} else {
		startingKeys = keys
	}
	// Sort into result
	for _, flagKey := range startingKeys {
		traversal, err := parentTraversal(flagKey, available, []string{})
		if err != nil {
			return nil, err
		}
		if len(traversal) > 0 {
			result = append(result, traversal...)
		}
	}
	return result, nil
}

func parentTraversal(flagKey string, available map[string]*evaluation.Flag, path []string) ([]*evaluation.Flag, error) {
	flag := available[flagKey]
	if flag == nil {
		return nil, nil
	}
	dependencies := flag.Dependencies
	if len(dependencies) == 0 {
		delete(available, flagKey)
		return []*evaluation.Flag{flag}, nil
	}
	path = append(path, flagKey)
	result := make([]*evaluation.Flag, 0)
	for _, parentKey := range dependencies {
		if contains(path, parentKey) {
			return nil, fmt.Errorf("detected a cycle between flags %v", path)
		}
		traversal, err := parentTraversal(parentKey, available, path)
		if err != nil {
			return nil, err
		}
		if len(traversal) > 0 {
			result = append(result, traversal...)
		}
	}
	result = append(result, flag)
	delete(available, flagKey)
	return result, nil
}
