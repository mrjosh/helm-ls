package lsp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/mrjosh/helm-ls/pkg/action"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	"github.com/mrjosh/helm-ls/pkg/lint/support"
	"github.com/pkg/errors"

	// nolint
	"github.com/mrjosh/helm-ls/pkg/lint/rules"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

var logger = log.GetLogger()

func NotifcationFromLint(ctx context.Context, conn jsonrpc2.Conn, chart *charts.Chart, doc *Document) (*jsonrpc2.Notification, error) {
	vals := chart.ValuesFiles.MainValuesFile.Values
	if chart.ValuesFiles.OverlayValuesFile != nil {
		vals = chartutil.CoalesceTables(chart.ValuesFiles.OverlayValuesFile.Values, chart.ValuesFiles.MainValuesFile.Values)
	}

	diagnostics, err := GetDiagnostics(doc.URI, vals)
	if err != nil {
		return nil, err
	}
	doc.DiagnosticsCache.HelmDiagnostics = diagnostics

	return nil, conn.Notify(
		ctx,
		lsp.MethodTextDocumentPublishDiagnostics,
		&lsp.PublishDiagnosticsParams{
			URI:         doc.URI,
			Diagnostics: doc.DiagnosticsCache.GetMergedDiagnostics(),
		},
	)
}

// GetDiagnostics will run helm linter against the currect document URI using the given values
// and converts the helm.support.Message to lsp.Diagnostics
func GetDiagnostics(uri uri.URI, vals chartutil.Values) ([]lsp.Diagnostic, error) {
	var (
		filename    = uri.Filename()
		paths       = strings.Split(filename, "/")
		dir         = strings.Join(paths, "/")
		diagnostics = make([]lsp.Diagnostic, 0)
	)

	pathfile := ""

	for i, p := range paths {
		if p == "templates" {
			dir = strings.Join(paths[0:i], "/")
			pathfile = strings.Join(paths[i:], "/")
		}
	}

	logger.Println(dir)
	client := action.NewLint()

	result := client.Run([]string{dir}, vals)
	logger.Println(fmt.Sprintf("helm lint: result for file %s : %s", uri, result.Messages))

	for _, msg := range result.Messages {
		d, filename, err := GetDiagnosticFromLinterErr(msg)
		if err != nil {
			continue
		}
		if filename != pathfile {
			continue
		}
		diagnostics = append(diagnostics, *d)
	}

	return diagnostics, nil
}

func GetDiagnosticFromLinterErr(supMsg support.Message) (*lsp.Diagnostic, string, error) {
	var (
		err      error
		msg      string
		line     = 1
		severity lsp.DiagnosticSeverity
		filename = getFilePathFromLinterErr(supMsg)
	)

	switch supMsg.Severity {
	case support.ErrorSev:

		severity = lsp.DiagnosticSeverityError

		if superr, ok := supMsg.Err.(*rules.YAMLToJSONParseError); ok {

			line = superr.Line
			msg = superr.Error()

		} else {

			fileLine := util.BetweenStrings(supMsg.Error(), "(", ")")
			fileLineArr := strings.Split(fileLine, ":")
			if len(fileLineArr) < 2 {
				return nil, filename, errors.Errorf("linter Err contains no position information")
			}
			lineStr := fileLineArr[1]
			line, err = strconv.Atoi(lineStr)
			if err != nil {
				return nil, filename, err
			}
			msgStr := util.AfterStrings(supMsg.Error(), "):")
			msg = strings.TrimSpace(msgStr)

		}

	case support.WarningSev:

		severity = lsp.DiagnosticSeverityWarning
		if err, ok := supMsg.Err.(*rules.MetadataError); ok {
			line = 1
			msg = err.Details().Error()
		}

	case support.InfoSev:

		severity = lsp.DiagnosticSeverityInformation
		msg = supMsg.Err.Error()

	}

	return &lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: uint32(line - 1)},
			End:   lsp.Position{Line: uint32(line - 1)},
		},
		Severity: severity,
		Message:  msg,
	}, filename, nil
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
