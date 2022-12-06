package quadtreeImage

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/h2non/filetype"
)

type ArchiveMode int

const (
	ArchiveModeTar ArchiveMode = iota
	ArchiveModeZip
)

type ArchiveWriter struct {
	mode      ArchiveMode
	tarWriter *tar.Writer
	zipWriter *zip.Writer
}

type ArchiveReader struct {
	mode      ArchiveMode
	tarReader *tar.Reader
	zipReader *zip.ReadCloser
}

// NewArchiveWriter creates a new archive writer for the archive type mode and configures it to write to writer.
func NewArchiveWriter(mode ArchiveMode, writer io.Writer) (*ArchiveWriter, error) {
	archiveWriter := new(ArchiveWriter)

	archiveWriter.mode = mode
	switch mode {
	case ArchiveModeTar:
		archiveWriter.tarWriter = tar.NewWriter(writer)
	case ArchiveModeZip:
		archiveWriter.zipWriter = zip.NewWriter(writer)
	default:
		return nil, fmt.Errorf("no corresponding switch case found for archive mode %d", mode)
	}

	return archiveWriter, nil
}

// CreateFile adds a file to the underlying archive using the provided name and returns a writer to it.
func (w *ArchiveWriter) CreateFile(name string) (io.Writer, error) {
	switch w.mode {
	case ArchiveModeTar:
		return nil, fmt.Errorf("not implemented")
	case ArchiveModeZip:
		return w.zipWriter.Create(name)
	default:
		return nil, fmt.Errorf("no corresponding switch case found for archive mode %d", w.mode)
	}
}

// Close finishes writing the underlying archive.
func (w *ArchiveWriter) Close() error {
	switch w.mode {
	case ArchiveModeTar:
		return fmt.Errorf("not implemented")
	case ArchiveModeZip:
		w.zipWriter.Close()
		return nil
	default:
		return fmt.Errorf("no corresponding switch case found for archive mode %d", w.mode)
	}
}

// OpenArchiveReader will open the archive file specified by name and return a reader.
func OpenArchiveReader(name string) (*ArchiveReader, error) {
	// Open archive to infer filetype
	archiveFile, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer archiveFile.Close()

	archiveContents, err := ioutil.ReadAll(archiveFile)
	if err != nil {
		return nil, err
	}

	filetype, err := filetype.Match(archiveContents)
	if err != nil {
		return nil, err
	}

	archiveReader := new(ArchiveReader)

	switch filetype.MIME.Subtype {
	case "tar":
		archiveReader.mode = ArchiveModeTar
		archiveReader.tarReader = tar.NewReader(archiveFile)
		return archiveReader, fmt.Errorf("not implemented")
	case "zip":
		archiveReader.mode = ArchiveModeZip
		archiveReader.zipReader, err = zip.OpenReader(name)
		if err != nil {
			return archiveReader, err
		}
	default:
		return archiveReader, fmt.Errorf("no corresponding switch case found for archive type %s", filetype.MIME.Subtype)
	}

	return archiveReader, nil
}

// Open opens the named file in the archive.
func (r *ArchiveReader) Open(name string) (fs.File, error) {
	switch r.mode {
	case ArchiveModeTar:
		return nil, fmt.Errorf("not implemented")
	case ArchiveModeZip:
		return r.zipReader.Open(name)
	default:
		return nil, fmt.Errorf("no corresponding switch case found for archive mode %d", r.mode)
	}
}

// File returns the list of files contained in the archive.
func (r *ArchiveReader) File() ([]*zip.File, error) {
	switch r.mode {
	case ArchiveModeTar:
		return nil, fmt.Errorf("not implemented")
	case ArchiveModeZip:
		return r.zipReader.File, nil
	default:
		return nil, fmt.Errorf("no corresponding switch case found for archive mode %d", r.mode)
	}
}
