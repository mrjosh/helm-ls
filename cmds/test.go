package cmds

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"helm.sh/helm/v3/pkg/action"
)

type LintError struct {
	Line    int
	Message string
	Type    string
}

func newTestCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use: "test PATH",
		RunE: func(cmd *cobra.Command, args []string) error {

			log.SetFlags(log.Lshortfile)
			uri := "/Users/josh/code/work/develop-dashboard/.ci/chart"
			client := action.NewLint()

			//diagnostics := make([]Diagnostic, 0)
			path := uriToPath(string(uri))
			dir, _ := filepath.Split(path)

			//pathfile := ""

			//paths := strings.Split(path, "/")

			//for i, p := range paths {
			//if p == "templates" {
			//dir = strings.Join(paths[0:i], "/")
			//pathfile = strings.Join(paths[i:], "/")
			//}
			//}

			vals := make(map[string]interface{})
			result := client.Run([]string{dir}, vals)

			for _, d := range result.Messages {

				if strings.Contains(d.Error(), "line") {

					//lineNumber, err := findLineNumber(d.Error())
					//if err != nil {
					//return err
					//}

					//log.Println(Diagnostic{
					//Range: Range{
					//Start: Position{Line: lineNumber - 1},
					//End:   Position{Line: lineNumber - 1},
					//},
					//Severity: DSError,
					//Message:  d.Error(),
					//})
				}

			}

			//log.Println("result:", result)

			//for _, err := range result.Errors {
			//d, filename, err := getDiagnosticFromLinterErr(err)
			//if err != nil {
			//continue
			//}
			//if filename != pathfile {
			//continue
			//}
			//diagnostics = append(diagnostics, *d)
			//}

			//for _, d := range diagnostics {
			//log.Println(d, "\n")
			//}

			return nil
		},
	}

	return cmd
}
