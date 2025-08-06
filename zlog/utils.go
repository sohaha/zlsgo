package zlog

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type astCacheEntry struct {
	File    *ast.File
	ModTime time.Time
	FileSet *token.FileSet
}

var astCache = struct {
	entries map[string]astCacheEntry
	sync.RWMutex
}{
	entries: make(map[string]astCacheEntry),
}

var fileSetPool = sync.Pool{
	New: func() interface{} {
		return token.NewFileSet()
	},
}

func getFileSet() *token.FileSet {
	return fileSetPool.Get().(*token.FileSet)
}

func putFileSet(fset *token.FileSet) {
	if fset == nil {
		return
	}
	fileSetPool.Put(fset)
}

func parseFileWithCache(filename string) (*ast.File, *token.FileSet, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return nil, nil, err
	}

	modTime := info.ModTime()
	absPath, err := filepath.Abs(filename)
	if err != nil {
		absPath = filename
	}

	astCache.RLock()
	entry, ok := astCache.entries[absPath]
	astCache.RUnlock()

	if ok && entry.ModTime.Equal(modTime) {
		return entry.File, entry.FileSet, nil
	}

	fset := getFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		putFileSet(fset)
		return nil, nil, err
	}

	newEntry := astCacheEntry{
		File:    file,
		ModTime: modTime,
		FileSet: fset,
	}

	astCache.Lock()
	astCache.entries[absPath] = newEntry
	astCache.Unlock()

	return file, fset, nil
}

var zprinterPool = sync.Pool{
	New: func() interface{} {
		p := new(zprinter)
		p.visited = make(map[visit]int)
		return p
	},
}

func getZprinter(w io.Writer) *zprinter {
	p := zprinterPool.Get().(*zprinter)
	p.Writer = w
	p.depth = 0

	for k := range p.visited {
		delete(p.visited, k)
	}

	return p
}

func putZprinter(p *zprinter) {
	if p == nil {
		return
	}
	p.Writer = nil
	p.tw = nil
	zprinterPool.Put(p)
}

func formatDateAppend(buf *bytes.Buffer, t time.Time) {
	year, month, day := t.Date()

	itoa(buf, year, 4)
	buf.WriteByte('/')

	if month < 10 {
		buf.WriteByte('0')
	}
	itoa(buf, int(month), 0)
	buf.WriteByte('/')
	if day < 10 {
		buf.WriteByte('0')
	}
	itoa(buf, day, 0)
	buf.WriteByte(' ')
}

func formatTimeAppend(buf *bytes.Buffer, t time.Time) {
	hour, min, sec := t.Clock()

	if hour < 10 {
		buf.WriteByte('0')
	}
	itoa(buf, hour, 0)
	buf.WriteByte(':')

	if min < 10 {
		buf.WriteByte('0')
	}
	itoa(buf, min, 0)
	buf.WriteByte(':')

	if sec < 10 {
		buf.WriteByte('0')
	}
	itoa(buf, sec, 0)
}
