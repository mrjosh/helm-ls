package templatehandler

import (
	"strings"

	"go.lsp.dev/protocol"
)

// Takes a string with a mark (^) in it and returns the position and the string without the mark
func getPositionForMarkedTestLine(buf string) (protocol.Position, string) {
	col := strings.Index(buf, "^")
	buf = strings.Replace(buf, "^", "", 1)
	pos := protocol.Position{Line: 0, Character: uint32(col)}
	return pos, buf
}
