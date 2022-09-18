/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package action

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/mrjosh/helm-ls/pkg/lint"
	"github.com/mrjosh/helm-ls/pkg/lint/support"
)

// Lint is the action for checking that the semantics of a chart are well-formed.
//
// It provides the implementation of 'helm lint'.
type Lint struct {
	Strict        bool
	Namespace     string
	WithSubcharts bool
	Quiet         bool
}

// LintResult is the result of Lint
type LintResult struct {
	TotalChartsLinted int
	Messages          []support.Message
	Errors            []error
}

// NewLint creates a new Lint object with the given configuration.
func NewLint() *Lint {
	return &Lint{}
}

// Run executes 'helm Lint' against the given chart.
func (l *Lint) Run(paths []string, vals map[string]interface{}) *LintResult {

	lowestTolerance := support.ErrorSev
	if l.Strict {
		lowestTolerance = support.WarningSev
	}

	result := &LintResult{}
	for _, path := range paths {

		linter, err := lintChart(path, vals, l.Namespace, l.Strict)
		if err != nil {
			result.Errors = append(result.Errors, err)
			continue
		}

		result.Messages = append(result.Messages, linter.Messages...)
		result.TotalChartsLinted++
		for _, msg := range linter.Messages {
			if msg.Severity >= lowestTolerance {
				result.Errors = append(result.Errors, msg.Err)
			}
		}

	}

	return result
}

// HasWaringsOrErrors checks is LintResult has any warnings or errors
func HasWarningsOrErrors(result *LintResult) bool {
	for _, msg := range result.Messages {
		if msg.Severity > support.InfoSev {
			return true
		}
	}
	return false
}

func lintChart(
	path string,
	vals map[string]interface{},
	namespace string,
	strict bool,
) (support.Linter, error) {

	linter := support.Linter{}

	// Guard: Error out if this is not a chart.
	if _, err := os.Stat(filepath.Join(path, "Chart.yaml")); err != nil {
		return linter, errors.Wrap(err, "unable to check Chart.yaml file in chart")
	}

	return lint.All(path, vals, namespace, strict), nil
}
