package languagefeatures

import (
	"github.com/mrjosh/helm-ls/internal/charts"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	sitter "github.com/smacker/go-tree-sitter"
)

type GenericDocumentUseCase struct {
	Document       *lsplocal.Document
	DocumentStore  *lsplocal.DocumentStore
	Chart          *charts.Chart
	Node           *sitter.Node
	ChartStore     *charts.ChartStore
	NodeType       string
	ParentNode     *sitter.Node
	ParentNodeType string
}

func (u *GenericDocumentUseCase) NodeContent() string {
	return u.Node.Content([]byte(u.Document.Content))
}
