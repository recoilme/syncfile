// Package syncfile implements concurrent write/read to/from file
package syncfile

import (
	"bytes"
	"errors"
	"os"
	"sync"
)

// SyncFile helps to append lines to a file.
//
// It is safe for concurrent usage.
type SyncFile struct {
	f  *os.File
	mu *sync.RWMutex
}

// NewSyncFile returns a new SyncFile.
//
// It immediatly opens (and creates) the file.
func NewSyncFile(name string, perm os.FileMode) (*SyncFile, error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, perm)
	if err != nil {
		return nil, err
	}
	return &SyncFile{
		f:  f,
		mu: new(sync.RWMutex),
	}, nil
}

// Append appends the given byte array to the file.
func (sf *SyncFile) Append(b []byte) (seek int64, n int, err error) {
	return sf.Write(b)
}

// Write writes len(b) bytes to the File.
// It returns the number of bytes written and an error, if any.
// Write returns a non-nil error when n != len(b).
func (sf *SyncFile) Write(b []byte) (seek int64, n int, err error) {
	buf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(buf)
	buf.Reset()
	n, err = buf.Write(b)
	if err != nil {
		return 0, n, err
	}
	return sf.write(buf.Bytes())
}

// ReadFile reads the file named by filename and returns the contents.
// A successful call returns err == nil, not err == EOF. Because ReadFile
// reads the whole file, it does not treat an EOF from Read as an error
// to be reported.
func (sf *SyncFile) ReadFile() ([]byte, error) {
	fi, err := sf.f.Stat()
	if err != nil || fi.Size() == 0 {
		return nil, err
	}
	return sf.Read(fi.Size(), 0)
}

// Read read size bytes from seek
func (sf *SyncFile) Read(size int64, seek int64) ([]byte, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(buf)
	buf.Reset()
	return sf.read(buf, size, seek)
}

func (sf *SyncFile) read(buf *bytes.Buffer, size int64, seek int64) ([]byte, error) {
	sf.mu.RLock() //with RLock -  not working
	defer sf.mu.RUnlock()
	var err error
	if err != nil {
		return nil, err
	}
	byteSlice := make([]byte, size)
	buf.Grow(int(size))
	//TODO read by chanks
	_, err = sf.f.ReadAt(byteSlice, seek) // .Read(byteSlice)

	buf.Write(byteSlice)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (sf *SyncFile) write(b []byte) (seek int64, n int, err error) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	seek, err = sf.f.Seek(0, 2)
	if err != nil {
		return seek, 0, err
	}
	n, err = sf.f.WriteAt(b, seek)
	if err != nil {
		return seek, n, err
	}
	return seek, n, sf.f.Sync() // ensure that the write is done.
}

// Close closes the underlying file.
func (sf *SyncFile) Close() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if sf.f == nil {
		return errors.New("Error: file is nil")
	}
	return sf.f.Close()
}

var bufPool = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}
