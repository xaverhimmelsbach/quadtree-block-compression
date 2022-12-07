package quadtreeImage

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/h2non/filetype"
)

type ArchiveMode string

const (
	ArchiveModeGzip ArchiveMode = "gzip"
	ArchiveModeZip  ArchiveMode = "zip"
)

// ArchiveWriter is an abstraction that allows writing several different compression algorithms.
type ArchiveWriter struct {
	// Compression algorithm in use for this ArchiveWriter.
	mode ArchiveMode
	// Only in use with gzip compression.
	gzipWriter *gzip.Writer
	// Only in use with gzip compression.
	tarWriter *tar.Writer
	// Only in use with zip compression.
	zipWriter *zip.Writer
}

// ArchiveReader is an abstraction that allows reading several different compression algorithms.
type ArchiveReader struct {
	// Compression algorithm in use for this ArchiveReader.
	mode ArchiveMode
	// Only in use with gzip compression.
	gzipReader *gzip.Reader
	// Only in use with gzip compression.
	tarReader *tar.Reader
	// Only in use with zip compression.
	zipReader *zip.ReadCloser
	// Caches all the files contained in the archive.
	fileCache map[string]*[]byte
}

// NewArchiveWriter creates a new archive writer for the archive type mode and configures it to write to writer.
func NewArchiveWriter(mode ArchiveMode, writer io.Writer) (*ArchiveWriter, error) {
	archiveWriter := new(ArchiveWriter)

	archiveWriter.mode = mode
	switch mode {
	case ArchiveModeGzip:
		archiveWriter.gzipWriter = gzip.NewWriter(writer)
		// Chain gzipWriter with tarWriter
		archiveWriter.tarWriter = tar.NewWriter(archiveWriter.gzipWriter)
	case ArchiveModeZip:
		archiveWriter.zipWriter = zip.NewWriter(writer)
	default:
		return nil, fmt.Errorf("no corresponding switch case found for archive mode %s", mode)
	}

	return archiveWriter, nil
}

// WriteFile adds a file to the underlying archive and writes the contents of reader to it.
func (w *ArchiveWriter) WriteFile(name string, reader io.Reader) error {
	switch w.mode {
	case ArchiveModeGzip:
		fileContents, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}

		// Create bare-bones file header
		header := new(tar.Header)
		// Always assume a regular file (for now)
		header.Typeflag = tar.TypeReg
		header.Name = name
		header.Size = int64(len(fileContents))
		header.Mode = 544

		err = w.tarWriter.WriteHeader(header)
		if err != nil {
			return err
		}

		_, err = w.tarWriter.Write(fileContents)
		if err != nil {
			return err
		}

		return nil
	case ArchiveModeZip:
		writer, err := w.zipWriter.Create(name)
		if err != nil {
			return err
		}

		fileContents, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		writer.Write(fileContents)
	default:
		return fmt.Errorf("no corresponding switch case found for archive mode %s", w.mode)
	}

	return nil
}

// Close flushes the underlying archive to its writer.
func (w *ArchiveWriter) Close() error {
	switch w.mode {
	case ArchiveModeGzip:
		w.tarWriter.Close()
		w.gzipWriter.Close()
	case ArchiveModeZip:
		w.zipWriter.Close()
	default:
		return fmt.Errorf("no corresponding switch case found for archive mode %s", w.mode)
	}
	return nil
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

	switch ArchiveMode(filetype.MIME.Subtype) {
	case ArchiveModeGzip:
		archiveReader.mode = ArchiveModeGzip
		// Load archiveContents into buffer to reread data from the beginning
		archiveBuffer := bytes.NewBuffer(archiveContents)
		archiveReader.gzipReader, err = gzip.NewReader(archiveBuffer)
		if err != nil && !errors.Is(err, io.EOF) {
			return archiveReader, err
		}

		// Create tar reader and cache archive files
		archiveReader.tarReader = tar.NewReader(archiveReader.gzipReader)
		archiveReader.populateFileCacheGzip()
	case ArchiveModeZip:
		archiveReader.mode = ArchiveModeZip
		// TODO: Unneccessary read
		archiveReader.zipReader, err = zip.OpenReader(name)
		if err != nil {
			return archiveReader, err
		}

		// Cache archive files
		archiveReader.populateFileCacheZip()
	default:
		return archiveReader, fmt.Errorf("no corresponding switch case found for archive type %s", filetype.MIME.Subtype)
	}

	return archiveReader, nil
}

// Open opens the named file in the archive and returns a reader to it.
func (r *ArchiveReader) Open(name string) (*[]byte, error) {
	fileContents, ok := r.fileCache[name]
	if !ok {
		// TODO: Is it ok to return a fs error here?
		return nil, fs.ErrNotExist
	}

	return fileContents, nil
}

// File returns the list of files contained in the archive.
func (r *ArchiveReader) Files() map[string]*[]byte {
	return r.fileCache
}

// populateFileCacheGzip populates the fileCache with all files contained in a tar.gz archive.
func (r *ArchiveReader) populateFileCacheGzip() error {
	// Init cache
	r.fileCache = make(map[string]*[]byte)

	for {
		header, err := r.tarReader.Next()

		if errors.Is(err, io.EOF) {
			// Last file read
			break
		} else if err != nil {
			return err
		}

		filename := header.Name

		// Read file contents
		fileContents, err := ioutil.ReadAll(r.tarReader)
		if err != nil {
			return err
		}

		// Add to cache
		r.fileCache[filename] = &fileContents
	}

	return nil
}

// populateFileCacheZip populates the fileCache with all files contained in a zip archive.
func (r *ArchiveReader) populateFileCacheZip() error {
	// Init cache
	r.fileCache = make(map[string]*[]byte)

	for _, file := range r.zipReader.File {
		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		// Read file contents
		fileContents, err := ioutil.ReadAll(fileReader)
		if err != nil {
			return err
		}

		r.fileCache[file.Name] = &fileContents
	}

	return nil
}
