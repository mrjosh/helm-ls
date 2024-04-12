package languagefeatures

import (
	"path/filepath"
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
	"helm.sh/helm/v3/pkg/chart"
)

func Test_langHandler_getValueHover(t *testing.T) {
	type args struct {
		chart        *charts.Chart
		parentCharts map[uri.URI]*charts.Chart
		splittedVar  []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "single values file",
			args: args{
				chart: &charts.Chart{
					ChartMetadata: &charts.ChartMetadata{},
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile: &charts.ValuesFile{
							Values: map[string]interface{}{
								"key": "value",
							},
							URI: "file://tmp/values.yaml",
						},
					},
				},
				splittedVar: []string{"key"},
			},
			want: `### values.yaml
value

`,
			wantErr: false,
		},
		{
			name: "multiple values files",
			args: args{
				chart: &charts.Chart{
					ChartMetadata: &charts.ChartMetadata{},
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile:        &charts.ValuesFile{Values: map[string]interface{}{"key": "value"}, URI: "file://tmp/values.yaml"},
						AdditionalValuesFiles: []*charts.ValuesFile{{Values: map[string]interface{}{"key": ""}, URI: "file://tmp/values.other.yaml"}},
					},
				},
				splittedVar: []string{"key"},
			},
			want: `### values.yaml
value

### values.other.yaml
""

`,
			wantErr: false,
		},
		{
			name: "yaml result",
			args: args{
				chart: &charts.Chart{
					ChartMetadata: &charts.ChartMetadata{},
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile: &charts.ValuesFile{Values: map[string]interface{}{"key": map[string]interface{}{"nested": "value"}}, URI: "file://tmp/values.yaml"},
					},
				},
				splittedVar: []string{"key"},
			},
			want: `### values.yaml
nested: value


`,
			wantErr: false,
		},
		{
			name: "yaml result as list",
			args: args{
				chart: &charts.Chart{
					ChartMetadata: &charts.ChartMetadata{},
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile: &charts.ValuesFile{Values: map[string]interface{}{"key": []map[string]interface{}{{"nested": "value"}}}, URI: "file://tmp/values.yaml"},
					},
				},
				splittedVar: []string{"key"},
			},
			want: `### values.yaml
key:
- nested: value


`,
			wantErr: false,
		},
		{
			name: "subchart includes parent values global",
			args: args{
				chart: &charts.Chart{
					ChartMetadata: &charts.ChartMetadata{},
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile: &charts.ValuesFile{Values: map[string]interface{}{"global": map[string]interface{}{"key": "value"}}, URI: "file://tmp/charts/subchart/values.yaml"},
					},
					ParentChart: charts.ParentChart{
						ParentChartURI: uri.New("file://tmp/"),
						HasParent:      true,
					},
				},
				parentCharts: map[uri.URI]*charts.Chart{
					uri.New("file://tmp/"): {
						ChartMetadata: &charts.ChartMetadata{},
						ValuesFiles: &charts.ValuesFiles{
							MainValuesFile: &charts.ValuesFile{Values: map[string]interface{}{"global": map[string]interface{}{"key": "parentValue"}}, URI: "file://tmp/values.yaml"},
						},
					},
				},
				splittedVar: []string{"global", "key"},
			},
			want: `### values.yaml
parentValue

### ` + filepath.Join("charts", "subchart", "values.yaml") + `
value

`,
			wantErr: false,
		},
		{
			name: "subchart includes parent values by chart name",
			args: args{
				chart: &charts.Chart{
					ChartMetadata: &charts.ChartMetadata{
						Metadata: chart.Metadata{Name: "subchart"},
					},
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile: &charts.ValuesFile{Values: map[string]interface{}{"key": "value"}, URI: "file://tmp/charts/subchart/values.yaml"},
					},
					ParentChart: charts.ParentChart{
						ParentChartURI: uri.New("file://tmp/"),
						HasParent:      true,
					},
				},
				parentCharts: map[uri.URI]*charts.Chart{
					uri.New("file://tmp/"): {
						ChartMetadata: &charts.ChartMetadata{},
						ValuesFiles: &charts.ValuesFiles{
							MainValuesFile: &charts.ValuesFile{Values: map[string]interface{}{"subchart": map[string]interface{}{"key": "parentValue"}}, URI: "file://tmp/values.yaml"},
						},
					},
				},
				splittedVar: []string{"key"},
			},
			want: `### values.yaml
parentValue

### ` + filepath.Join("charts", "subchart", "values.yaml") + `
value

`,
			wantErr: false,
		},
		{
			name: "subsubchart includes parent values by chart name",
			args: args{
				chart: &charts.Chart{
					ChartMetadata: &charts.ChartMetadata{Metadata: chart.Metadata{Name: "subsubchart"}},
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile: &charts.ValuesFile{Values: map[string]interface{}{"key": "value"}, URI: "file://tmp/charts/subchart/charts/subsubchart/values.yaml"},
					},
					ParentChart: charts.ParentChart{
						ParentChartURI: uri.New("file://tmp/charts/subchart"),
						HasParent:      true,
					},
				},
				parentCharts: map[uri.URI]*charts.Chart{
					uri.New("file://tmp/charts/subchart"): {
						ChartMetadata: &charts.ChartMetadata{Metadata: chart.Metadata{Name: "subchart"}},
						ValuesFiles: &charts.ValuesFiles{
							MainValuesFile: &charts.ValuesFile{Values: map[string]interface{}{"subsubchart": map[string]interface{}{"key": "middleValue"}}, URI: "file://tmp/charts/subchart/values.yaml"},
						},
						ParentChart: charts.ParentChart{
							ParentChartURI: uri.New("file://tmp/"),
							HasParent:      true,
						},
					},
					uri.New("file://tmp/"): {
						ChartMetadata: &charts.ChartMetadata{
							Metadata: chart.Metadata{Name: "parent"},
						},
						ValuesFiles: &charts.ValuesFiles{
							MainValuesFile: &charts.ValuesFile{Values: map[string]interface{}{"subchart": map[string]interface{}{"subsubchart": map[string]interface{}{"key": "parentValue"}}}, URI: "file://tmp/values.yaml"},
						},
					},
				},
				splittedVar: []string{"key"},
			},
			want: `### values.yaml
parentValue

### ` + filepath.Join("charts", "subchart", "values.yaml") + `
middleValue

### ` + filepath.Join("charts", "subchart", "charts", "subsubchart", "values.yaml") + `
value

`,
			wantErr: false,
		},
		{
			name: "Formatting of number",
			args: args{
				chart: &charts.Chart{
					ChartMetadata: &charts.ChartMetadata{},
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile: &charts.ValuesFile{
							Values: map[string]interface{}{
								"key": float64(1.2345),
							},
							URI: "file://tmp/values.yaml",
						},
					},
				},
				splittedVar: []string{"key"},
			},
			want: `### values.yaml
1.2345

`,
			wantErr: false,
		},
		{
			name: "Lookup in list",
			args: args{
				chart: &charts.Chart{
					ChartMetadata: &charts.ChartMetadata{},
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile: &charts.ValuesFile{
							Values: map[string]interface{}{
								"key": []interface{}{"hello"},
							},
							URI: "file://tmp/values.yaml",
						},
					},
				},
				splittedVar: []string{"key[0]"},
			},
			want: `### values.yaml
hello

`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genericDocumentUseCase := &GenericDocumentUseCase{
				Chart: tt.args.chart,
				ChartStore: &charts.ChartStore{
					RootURI: uri.New("file://tmp/"),
					Charts:  tt.args.parentCharts,
				},
				// Node: tt.args.chart.ValuesFiles.MainValuesFile.Node,
			}
			valuesFeature := NewTemplateContextFeature(genericDocumentUseCase)
			got, err := valuesFeature.valuesHover(tt.args.splittedVar)
			if (err != nil) != tt.wantErr {
				t.Errorf("langHandler.getValueHover() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
