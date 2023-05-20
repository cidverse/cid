package protectoutput

import (
	"io"
	"os"
	"sync"
)

// lastProxyWrite contains the last write of the proxy writer for testing purposes, if no os.File or io.Writer is provided
var lastProxyWrite string

var globalMutex = sync.Mutex{}

type FileProxyWriter struct {
	file   *os.File
	writer io.Writer
	mutex  *sync.Mutex
}

// NewProtectedWriter proxies all output to stdout/stderr to omit/remove any kind of credentials from all logs
func NewProtectedWriter(file *os.File, writer io.Writer) *FileProxyWriter {
	return &FileProxyWriter{
		file:   file,
		writer: writer,
		mutex:  &globalMutex,
	}
}

func (w *FileProxyWriter) Write(p []byte) (int, error) {
	// redact protected phrases in log
	output := RedactProtectedPhrases(string(p))

	// use mutex
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// write data
	if w.file != nil {
		_, err := w.file.WriteString(output)
		if err != nil {
			return 0, err
		}
	} else if w.writer != nil {
		_, err := w.writer.Write([]byte(output))
		if err != nil {
			return 0, err
		}
	} else {
		lastProxyWrite = output
	}

	return len(p), nil
}
