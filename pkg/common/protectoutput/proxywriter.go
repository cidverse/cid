package protectoutput

import (
	"io"
	"os"
)

var lastProxyWrite string

type FileProxyWriter struct {
	file *os.File
	writer io.Writer
}

// NewProtectedWriter proxies all output to stdout/stderr to omit/remove any kind of credentials from all logs
func NewProtectedWriter(file *os.File, writer io.Writer) *FileProxyWriter {
	return &FileProxyWriter{
		file: file,
		writer: writer,
	}
}

func (w *FileProxyWriter) Write(p []byte) (int, error) {
	// redact protected phrases in log
	output := RedactProtectedPhrases(string(p))

	// write data
	if w.file != nil {
		w.file.Write([]byte(output))
	} else if w.writer != nil {
		w.writer.Write([]byte(output))
	} else {
		lastProxyWrite = output
	}

	return len(p), nil
}
