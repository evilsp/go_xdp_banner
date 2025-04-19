package log

import (
	"fmt"
	"io"
	"os"
)

// WriterOpener is a module that can open a log writer.
// It can return a human-readable string representation
// of itself so that operators can understand where
// the logs are going.
type WriterOpener interface {
	fmt.Stringer

	// WriterKey is a string that uniquely identifies this
	// writer configuration. It is not shown to humans.
	WriterKey() string

	// OpenWriter opens a log for writing. The writer
	// should be safe for concurrent use but need not
	// be synchronous.
	OpenWriter() (io.WriteCloser, error)
}

// IsWriterStandardStream returns true if the input is a
// writer-opener to a standard stream (stdout, stderr).
func IsWriterStandardStream(wo WriterOpener) bool {
	switch wo.(type) {
	case StdoutWriter, StderrWriter,
		*StdoutWriter, *StderrWriter:
		return true
	}
	return false
}

// notClosable is an io.WriteCloser that can't be closed.
type notClosable struct{ io.Writer }

func (fc notClosable) Close() error { return nil }

type (
	// StdoutWriter writes logs to standard out.
	StdoutWriter struct{}

	// StderrWriter writes logs to standard error.
	StderrWriter struct{}

	// DiscardWriter discards all writes.
	DiscardWriter struct{}
)

func (StdoutWriter) String() string  { return "stdout" }
func (StderrWriter) String() string  { return "stderr" }
func (DiscardWriter) String() string { return "discard" }

// WriterKey returns a unique key representing stdout.
func (StdoutWriter) WriterKey() string { return "std:out" }

// WriterKey returns a unique key representing stderr.
func (StderrWriter) WriterKey() string { return "std:err" }

// WriterKey returns a unique key representing discard.
func (DiscardWriter) WriterKey() string { return "discard" }

// OpenWriter returns os.Stdout that can't be closed.
func (StdoutWriter) OpenWriter() (io.WriteCloser, error) {
	return notClosable{os.Stdout}, nil
}

// OpenWriter returns os.Stderr that can't be closed.
func (StderrWriter) OpenWriter() (io.WriteCloser, error) {
	return notClosable{os.Stderr}, nil
}

// OpenWriter returns io.Discard that can't be closed.
func (DiscardWriter) OpenWriter() (io.WriteCloser, error) {
	return notClosable{io.Discard}, nil
}
