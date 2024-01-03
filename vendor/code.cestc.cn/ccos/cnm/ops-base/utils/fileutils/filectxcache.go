package fileutils

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	m = sync.Map{}
	l = sync.Mutex{}
)

type fileCtxCache struct {
	fileName string
	once     sync.Once
	ticker   *time.Ticker
	value    string
	err      error
}

func newFileCtxCache(fileName string) *fileCtxCache {
	f := &fileCtxCache{
		fileName: fileName,
		once:     sync.Once{},
		ticker:   time.NewTicker(10 * time.Second),
	}
	f.value, f.err = f.getContext()
	if f.err != nil {
		f.run()
	}
	return f
}

func Get(path, name string) (string, error) {

	fileName := fmt.Sprintf("%s/%s", path, name)

	return GetByFileName(fileName)
}

func GetByFileName(fileName string) (string, error) {
	var (
		fileCtx *fileCtxCache
	)
	l.Lock()
	defer l.Unlock()

	val, ok := m.Load(fileName)
	if !ok {
		fileCtx = newFileCtxCache(fileName)
		if fileCtx.err != nil {
			m.Store(fileName, fileCtx)
		}
	} else {
		fileCtx, _ = val.(*fileCtxCache)
	}
	return fileCtx.value, fileCtx.err
}

func (f *fileCtxCache) getContext() (string, error) {
	file, err := os.Open(f.fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	filesize := fileInfo.Size()
	buffer := make([]byte, filesize)

	_, err = file.Read(buffer)
	if err != nil {
		return "", err
	}
	return string(buffer), nil
}

func (f *fileCtxCache) run() {
	f.once.Do(func() {
		go func() {
			for {
				select {
				case <-f.ticker.C:
					f.value, f.err = f.getContext()
				}
			}
		}()
	})
}
