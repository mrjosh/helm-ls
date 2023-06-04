package yamlls

import (
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"testing"
)

func TestTrimTemplate(t *testing.T) {

	var documentText = `
{{ .Values.global. }}
yaml: test

{{block "name"}} T1 {{end}}
`

	var trimmedText = `
{{ .Values.global. }}
yaml: test

                 T1        
`
	doc := &lsplocal.Document{
		Content: documentText,
		Ast:     lsplocal.ParseAst(documentText),
	}

	var trimmed = trimTemplateForYamllsFromAst(doc.Ast, documentText)

	var result = trimmed == trimmedText

	if !result {
		t.Errorf("Trimmed templated was not as expected but was %s ", trimmed)
	} else {
		t.Log("Trimmed templated was as expected")
	}
}

func TestTrimTemplateFromAst(t *testing.T) {

	var documentText = `
{{ .Values.global. }}
yaml: test

{{block "name"}} T1 {{end}}
`

	var trimmedText = `
{{ .Values.global. }}
yaml: test

                 T1        
`
	doc := &lsplocal.Document{
		Content: documentText,
		Ast:     lsplocal.ParseAst(documentText),
	}

	var trimmed = trimTemplateForYamllsFromAst(doc.Ast, documentText)

	var result = trimmed == trimmedText

	if !result {
		t.Errorf("Trimmed templated was not as expected but was %s ", trimmed)
	} else {
		t.Log("Trimmed templated was as expected")
	}
}

func TestTrimTemplateFromAst2(t *testing.T) {

	var documentText = `
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
`

	var trimmedText = `
                                              
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
         
`
	doc := &lsplocal.Document{
		Content: documentText,
		Ast:     lsplocal.ParseAst(documentText),
	}

	var trimmed = trimTemplateForYamllsFromAst(doc.Ast, documentText)

	var result = trimmed == trimmedText

	if !result {
		t.Errorf("Trimmed templated was not as expected but was %s ", trimmed)
	} else {
		t.Log("Trimmed templated was as expected")
	}
}

func TestTrimTemplateFromAst3(t *testing.T) {

	var documentText = `
{{ if eq .Values.service.myParameter "true" }}
{{ if eq .Values.service.second "true" }}
apiVersion: keda.sh/v1alpha1
{{ end }}
{{ end }}
`

	var trimmedText = `
                                              
                                         
apiVersion: keda.sh/v1alpha1
         
         
`
	doc := &lsplocal.Document{
		Content: documentText,
		Ast:     lsplocal.ParseAst(documentText),
	}

	var trimmed = trimTemplateForYamllsFromAst(doc.Ast, documentText)

	var result = trimmed == trimmedText

	if !result {
		t.Errorf("Trimmed templated was not as expected but was %s ", trimmed)
	} else {
		t.Log("Trimmed templated was as expected")
	}
}

func TestTrimTemplateFromAst4(t *testing.T) {

	var documentText = `
{{- if .Values.ingress.enabled }}
apiVersion: apps/v1
kind: Ingress
{{- end }}
`

	var trimmedText = `
                                 
apiVersion: apps/v1
kind: Ingress
          
`
	doc := &lsplocal.Document{
		Content: documentText,
		Ast:     lsplocal.ParseAst(documentText),
	}

	var trimmed = trimTemplateForYamllsFromAst(doc.Ast, documentText)

	var result = trimmed == trimmedText

	if !result {
		t.Errorf("Trimmed templated was not as expected but was %s ", trimmed)
	} else {
		t.Log("Trimmed templated was as expected")
	}
}
