package gotemplate_test

import (
	"context"
	"testing"

	"github.com/mrjosh/helm-ls/internal/tree-sitter/gotemplate"
	sitter "github.com/smacker/go-tree-sitter"
)

func TestGrammar(t *testing.T) {

	n, _ := sitter.ParseCtx(context.Background(), []byte("{{ nil }}"), gotemplate.GetLanguage())

	if n.String() != "(template (nil))" {

		t.Errorf("Parsing did not work")

	}
}
