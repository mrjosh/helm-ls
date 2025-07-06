package testutil

import (
	"strings"

	"go.lsp.dev/protocol"
)

// Finds the first occurrence of the marked line (or a substring) in the file
// and returns its position.
func GetPositionOfMarkedLineInFile(fileContent, markedLine, marker string) (pos protocol.Position, found bool) {
	lines := strings.Split(fileContent, "\n")
	col := strings.Index(markedLine, marker)
	buf := strings.Replace(markedLine, marker, "", 1)
	line := uint32(0)

	for i, v := range lines {
		if strings.Contains(v, buf) {
			found = true
			line = uint32(i)
			col = col + strings.Index(v, buf)
			break
		}
	}
	pos = protocol.Position{Line: line, Character: uint32(col)}
	return pos, found
}

func GetRangeOfMarkedLineInFile(fileContent, markedLine, marker string) (pos protocol.Range, found bool) {
	lineWithFirstMarker, lineWithSecondMarker := RangeMarkedLineToPositionMarkedLines(markedLine, marker)
	start, found := GetPositionOfMarkedLineInFile(fileContent, lineWithFirstMarker, marker)
	if !found {
		return pos, false
	}
	end, found := GetPositionOfMarkedLineInFile(fileContent, lineWithSecondMarker, marker)

	return protocol.Range{Start: start, End: end}, found
}

func RangeMarkedLineToPositionMarkedLines(markedLine string, marker string) (string, string) {
	lastIndex := strings.LastIndex(markedLine, marker)
	lineWithFirstMarker := markedLine[:lastIndex] + markedLine[lastIndex+len(marker):]
	lineWithSecondMarker := strings.Replace(markedLine, marker, "", 1)

	return lineWithFirstMarker, lineWithSecondMarker
}
