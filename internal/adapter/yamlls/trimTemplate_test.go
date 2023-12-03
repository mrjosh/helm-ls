package yamlls

import (
	"testing"

	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
)

type TrimTemplateTestData struct {
	documentText string
	trimmedText  string
}

var testTrimTemplateTestData = []TrimTemplateTestData{
	{
		documentText: `
{{ .Values.global. }}
yaml: test

{{block "name"}} T1 {{end}}
`,
		trimmedText: `
{{ .Values.global. }}
yaml: test

                 T1        
`,
	},
	{
		documentText: `
{{ .Values.global. }}
yaml: test

{{block "name"}} T1 {{end}}
`,

		trimmedText: `
{{ .Values.global. }}
yaml: test

                 T1        
`,
	},
	{
		documentText: `
{{ if eq .Values.service.myParameter "true" }}
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: prometheus-scaledobject
  namespace: default
spec:
  scaleTargetRef:
    name: hasd
  triggers:
  - type: prometheus
    metadata:
      serverAdress: http://<prometheus-host>:9090
      metricName: http_requests_total # DEPRECATED: This parameter is deprecated as of KEDA v2.10 and will be removed in version 2.12
      threshold: '100'
      query: sum(rate(http_requests_total{deployment="my-deployment"}[2m]))
# yaml-language-server: $schema=https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/keda.sh/scaledobject_v1alpha1.json
{{ end }}
`,
		trimmedText: `
                                              
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: prometheus-scaledobject
  namespace: default
spec:
  scaleTargetRef:
    name: hasd
  triggers:
  - type: prometheus
    metadata:
      serverAdress: http://<prometheus-host>:9090
      metricName: http_requests_total # DEPRECATED: This parameter is deprecated as of KEDA v2.10 and will be removed in version 2.12
      threshold: '100'
      query: sum(rate(http_requests_total{deployment="my-deployment"}[2m]))
# yaml-language-server: $schema=https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/keda.sh/scaledobject_v1alpha1.json
         
`,
	},
	{
		documentText: `
{{ if eq .Values.service.myParameter "true" }}
{{ if eq .Values.service.second "true" }}
apiVersion: keda.sh/v1alpha1
{{ end }}
{{ end }}
`,
		trimmedText: `
                                              
                                         
apiVersion: keda.sh/v1alpha1
         
         
`,
	},
	{
		documentText: `
{{- if .Values.ingress.enabled }}
apiVersion: apps/v1
kind: Ingress
{{- end }}
`,

		trimmedText: `
                                 
apiVersion: apps/v1
kind: Ingress
          
`,
	},
	{
		documentText: `
{{- if .Values.ingress.enabled }}
apiVersion: apps/v1
{{- else }}
apiVersion: apps/v2
{{- end }}
`,

		trimmedText: `
                                 
apiVersion: apps/v1
           
apiVersion: apps/v2
          
`,
	},
	{
		documentText: `
apiVersion: {{ include "common.capabilities.ingress.apiVersion" . }}
kind: Ingress
metadata:
  name: {{ include "common.names.fullname" . }}
  namespace: {{ .Release.Namespace | quote }}
  labels: {{- include "common.labels.standard" . | nindent 4 }}
    {{- if .Values.commonLabels }}
    {{- include "common.tplvalues.render" ( dict "value" .Values.commonLabels "context" $ ) | nindent 4 }}
    {{- end }}
    app.kubernetes.io/component: grafana
  annotations:
    {{- if .Values.ingress.certManager }}
    kubernetes.io/tls-acme: "true"
    {{- end }}
    {{- if .Values.ingress.annotations }}
    {{- include "common.tplvalues.render" (dict "value" .Values.ingress.annotations "context" $) | nindent 4 }}
    {{- end }}
    {{- if .Values.commonAnnotations }}
    {{- include "common.tplvalues.render" ( dict "value" .Values.commonAnnotations "context" $ ) | nindent 4 }}
    {{- end }}
		`,
		trimmedText: `
apiVersion: {{ include "common.capabilities.ingress.apiVersion" . }}
kind: Ingress
metadata:
  name: {{ include "common.names.fullname" . }}
  namespace: {{ .Release.Namespace | quote }}
  labels:                                                      
                                  
                                                                                                          
              
    app.kubernetes.io/component: grafana
  annotations:
                                         
    kubernetes.io/tls-acme: "true"
              
                                         
                                                                                                               
              
                                       
                                                                                                               
              
		`,
	},
	{documentText: `{{- $x := "test" -}}`, trimmedText: `                    `},
	{documentText: `{{ $x := "test" }}`, trimmedText: `                  `},
	{documentText: `{{ /* comment */ }}`, trimmedText: `                   `},
	{documentText: `{{define "name"}} T1 {{end}}`, trimmedText: `                            `},
	{
		documentText: `
          {{- if .Values.controller.customStartupProbe }}
          startupProbe: {}
          {{- else if .Values.controller.startupProbe.enabled }}
          startupProbe:
            httpGet:
              path: /healthz
              port: {{ .Values.controller.containerPorts.controller }}
          {{- end }}
	  `,
		trimmedText: `
                                                         
          startupProbe: {}
                                                                
          startupProbe:
            httpGet:
              path: /healthz
              port: {{ .Values.controller.containerPorts.controller }}
                    
	  `,
	},
	{
		documentText: `
      {{ if eq .Values.replicaCout 1 }}
      {{- $kube := ""  -}}
      apiVersion: v1
      kind: Service
      bka: dsa
      metadata:
        name: {{ include "hello-world.fullname" . }}
        labels:
          {{- include "hello-world.labels" . | nindent 4 }}
      spec:
        type: {{ .Values.service.type }}
        ports:
          - port:  {{ .Values.service.port }}
            targetPort: http
      {{ end }}
      `,
		trimmedText: `
                                       
                          
      apiVersion: v1
      kind: Service
      bka: dsa
      metadata:
        name: {{ include "hello-world.fullname" . }}
        labels:
                                                           
      spec:
        type: {{ .Values.service.type }}
        ports:
          - port:  {{ .Values.service.port }}
            targetPort: http
               
      `,
	},
	{
		// todo: Handle this case better
		documentText: `{{ if }}{{- end -}}`,
		trimmedText:  `      }}           `,
	},
	{
		documentText: `{{- $shards := $.Values.shards | int }}`,
		trimmedText:  `                                       `,
	},
	{
		documentText: `
{{- if $.Values.externalAccess.enabled }}
{{- $shards := $.Values.shards | int }}
{{- $replicas := $.Values.replicaCount | int }}
{{- $totalNodes := mul $shards $replicas }}
{{- range $shard, $e := until $shards }}
{{- range $i, $_e := until $replicas }}
{{- $targetPod := printf "%s-shard%d-%d" (include "common.names.fullname" $) $shard $i }}
{{- end }}
{{- end }}
{{- end }}
		`,
		trimmedText: `
                                         
                                       
                                               
                                           
                                        
                                       
                                                                                         
          
          
          
		`,
	},
}

func TestTrimTemplate(t *testing.T) {
	for _, testData := range testTrimTemplateTestData {
		testTrimTemplateWithTestData(t, testData)
	}
}

func testTrimTemplateWithTestData(t *testing.T, testData TrimTemplateTestData) {
	doc := &lsplocal.Document{
		Content: testData.documentText,
		Ast:     lsplocal.ParseAst(testData.documentText),
	}

	var trimmed = trimTemplateForYamllsFromAst(doc.Ast, testData.documentText)

	var result = trimmed == testData.trimmedText

	if !result {
		t.Errorf("Trimmed templated was not as expected but was %s ", trimmed)
	} else {
		t.Log("Trimmed templated was as expected")
	}
}
