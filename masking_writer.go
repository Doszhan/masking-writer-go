package maskingwriter

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"sync"
)

// Writer implements writer, that will proxy to specified `backend` writer only
// complete lines, e.g. that ends in newline. This writer is thread-safe.
type Writer struct {
	lock    sync.Locker
	backend io.WriteCloser
	buffer  *bytes.Buffer
	masking []string

	ensureNewline bool
}

// New returns new Writer, that will proxy data to the `backend` writer,
// thread-safety is guaranteed via `lock`. Optionally, writer can ensure, that
// last line of output ends with newline, if `ensureNewline` is true.
func New(
	writer io.WriteCloser,
	lock sync.Locker,
	ensureNewline bool,
	masking []string,
) *Writer {
	return &Writer{
		backend: writer,
		lock:    lock,
		buffer:  &bytes.Buffer{},
		masking: masking,

		ensureNewline: ensureNewline,
	}
}

// Writer writes data into Writer.
//
// Signature matches with io.Writer's Write().
func (writer *Writer) Write(data []byte) (int, error) {
	writer.lock.Lock()
	written, err := writer.buffer.Write(data)
	writer.lock.Unlock()
	if err != nil {
		return written, err
	}

	var reader = bufio.NewReader(writer.buffer)
	writer.lock.Lock()
	line, err := reader.ReadString('\n')
	line = strings.ReplaceAll(line, writer.masking[0], "***")
	_, err = io.WriteString(writer.backend, line)
	writer.lock.Unlock()

	if err != nil {
		return 0, err
	}

	return written, nil
}

// Close flushes all remaining data and closes underlying backend writer.
// If `ensureNewLine` was specified and remaining data does not ends with
// newline, then newline will be added.
//
// Signature matches with io.WriteCloser's Close().
func (writer *Writer) Close() error {
	return writer.backend.Close()
}
