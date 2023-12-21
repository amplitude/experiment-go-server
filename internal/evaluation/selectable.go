package evaluation

import "reflect"

type selectable interface {
	Select(selector string) interface{}
}

func (t target) Select(selector string) interface{} {
	switch selector {
	case "context":
		return t.context
	case "result":
		return t.result
	default:
		return nil
	}
}

func (v Variant) Select(selector string) interface{} {
	switch selector {
	case "key":
		return v.Key
	case "value":
		return v.Value
	case "payload":
		return v.Payload
	case "metadata":
		return v.Metadata
	default:
		return nil
	}
}

func selectEach(s interface{}, selector []string) interface{} {
	if s == nil || selector == nil || len(selector) == 0 {
		return nil
	}
	for _, selectorElement := range selector {
		if s == nil {
			return nil
		}
		switch t := s.(type) {
		case selectable:
			s = t.Select(selectorElement)
		case map[string]interface{}:
			s = t[selectorElement]
		case map[string]Variant:
			s = t[selectorElement]
		default:
			// Fall back to reflection for maps with unexpected value types
			isMap := reflect.TypeOf(s).Kind() == reflect.Map
			if isMap {
				iter := reflect.ValueOf(s).MapRange()
				for iter.Next() {
					mapKey := iter.Key().String()
					mapValue := iter.Value().Interface()
					if mapKey == selectorElement {
						s = mapValue
					}
				}
			} else {
				return nil
			}
		}
	}
	return s
}
