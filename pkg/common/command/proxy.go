package command

import "os"

type FileProxyWriter struct {
	file *os.File
}

// NewFileProxyWriter proxies all output to stdout/stderr to omit/remove any kind of credentials from all logs
func NewFileProxyWriter(file *os.File) *FileProxyWriter {
	return &FileProxyWriter{
		file: file,
	}
}

func (w *FileProxyWriter) Write(p []byte) (int, error) {
	// TODO: support to filter output / secrets from log

	// write data
	w.file.Write(p)

	return len(p), nil
}
