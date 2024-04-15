package languagefeatures

import (
	lsp "go.lsp.dev/protocol"
)

// interface for use cases
type UseCase interface {
	AppropriateForNode() bool
}

type ReferencesUseCase interface {
	UseCase
	References() (result []lsp.Location, err error)
}

type HoverUseCase interface {
	UseCase
	Hover() (result string, err error)
}

type DefinitionUseCase interface {
	UseCase
	Definition() (result []lsp.Location, err error)
}

type CompletionUseCase interface {
	UseCase
	Completion() (result *lsp.CompletionList, err error)
}
