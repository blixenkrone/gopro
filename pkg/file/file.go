package file

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/pkg/errors"

	"github.com/blixenkrone/gopro/pkg/conversion"
)

type File struct {
	file *os.File
}

var mut = sync.Mutex{}

type FileGenerator interface {
	File() *os.File
	Close() error
	RemoveFile() error
	FileName() string
	FileSize() (size float64, err error)
	FileStat() (os.FileInfo, error)
}

func NewFileLtdRead(r io.Reader, limit int64) (FileGenerator, error) {
	// rd := io.LimitReader(r, limit)
	rd := io.LimitReader(r, 1000)
	b, err := ioutil.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	return writeTmpFile(b)

}

// Read whole file at once
func NewFile(r io.Reader) (FileGenerator, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return writeTmpFile(b)
}

// Read file buffered as scanner ! not tested !
func NewFileBuffer(r *bufio.Scanner) (FileGenerator, error) {
	var b []byte
	for r.Scan() {
		if err := r.Err(); err != nil {
			return nil, err
		}
	}
	return writeTmpFile(b)
}

func writeTmpFile(b []byte) (FileGenerator, error) {
	file, err := ioutil.TempFile(os.TempDir(), "prefix-*")
	if err != nil {
		return nil, errors.Wrap(err, "error creating tmp file")
	}

	mut.Lock()
	if _, err = file.Write(b); err != nil {
		return nil, errors.Wrap(err, "error writing to tmp file")
	}
	mut.Unlock()
	if err := file.Chmod(0777); err != nil {
		return nil, errors.Wrap(err, "error chmod tmp file")
	}
	if err := file.Sync(); err != nil {
		return nil, errors.Wrap(err, "error sync tmp file")
	}
	return &File{file}, nil
}

func (f *File) File() *os.File {
	return f.file
}

func (f *File) Close() error {
	return f.file.Close()
}

func (f *File) RemoveFile() error {
	return os.Remove(f.file.Name())
}

func (f *File) FileName() string {
	return f.file.Name()
}

func (f *File) FileStat() (os.FileInfo, error) {
	return f.file.Stat()
}

func (f *File) FileSize() (size float64, err error) {
	fInfo, err := f.file.Stat()
	if err != nil {
		return size, err
	}
	size = conversion.FileSizeBytesToFloat(int(fInfo.Size()))
	return size, err
}

func (f *File) EncodeExif(metaTag, value string) error {
	// Handle file types

	return nil

}
