package yamlhandler

import (
	"fmt"
	"regexp"
	"strconv"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// GetDiagnostics implements handler.LangHandler.
func (h *YamlHandler) GetDiagnostics(uri uri.URI) []protocol.PublishDiagnosticsParams {
	doc, ok := h.documents.GetYamlDoc(uri)

	if !ok {
		return nil
	}

	if doc.ParseErr == nil {
		logger.Debug("YamlHandler:  No parse error")
		return []protocol.PublishDiagnosticsParams{{
			URI:         uri,
			Diagnostics: []protocol.Diagnostic{},
		}}
	}

	errString := doc.ParseErr.Error()

	// find "line 4" in string

	re, err := regexp.Compile("line ([0-9]+): (.*)")
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return nil
	}

	matches := re.FindStringSubmatch(errString)

	if len(matches) < 2 {
		return nil
	}

	// convert to int

	line, err := strconv.Atoi(matches[1])
	if err != nil {
		fmt.Println("Error converting string to int:", err)
		return nil
	}

	return []protocol.PublishDiagnosticsParams{
		{
			URI: uri,
			Diagnostics: []protocol.Diagnostic{
				{
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      uint32(line - 1),
							Character: 0,
						},
						End: protocol.Position{
							Line:      uint32(line),
							Character: 0,
						},
					},
					Source:             "",
					Message:            matches[2],
					Tags:               []protocol.DiagnosticTag{},
					RelatedInformation: []protocol.DiagnosticRelatedInformation{},
					Data:               nil,
				},
			},
		},
	}
}
