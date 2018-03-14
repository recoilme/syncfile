// Package syncfile implements concurrent write/read to/from file
package syncfile

import (
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

// WriteNoSync writes len(b) bytes to the File without sync
// It returns the number of bytes written and an error, if any.
// Write returns a non-nil error when n != len(b).
func (sf *SyncFile) WriteNoSync(b []byte) (seek int64, n int, err error) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	seek, err = sf.f.Seek(0, 2)
	if err != nil {
		return seek, 0, err
	}
	n, err = sf.f.WriteAt(b, seek)
	return seek, n, err // not ensure that the write is done.
}

// Sync do sync on file
func (sf *SyncFile) Sync() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.f.Sync()
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
	sf.mu.RLock() //with RLock -  not working
	defer sf.mu.RUnlock()
	var err error

	byteSlice := make([]byte, size)

	//TODO read by chanks
	_, err = sf.f.ReadAt(byteSlice, seek) // .Read(byteSlice)

	//buf.Write(byteSlice)
	if err != nil {
		return nil, err
	}
	return byteSlice, nil
}

// WriteAt writes len(b) bytes to the File at offest off.
// If off<0 write at the end of file
// It returns the number of bytes written and an error, if any.
// Write returns a non-nil error when n != len(b).
func (sf *SyncFile) WriteAt(b []byte, off int64) (seek int64, n int, err error) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if off < 0 {
		off, err = sf.f.Seek(0, 2)
		if err != nil {
			return off, 0, err
		}
	}
	n, err = sf.f.WriteAt(b, off)
	if err != nil {
		return off, n, err
	}
	return seek, n, sf.f.Sync()
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
