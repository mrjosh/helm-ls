package handler

import (
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/uri"
)

func Test_langHandler_getValueHover(t *testing.T) {
	type args struct {
		chart       *charts.Chart
		parentChart *charts.Chart
		splittedVar []string
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
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile: &charts.ValuesFile{
							Values: map[string]interface{}{
								"key": "value",
							},
							URI: "file://tmp/values.yaml"},
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
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile: &charts.ValuesFile{Values: map[string]interface{}{"key": []map[string]interface{}{{"nested": "value"}}}, URI: "file://tmp/values.yaml"},
					},
				},
				splittedVar: []string{"key"},
			},
			want: `### values.yaml
[map[nested:value]]

`,
			wantErr: false,
		},
		{
			name: "subchart includes parent values",
			args: args{
				chart: &charts.Chart{
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile: &charts.ValuesFile{Values: map[string]interface{}{"global": map[string]interface{}{"key": "value"}}, URI: "file://tmp/charts/subchart/values.yaml"},
					},
					ParentChart: charts.ParentChart{
						ParentChartURI: uri.New("file://tmp/"),
						HasParent:      true,
					},
				},
				parentChart: &charts.Chart{
					ValuesFiles: &charts.ValuesFiles{
						MainValuesFile: &charts.ValuesFile{Values: map[string]interface{}{"global": map[string]interface{}{"key": "parentValue"}}, URI: "file://tmp/values.yaml"},
					},
				},
				splittedVar: []string{"global", "key"},
			},
			want: `### values.yaml
parentValue

### charts/subchart/values.yaml
value

`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &langHandler{
				chartStore: &charts.ChartStore{
					RootURI: uri.New("file://tmp/"),
					Charts: map[uri.URI]*charts.Chart{
						uri.New("file://tmp/"): tt.args.parentChart,
					},
				},
			}
			got, err := h.getValueHover(tt.args.chart, tt.args.splittedVar)
			if (err != nil) != tt.wantErr {
				t.Errorf("langHandler.getValueHover() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
