package charts

import (
	"path/filepath"

	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"

	"go.lsp.dev/uri"
)

type ValuesFiles struct {
	MainValuesFile        *ValuesFile
	AdditionalValuesFiles []*ValuesFile
}

func NewValuesFiles(rootURI uri.URI, mainValuesFileName string, additionalValuesFilesGlob string) *ValuesFiles {
	additionalValuesFiles := getAdditionalValuesFiles(additionalValuesFilesGlob, rootURI, mainValuesFileName)

	return &ValuesFiles{
		MainValuesFile:        NewValuesFile(filepath.Join(rootURI.Filename(), mainValuesFileName)),
		AdditionalValuesFiles: additionalValuesFiles,
	}
}

func getAdditionalValuesFiles(additionalValuesFilesGlob string, rootURI uri.URI, mainValuesFileName string) []*ValuesFile {
	additionalValuesFiles := []*ValuesFile{}
	if additionalValuesFilesGlob != "" {

		matches, err := filepath.Glob(filepath.Join(rootURI.Filename(), additionalValuesFilesGlob))
		if err != nil {
			logger.Error("Error loading additional values files with glob pattern", additionalValuesFilesGlob, err)
		} else {

			for _, match := range matches {
				if match == filepath.Join(rootURI.Filename(), mainValuesFileName) {
					continue
				}
				additionalValuesFiles = append(additionalValuesFiles, NewValuesFile(match))
			}
		}
	}
	return additionalValuesFiles
}

func (v *ValuesFiles) AllValuesFiles() []*ValuesFile {
	return append([]*ValuesFile{v.MainValuesFile}, v.AdditionalValuesFiles...)
}

func (v *ValuesFiles) GetPositionsForValue(query []string) []lsp.Location {
	var result = []lsp.Location{}
	for _, value := range v.AllValuesFiles() {
		pos, err := util.GetPositionOfNode(&value.ValueNode, query)
		if err != nil {
			logger.Error("Error getting position for value", value, query, err)
			continue
		}
		result = append(result, lsp.Location{URI: value.URI, Range: lsp.Range{Start: pos, End: pos}})
	}

	return result
}
