package yamlls

import (
	"io"
)

type readWriteCloseSubprocess struct {
	stdout io.ReadCloser
	stdin  io.WriteCloser
}

func (proc readWriteCloseSubprocess) Read(p []byte) (int, error) {
	return proc.stdout.Read(p)
}

func (proc readWriteCloseSubprocess) Write(p []byte) (int, error) {
	return proc.stdin.Write(p)
}

func (proc readWriteCloseSubprocess) Close() error {
	return proc.stdin.Close()
}
