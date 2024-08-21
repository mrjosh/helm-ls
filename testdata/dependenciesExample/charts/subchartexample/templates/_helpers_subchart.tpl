{{/*
Expand the name of the chart.
*/}}
{{- define "subchartexample.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}
