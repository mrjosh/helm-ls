package languagefeatures

import (
	"testing"

	"github.com/mrjosh/helm-ls/internal/charts"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
)

func TestGetValuesCompletions(t *testing.T) {
	nested := map[string]interface{}{"nested": "value"}
	valuesMain := map[string]interface{}{"global": nested}
	valuesAdditional := map[string]interface{}{"glob": nested}
	chart := &charts.Chart{
		ChartMetadata: &charts.ChartMetadata{Metadata: chart.Metadata{Name: "test"}},
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile: &charts.ValuesFile{
				Values:    valuesMain,
				ValueNode: yaml.Node{},
				URI:       "",
			},
			AdditionalValuesFiles: []*charts.ValuesFile{
				{
					Values:    valuesAdditional,
					ValueNode: yaml.Node{},
					URI:       "",
				},
			},
		},
		RootURI: "", HelmChart: &chart.Chart{},
	}

	templateConextFeature := TemplateContextFeature{
		GenericTemplateContextFeature: &GenericTemplateContextFeature{
			GenericDocumentUseCase: &GenericDocumentUseCase{
				Chart: chart,
			},
		},
	}

	result, err := templateConextFeature.valuesCompletion([]string{"Values", "g"})
	assert.NoError(t, err)
	assert.Len(t, result.Items, 2)

	result, err = templateConextFeature.valuesCompletion([]string{"Values", "something", "different"})
	assert.NoError(t, err)
	assert.Len(t, result.Items, 0)
}

func TestGetValuesCompletionsContainsNoDupliactes(t *testing.T) {
	nested := map[string]interface{}{"nested": "value"}
	valuesMain := map[string]interface{}{"global": nested}
	valuesAdditional := map[string]interface{}{"global": nested}
	chart := &charts.Chart{
		ChartMetadata: &charts.ChartMetadata{Metadata: chart.Metadata{Name: "test"}},
		ValuesFiles: &charts.ValuesFiles{
			MainValuesFile: &charts.ValuesFile{
				Values:    valuesMain,
				ValueNode: yaml.Node{},
				URI:       "",
			},
			AdditionalValuesFiles: []*charts.ValuesFile{
				{
					Values: valuesAdditional,
					URI:    "",
				},
			},
		},
		RootURI: "", HelmChart: &chart.Chart{},
	}

	templateConextFeature := TemplateContextFeature{
		GenericTemplateContextFeature: &GenericTemplateContextFeature{
			GenericDocumentUseCase: &GenericDocumentUseCase{
				Chart: chart,
			},
		},
	}

	result, err := templateConextFeature.valuesCompletion([]string{"Values", "g"})
	assert.NoError(t, err)
	assert.Len(t, result.Items, 1)
}
