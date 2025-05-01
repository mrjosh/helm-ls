package yamlhandler

import (
	"regexp"
	"strconv"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// find the pattern "line 4: message"
var lineNumberRegex = regexp.MustCompile("line ([0-9]+): (.*)")

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
	matches := lineNumberRegex.FindStringSubmatch(errString)

	if len(matches) < 3 {
		logger.Debug("YamlHandler: Regex pattern didn't match error format: %s", errString)
		return nil
	}

	// convert to int
	line, err := strconv.Atoi(matches[1])
	if err != nil {
		logger.Error("YamlHandler: Error converting string to int:", err)
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
					Source:             "Helm-ls YamlHandler",
					Message:            matches[2],
					Tags:               []protocol.DiagnosticTag{},
					RelatedInformation: []protocol.DiagnosticRelatedInformation{},
					Data:               nil,
				},
			},
		},
	}
}
