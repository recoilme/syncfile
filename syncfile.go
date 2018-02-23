// Package syncfile implements concurrent write/read to/from file
package syncfile

import (
	"bytes"
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
func (sf *SyncFile) Append(b []byte) error {
	buf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(buf)
	buf.Reset()
	_, _ = buf.Write(b)
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
	//_, err = sf.f.Seek(seek, 0) //problem here - not real seek issue https://github.com/golang/go/issues/24035
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

func (sf *SyncFile) write(b []byte) error {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.f.Seek(0, 2)
	_, err := sf.f.Write(b)
	if err != nil {
		return nil
	}
	return sf.f.Sync() // ensure that the write is done.
}

// Close closes the underlying file.
func (sf *SyncFile) Close() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.f.Close()
}

var bufPool = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}
