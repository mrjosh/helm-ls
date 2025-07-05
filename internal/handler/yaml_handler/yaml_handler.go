package yamlhandler

import (
	"context"
	"fmt"

	"github.com/mrjosh/helm-ls/internal/adapter/yamlls"
	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/jsonschema"
	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/lsp/document"
	"go.lsp.dev/protocol"
	lsp "go.lsp.dev/protocol"
)

var logger = log.GetLogger()

type YamlHandler struct {
	documents       *document.DocumentStore
	chartStore      *charts.ChartStore
	client          protocol.Client
	yamllsConnector *yamlls.Connector
	jsonSchemas     *jsonschema.JSONSchemaCache
}

// SetClient implements handler.LangHandler.
func (h *YamlHandler) SetClient(client protocol.Client) {
	h.client = client
}

func NewYamlHandler(client protocol.Client, documents *document.DocumentStore, chartStore *charts.ChartStore) *YamlHandler {
	jsonSchemas, err := jsonschema.NewJSONSchemaCache(jsonschema.JSONSchemaConfig{}, chartStore)
	if err != nil || jsonSchemas == nil {
		client.ShowMessage(context.Background(), &protocol.ShowMessageParams{
			Type: protocol.MessageTypeError, Message: fmt.Sprintf("Helm-ls: Failed to create JSON schema cache: %s", err.Error()),
		})
	}
	return &YamlHandler{
		documents:       documents,
		chartStore:      chartStore,
		yamllsConnector: &yamlls.Connector{},
		jsonSchemas:     jsonSchemas,
	}
}

func (h *YamlHandler) setYamllsConnector(yamllsConnector *yamlls.Connector) {
	h.yamllsConnector = yamllsConnector
}

func (h *YamlHandler) getYamlPath(uri lsp.URI, pos lsp.Position) (path string, err error) {
	doc, ok := h.documents.GetYamlDoc(uri)

	if !ok {
		return "", fmt.Errorf("document not found: %s", uri)
	}

	return doc.GetPathForPosition(pos)
}
