package yamlls

import (
	"io"
	"log"
	"os"
)

type readWriteCloseSubprocess struct {
	stdout io.ReadCloser
	stdin  io.WriteCloser
}

func (proc readWriteCloseSubprocess) Read(p []byte) (int, error) {

	l := log.New(os.Stderr, "", 1)
	l.Println("read called")
	// l.Println(string(p))
	return proc.stdout.Read(p)
}

func (proc readWriteCloseSubprocess) Write(p []byte) (int, error) {
	return proc.stdin.Write(p)
}

func (proc readWriteCloseSubprocess) Close() error {
	return proc.stdin.Close()
}
