package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyValues(t *testing.T) {
	_, err := GetTableOrValueForSelector(make(map[string]interface{}), []string{"global"})
	assert.Error(t, err)

	result, err := GetTableOrValueForSelector(make(map[string]interface{}), []string{""})
	assert.NoError(t, err)
	assert.Equal(t, "{}\n", result)
}

func TestValues(t *testing.T) {
	nested := map[string]interface{}{"nested": "value"}
	values := map[string]interface{}{"global": nested}

	result := GetValueCompletion(values, []string{"global", "nes"})
	assert.Len(t, result, 1)
	assert.Equal(t, "nested", result[0].InsertText)

	result = GetValueCompletion(values, []string{"g"})
	assert.Len(t, result, 1)
	assert.Equal(t, "global", result[0].InsertText)

	result = GetValueCompletion(values, []string{""})
	assert.Len(t, result, 1)
	assert.Equal(t, "global", result[0].InsertText)
}

func TestWrongValues(t *testing.T) {
	nested := map[string]interface{}{"nested": 1}
	values := map[string]interface{}{"global": nested}

	_, err := GetTableOrValueForSelector(values, []string{"some", "wrong", "values"})
	assert.Error(t, err)

	_, err = GetTableOrValueForSelector(values, []string{"some", "wrong"})
	assert.Error(t, err)

	_, err = GetTableOrValueForSelector(values, []string{"some"})
	assert.Error(t, err)
}

func TestValuesList(t *testing.T) {
	nested := []interface{}{1, 2, 3}
	values := map[string]interface{}{"global": nested}

	result, err := GetTableOrValueForSelector(values, []string{"global[]"})
	assert.NoError(t, err)
	assert.Equal(t, "1", result)
}

func TestValuesListNested(t *testing.T) {
	doubleNested := []interface{}{1, 2, 3}
	nested := map[string]interface{}{"nested": doubleNested}
	values := map[string]interface{}{"global": nested}

	result, err := GetTableOrValueForSelector(values, []string{"global", "nested[]"})
	assert.NoError(t, err)
	assert.Equal(t, "1", result)
}

func TestValuesRangeLookupOnMapping(t *testing.T) {
	doubleNested := map[string]interface{}{"a": 1}
	nested := map[string]interface{}{"nested": doubleNested, "other": "value"}
	values := map[string]interface{}{"global": nested}

	result, err := GetTableOrValueForSelector(values, []string{"global[]"})
	assert.NoError(t, err)
	assert.Equal(t, "a: 1\n", result)
}
