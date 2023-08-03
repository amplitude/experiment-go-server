package local

import "fmt"

func topologicalSort(flags map[string]interface{}, flagKeys []string) ([]interface{}, error) {
	result := make([]interface{}, 0)
	// Extract keys and copy flags map
	keys := make([]string, 0)
	available := make(map[string]interface{})
	for k, v := range flags {
		keys = append(keys, k)
		available[k] = v
	}
	// Get the starting keys
	startingKeys := make([]string, 0)
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

func parentTraversal(flagKey string, available map[string]interface{}, path []string) ([]interface{}, error) {
	flag := available[flagKey]
	if flag == nil {
		return nil, nil
	}
	dependencies := extractDependencies(flag)
	if len(dependencies) == 0 {
		delete(available, flagKey)
		return []interface{}{flag}, nil
	}
	path = append(path, flagKey)
	result := make([]interface{}, 0)
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
	path = path[:len(path)-1]
	delete(available, flagKey)
	return result, nil
}

func extractDependencies(flag interface{}) []string {
	switch f := flag.(type) {
	case map[string]interface{}:
		parentDependenciesAny := f["parentDependencies"]
		if parentDependenciesAny == nil {
			return nil
		}
		switch parentDependencies := parentDependenciesAny.(type) {
		case map[string]interface{}:
			flagsAny := parentDependencies["flags"]
			if flagsAny == nil {
				return nil
			}
			switch flags := flagsAny.(type) {
			case map[string]interface{}:
				result := make([]string, 0)
				for k, _ := range flags {
					result = append(result, k)
				}
				return result
			}
		}
	}
	return nil
}
