package helmlint

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/log"
	lsplocal "github.com/mrjosh/helm-ls/internal/lsp"
	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/lint/support"
)

var logger = log.GetLogger()

func GetDiagnosticsNotifications(chart *charts.Chart, doc *lsplocal.TemplateDocument) []lsp.PublishDiagnosticsParams {
	vals := chart.ValuesFiles.MainValuesFile.Values
	if chart.ValuesFiles.OverlayValuesFile != nil {
		vals = chartutil.CoalesceTables(chart.ValuesFiles.OverlayValuesFile.Values, chart.ValuesFiles.MainValuesFile.Values)
	}

	diagnostics := GetDiagnostics(chart.RootURI, vals)

	// Update the diagnostics cache only for the currently opened document
	// as it will also get diagnostics from yamlls
	// if currentDocDiagnostics is empty it means that all issues in that file have been fixed
	// we need to send this to the client
	currentDocDiagnostics := diagnostics[string(doc.URI.Filename())]
	doc.DiagnosticsCache.HelmDiagnostics = currentDocDiagnostics
	diagnostics[string(doc.URI.Filename())] = doc.DiagnosticsCache.GetMergedDiagnostics()

	result := []lsp.PublishDiagnosticsParams{}

	for diagnosticsURI, diagnostics := range diagnostics {
		result = append(result,
			lsp.PublishDiagnosticsParams{
				URI:         uri.File(diagnosticsURI),
				Diagnostics: diagnostics,
			},
		)
	}

	return result
}

// GetDiagnostics will run helm linter against the chart root URI using the given values
// and converts the helm.support.Message to lsp.Diagnostics
func GetDiagnostics(rootURI uri.URI, vals chartutil.Values) map[string][]lsp.Diagnostic {
	diagnostics := map[string][]lsp.Diagnostic{}

	client := action.NewLint()

	result := client.Run([]string{rootURI.Filename()}, vals)

	for _, msg := range result.Messages {
		d, relativeFilePath, _ := GetDiagnosticFromLinterErr(msg)
		absoluteFilePath := filepath.Join(rootURI.Filename(), string(relativeFilePath))
		if d != nil {
			diagnostics[absoluteFilePath] = append(diagnostics[absoluteFilePath], *d)
		}
	}
	logger.Println(fmt.Sprintf("helm lint: result for chart %s : %v", rootURI.Filename(), diagnostics))

	return diagnostics
}

func GetDiagnosticFromLinterErr(supMsg support.Message) (*lsp.Diagnostic, string, error) {
	severity := parseSeverity(supMsg)

	if strings.HasPrefix(supMsg.Path, "templates") {
		message, err := parseTemplatesMessage(supMsg, severity)
		path := getFilePathFromLinterErr(supMsg)
		if err != nil {
			return nil, "", err
		}
		return &message, path, nil
	}

	message := string(supMsg.Err.Error())
	// NOTE: The diagnostics may not be shown correctly in the Chart.yaml file in neovim
	// because the lsp is not active for that file
	if supMsg.Path == "Chart.yaml" || strings.Contains(message, "chart metadata") {
		return &lsp.Diagnostic{
			Severity: severity,
			Source:   "Helm lint",
			Message:  message,
		}, "Chart.yaml", nil
	}

	return nil, "", nil
}

func parseTemplatesMessage(supMsg support.Message, severity lsp.DiagnosticSeverity) (lsp.Diagnostic, error) {
	var (
		err         error
		line        int
		fileLine    = util.BetweenStrings(supMsg.Error(), "(", ")")
		fileLineArr = strings.Split(fileLine, ":")
	)
	if len(fileLineArr) < 2 {
		return lsp.Diagnostic{}, fmt.Errorf("linter Err contains no position information")
	}
	lineStr := fileLineArr[1]
	line, err = strconv.Atoi(lineStr)
	if err != nil {
		return lsp.Diagnostic{}, err
	}
	msgStr := util.AfterStrings(supMsg.Error(), "):")
	msg := strings.TrimSpace(msgStr)

	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: uint32(line - 1)},
			End:   lsp.Position{Line: uint32(line - 1)},
		},
		Severity: severity,
		Source:   "Helm lint",
		Message:  msg,
	}, nil
}

func parseSeverity(supMsg support.Message) lsp.DiagnosticSeverity {
	var severity lsp.DiagnosticSeverity
	switch supMsg.Severity {
	case support.ErrorSev:
		severity = lsp.DiagnosticSeverityError
	case support.WarningSev:
		severity = lsp.DiagnosticSeverityWarning
	case support.InfoSev:
		severity = lsp.DiagnosticSeverityInformation
	}
	return severity
}

func getFilePathFromLinterErr(msg support.Message) string {
	var (
		filename       string
		fileLine       = util.BetweenStrings(msg.Error(), "(", ")")
		file, _, found = strings.Cut(fileLine, ":")
	)

	if !found {
		return msg.Path
	}

	paths := strings.Split(file, "/")

	for i, p := range paths {
		if p == "templates" {
			filename = strings.Join(paths[i:], "/")
		}
	}

	return filename
}
