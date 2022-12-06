package quadtreeImage

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
)

type ArchiveWriterMode int

const (
	ArchiveWriterModeTar ArchiveWriterMode = iota
	ArchiveWriterModeZip
)

type ArchiveWriter struct {
	mode      ArchiveWriterMode
	tarWriter *tar.Writer
	zipWriter *zip.Writer
}

// NewArchiveWriter creates a new archive writer for the archive type mode and configures it to write to writer.
func NewArchiveWriter(mode ArchiveWriterMode, writer io.Writer) (*ArchiveWriter, error) {
	archiveWriter := new(ArchiveWriter)

	archiveWriter.mode = mode
	switch mode {
	case ArchiveWriterModeTar:
		archiveWriter.tarWriter = tar.NewWriter(writer)
	case ArchiveWriterModeZip:
		archiveWriter.zipWriter = zip.NewWriter(writer)
	default:
		return nil, fmt.Errorf("no corresponding switch case found for archive mode %d", mode)
	}

	return archiveWriter, nil
}

// CreateFile adds a file to the underlying archive using the provided name and returns a writer to it.
func (w *ArchiveWriter) CreateFile(name string) (io.Writer, error) {
	switch w.mode {
	case ArchiveWriterModeTar:
		return nil, fmt.Errorf("not implemented")
	case ArchiveWriterModeZip:
		return w.zipWriter.Create(name)
	default:
		return nil, fmt.Errorf("no corresponding switch case found for archive mode %d", w.mode)
	}
}

// Close finishes writing the underlying archive.
func (w *ArchiveWriter) Close() error {
	switch w.mode {
	case ArchiveWriterModeTar:
		return fmt.Errorf("not implemented")
	case ArchiveWriterModeZip:
		w.zipWriter.Close()
		return nil
	default:
		return fmt.Errorf("no corresponding switch case found for archive mode %d", w.mode)
	}
}
