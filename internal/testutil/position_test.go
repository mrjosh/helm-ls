package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var fileContent = `
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}
	`

func TestGetPositionOfMarkedLineInFile(t *testing.T) {
	markedLine := "{{- el^se }}"

	pos, found := GetPositionOfMarkedLineInFile(fileContent, markedLine, "^")

	assert.True(t, found)
	assert.Equal(t, uint32(3), pos.Line)
	assert.Equal(t, uint32(6), pos.Character)
}

func TestGetRangeOfMarkedLineInFile(t *testing.T) {
	markedLine := "{{- el^se^ }}"

	result, found := GetRangeOfMarkedLineInFile(fileContent, markedLine, "^")

	assert.True(t, found)
	assert.Equal(t, uint32(3), result.Start.Line)
	assert.Equal(t, uint32(6), result.Start.Character)
	assert.Equal(t, uint32(3), result.End.Line)
	assert.Equal(t, uint32(8), result.End.Character)
}
