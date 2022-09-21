package lsp

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mrjosh/helm-ls/internal/log"
	"github.com/mrjosh/helm-ls/internal/util"
	"github.com/mrjosh/helm-ls/pkg/action"
	"github.com/mrjosh/helm-ls/pkg/chartutil"
	"github.com/mrjosh/helm-ls/pkg/lint/support"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"

	// nolint
	"github.com/mrjosh/helm-ls/pkg/lint/rules"
	"go.lsp.dev/jsonrpc2"
	lsp "go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

var logger = log.GetLogger()

func NotifcationFromLint(ctx context.Context, conn jsonrpc2.Conn, uri uri.URI) (*jsonrpc2.Notification, error) {
	diagnostics, err := GetDiagnostics(uri)
	if err != nil {
		return nil, err
	}
	return nil, conn.Notify(
		ctx,
		lsp.MethodTextDocumentPublishDiagnostics,
		&lsp.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diagnostics,
		},
	)
}

// loadValues will load the values files into a map[string]interface{}
// the filename arg default is values.yaml
func loadValues(dir string, filename ...string) (map[string]interface{}, error) {

	vals := make(map[string]interface{})
	if len(filename) == 0 {
		filename = append(filename, chartutil.ValuesfileName)
	}

	if len(filename) > 1 {
		return vals, errors.New("filename should be a single string")
	}

	file, err := os.Open(fmt.Sprintf("%s/%s", dir, filename[0]))
	if err != nil {
		return vals, err
	}

	if err := yaml.NewDecoder(file).Decode(&vals); err != nil {
		return vals, err
	}

	logger.Println(fmt.Sprintf("%s file loaded successfully", file.Name()))
	logger.Debug(vals)

	return vals, nil
}

// GetDiagnostics will run helm linter agains the currect document URI
// and converts the helm.support.Message to lsp.Diagnostics
func GetDiagnostics(uri uri.URI) ([]lsp.Diagnostic, error) {

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

	vals, err := loadValues(dir)
	if err != nil {

		logger.Println(errors.Wrap(err, "could not load values.yaml, trying to load values.yml instead"))

		vals, err = loadValues(dir, "values.yml")
		if err != nil {
			logger.Println(errors.Wrap(err, "could not load values.yml, ignoring values"))
		}

	}

	result := client.Run([]string{dir}, vals)
	logger.Println("helm lint: result:", result.Messages)

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
		line     int = 1
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
			lineStr := fileLineArr[1]
			msgStr := util.AfterStrings(supMsg.Error(), "):")
			msg = strings.TrimSpace(msgStr)

			line, err = strconv.Atoi(lineStr)
			if err != nil {
				return nil, filename, err
			}

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
