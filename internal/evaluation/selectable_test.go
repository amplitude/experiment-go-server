package evaluation

import (
	"encoding/json"
	"reflect"
	"testing"
)

var primitiveObjectJson = `
{
  "null": null,
  "string": "value",
  "int": 13,
  "double": 13.12,
  "boolean": true,
  "array": [null, "value", 13, 13.12, true],
  "object": {
    "null": null,
    "string": "value",
    "int": 13,
    "double": 13.12,
    "boolean": true,
    "array": [null, "value", 13, 13.12, true],
	"object": {
      "null": null,
	  "string": "value",
	  "int": 13,
	  "double": 13.12,
	  "boolean": true
    }
  }
}
`

func TestUnstructuredJson(t *testing.T) {
	var object map[string]interface{}
	err := json.Unmarshal([]byte(primitiveObjectJson), &object)
	if err != nil {
		panic(err)
	}
	missingValue := selectEach(object, []string{"does", "not", "exist"})
	nullValue := selectEach(object, []string{"null"})
	stringValue := selectEach(object, []string{"string"})
	intValue := selectEach(object, []string{"int"})
	doubleValue := selectEach(object, []string{"double"})
	booleanValue := selectEach(object, []string{"boolean"})
	arrayValue := selectEach(object, []string{"array"})
	objectValue := selectEach(object, []string{"object"})

	if missingValue != nil {
		t.Fatalf("unexpected value %v", missingValue)
	}
	if nullValue != nil {
		t.Fatalf("unexpected value %v", nullValue)
	}
	if stringValue != "value" {
		t.Fatalf("unexpected value %v", stringValue)
	}
	if intValue != float64(13) {
		t.Fatalf("unexpected value %v", intValue)
	}
	if doubleValue != 13.12 {
		t.Fatalf("unexpected value %v", doubleValue)
	}
	if booleanValue != true {
		t.Fatalf("unexpected value %v", booleanValue)
	}
	if !reflect.DeepEqual(arrayValue, []interface{}{nil, "value", float64(13), 13.12, true}) {
		t.Fatalf("unexpected value %v", arrayValue)
	}
	if !reflect.DeepEqual(objectValue, map[string]interface{}{
		"null": nil,
		"string": "value",
		"int": float64(13),
		"double": 13.12,
		"boolean": true,
		"array": []interface{}{nil, "value", float64(13), 13.12, true},
		"object": map[string]interface{}{
			"null":    nil,
			"string":  "value",
			"int":     float64(13),
			"double":  13.12,
			"boolean": true,
		},
	}) {
		t.Fatalf("unexpected value %v", objectValue)
	}

	nestedMissingValue := selectEach(object, []string{"object", "does", "not", "exist"})
	nestedNullValue := selectEach(object, []string{"object", "null"})
	nestedStringValue := selectEach(object, []string{"object", "string"})
	nestedIntValue := selectEach(object, []string{"object", "int"})
	nestedDoubleValue := selectEach(object, []string{"object", "double"})
	nestedBooleanValue := selectEach(object, []string{"object", "boolean"})
	nestedArrayValue := selectEach(object, []string{"object", "array"})
	nestedObjectValue := selectEach(object, []string{"object", "object"})

	if nestedMissingValue != nil {
		t.Fatalf("unexpected value %v", nestedMissingValue)
	}
	if nestedNullValue != nil {
		t.Fatalf("unexpected value %v", nestedNullValue)
	}
	if nestedStringValue != "value" {
		t.Fatalf("unexpected value %v", nestedStringValue)
	}
	if nestedIntValue != float64(13) {
		t.Fatalf("unexpected value %v", nestedIntValue)
	}
	if nestedDoubleValue != 13.12 {
		t.Fatalf("unexpected value %v", nestedDoubleValue)
	}
	if nestedBooleanValue != true {
		t.Fatalf("unexpected value %v", nestedBooleanValue)
	}
	if !reflect.DeepEqual(nestedArrayValue, []interface{}{nil, "value", float64(13), 13.12, true}) {
		t.Fatalf("unexpected value %v", nestedArrayValue)
	}
	if !reflect.DeepEqual(nestedObjectValue, map[string]interface{}{
		"null": nil,
		"string": "value",
		"int": float64(13),
		"double": 13.12,
		"boolean": true,
	}) {
		t.Fatalf("unexpected value %v", nestedObjectValue)
	}

	nestedMissingValue2 := selectEach(object, []string{"object", "object", "does", "not", "exist"})
	nestedNullValue2 := selectEach(object, []string{"object", "object", "null"})
	nestedStringValue2 := selectEach(object, []string{"object", "object", "string"})
	nestedIntValue2 := selectEach(object, []string{"object", "object", "int"})
	nestedDoubleValue2 := selectEach(object, []string{"object", "object", "double"})
	nestedBooleanValue2 := selectEach(object, []string{"object", "object", "boolean"})

	if nestedMissingValue2 != nil {
		t.Fatalf("unexpected value %v", nestedMissingValue2)
	}
	if nestedNullValue2 != nil {
		t.Fatalf("unexpected value %v", nestedNullValue2)
	}
	if nestedStringValue2 != "value" {
		t.Fatalf("unexpected value %v", nestedStringValue2)
	}
	if nestedIntValue2 != float64(13) {
		t.Fatalf("unexpected value %v", nestedIntValue2)
	}
	if nestedDoubleValue2 != 13.12 {
		t.Fatalf("unexpected value %v", nestedDoubleValue2)
	}
	if nestedBooleanValue2 != true {
		t.Fatalf("unexpected value %v", nestedBooleanValue2)
	}
}

