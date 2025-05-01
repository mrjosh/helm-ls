package charts

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/mrjosh/helm-ls/internal/util"
	lsp "go.lsp.dev/protocol"

	"go.lsp.dev/uri"
)

type ValuesFiles struct {
	MainValuesFile        *ValuesFile
	OverlayValuesFile     *ValuesFile
	AdditionalValuesFiles []*ValuesFile
}

func NewValuesFiles(rootURI uri.URI, mainValuesFileName string, lintOverlayValuesFile string, additionalValuesFilesGlob string) *ValuesFiles {
	additionalValuesFiles := getAdditionalValuesFiles(additionalValuesFilesGlob, rootURI, mainValuesFileName)

	overlayValuesFile := getLintOverlayValuesFile(lintOverlayValuesFile, additionalValuesFiles, rootURI)

	return &ValuesFiles{
		MainValuesFile:        NewValuesFileFromPath(filepath.Join(rootURI.Filename(), mainValuesFileName)),
		OverlayValuesFile:     overlayValuesFile,
		AdditionalValuesFiles: additionalValuesFiles,
	}
}

func getLintOverlayValuesFile(lintOverlayValuesFile string, additionalValuesFiles []*ValuesFile, rootURI uri.URI) (overlayValuesFile *ValuesFile) {
	if lintOverlayValuesFile == "" && len(additionalValuesFiles) == 1 {
		overlayValuesFile = additionalValuesFiles[0]
	}
	if lintOverlayValuesFile != "" {
		for _, additionalValuesFile := range additionalValuesFiles {
			if filepath.Base(additionalValuesFile.URI.Filename()) == lintOverlayValuesFile {
				overlayValuesFile = additionalValuesFile
				break
			}
		}
		if overlayValuesFile == nil {
			overlayValuesFile = NewValuesFileFromPath(filepath.Join(rootURI.Filename(), lintOverlayValuesFile))
		}
	}
	return overlayValuesFile
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
				additionalValuesFiles = append(additionalValuesFiles, NewValuesFileFromPath(match))
			}
		}
	}
	return additionalValuesFiles
}

func (v *ValuesFiles) AllValuesFiles() []*ValuesFile {
	// NOTE: we do not include the OverlayValuesFile because it values are not "real"
	return append([]*ValuesFile{v.MainValuesFile}, v.AdditionalValuesFiles...)
}

func (v *ValuesFiles) GetPositionsForValue(query []string) []lsp.Location {
	logger.Debug(fmt.Sprintf("GetPositionsForValue with query %v", query))
	result := []lsp.Location{}
	for _, value := range v.AllValuesFiles() {
		queryCopy := slices.Clone(query)
		pos, err := util.GetPositionOfNode(&value.ValueNode, queryCopy)
		if err != nil {
			yaml, _ := value.Values.YAML()
			logger.Error(fmt.Sprintf("Error getting position for value in yaml file %s with query %v ", yaml, query), err)
			continue
		}
		result = append(result, lsp.Location{URI: value.URI, Range: lsp.Range{Start: pos, End: pos}})
	}

	return result
}
