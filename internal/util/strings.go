package util

import (
	"regexp"
	"strings"

	"github.com/mrjosh/helm-ls/internal/log"
	"go.lsp.dev/protocol"
)

var (
	logger    = log.GetLogger()
	wordRegex = regexp.MustCompile(`[^ \t\n\f\r,;\[\]\"\']+`)
)

// BetweenStrings gets the substring between two strings.
func BetweenStrings(value string, a string, b string) string {
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

// AfterStrings gets the substring after a string.
func AfterStrings(value string, a string) string {
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:]
}

// WordAt returns the word found at the given character position.
// Credit https://github.com/aca/neuron-language-server/blob/450a7cff71c14e291ee85ff8a0614fa9d4dd5145/utils.go#L13
func WordAt(str string, index int) string {
	wordIdxs := wordRegex.FindAllStringIndex(str, -1)
	for _, wordIdx := range wordIdxs {
		if wordIdx[0] <= index && index <= wordIdx[1] {
			return str[wordIdx[0]:wordIdx[1]]
		}
	}

	return ""
}

func PositionToIndex(pos protocol.Position, content []byte) int {
	index := 0
	for i := 0; i < int(pos.Line); i++ {
		if i < int(pos.Line) {
			index = index + strings.Index(string(content[index:]), "\n") + 1
		}
	}

	index = index + int(pos.Character)
	return index
}

func IndexToPosition(index int, content []byte) protocol.Position {
	line := 0
	char := 0
	for i := 0; i < index-1; i++ {
		if string(content[i]) == "\n" {
			line++
			char = 0
		} else {
			char++
		}
	}

	return protocol.Position{
		Line:      uint32(line),
		Character: uint32(char),
	}
}
