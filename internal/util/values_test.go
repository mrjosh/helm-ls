package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyValues(t *testing.T) {
	_, err := GetTableOrValueForSelector(make(map[string]interface{}), "global")
	assert.Error(t, err)

	result, err := GetTableOrValueForSelector(make(map[string]interface{}), "")
	assert.NoError(t, err)
	assert.Equal(t, "{}\n", result)
}

func TestValues(t *testing.T) {
	nested := map[string]interface{}{"nested": "value"}
	values := map[string]interface{}{"global": nested}

	result := GetValueCompletion(values, []string{"g"})
	assert.Len(t, result, 1)
	assert.Equal(t, "global", result[0].InsertText)

	result = GetValueCompletion(values, []string{""})
	assert.Len(t, result, 1)
	assert.Equal(t, "global", result[0].InsertText)

	result = GetValueCompletion(values, []string{"global", "nes"})
	assert.Len(t, result, 1)
	assert.Equal(t, "nested", result[0].InsertText)
}

func TestWrongValues(t *testing.T) {
	nested := map[string]interface{}{"nested": 1}
	values := map[string]interface{}{"global": nested}

	_, err := GetTableOrValueForSelector(values, "some.wrong.values")
	assert.Error(t, err)

	_, err = GetTableOrValueForSelector(values, "some.wrong")
	assert.Error(t, err)

	_, err = GetTableOrValueForSelector(values, "some")
	assert.Error(t, err)
}
