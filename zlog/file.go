package zlog

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/ztime"
)

func openFile(filepa string, archive bool) (file *zfile.MemoryFile, fileName, fileDir string, err error) {
	fullPath := zfile.RealPath(filepa)
	fileDir, fileName = filepath.Split(fullPath)
	opt := []zfile.MemoryFileOption{zfile.MemoryFileAutoFlush(1)}
	if archive {
		ext := filepath.Ext(fileName)
		base := strings.TrimSuffix(fileName, ext)
		fileDir = zfile.RealPathMkdir(fileDir+base, true)
		fullPath = fileDir + fileName
		opt = append(opt, zfile.MemoryFileFlushBefore(func(f *zfile.MemoryFile) error {
			archiveName := ztime.FormatTime(time.Now(), "Y-m-d")
			fileName = archiveName + ext
			f.SetName(fileDir + fileName)
			return nil
		}))
	}
	f := zfile.NewMemoryFile(fullPath, opt...)
	return f, fileName, fileDir, nil
}

// SetFile Setting log file output
func (log *Logger) SetFile(filepath string, archive ...bool) {
	log.DisableConsoleColor()
	logArchive := len(archive) > 0 && archive[0]
	log.setLogfile(filepath, logArchive)
}

func (log *Logger) setLogfile(filepath string, archive bool) {
	fileObj, fileName, fileDir, _ := openFile(filepath, archive)
	log.mu.Lock()
	defer log.mu.Unlock()
	log.CloseFile()
	log.file = fileObj
	log.fileDir = fileDir
	log.fileName = fileName
	if log.fileAndStdout {
		log.out = io.MultiWriter(log.file, os.Stdout)
	} else {
		log.out = fileObj
	}
}

func (log *Logger) Discard() {
	log.mu.Lock()
	log.out = ioutil.Discard
	if log.file != nil {
		_ = log.file.Close()
	}
	log.level = LogNot
	log.mu.Unlock()
}

func (log *Logger) SetSaveFile(filepath string, archive ...bool) {
	log.SetFile(filepath, archive...)
	log.mu.Lock()
	log.fileAndStdout = true
	log.out = io.MultiWriter(log.file, os.Stdout)
	log.mu.Unlock()
}

func (log *Logger) CloseFile() {
	if log.file != nil {
		_ = log.file.Close()
		log.file = nil
		log.out = os.Stderr
	}
}

// func oldLogFile(fileDir, fileName string) string {
// 	ext := path.Ext(fileName)
// 	name := strings.TrimSuffix(fileName, ext)
// 	timeStr := time.Now().Format("2006-01-02")
// 	oldLogFile := fileDir + "/" + name + "." + timeStr + ext
// judge:
// 	for {
// 		if !zfile.FileExist(oldLogFile) {
// 			break judge
// 		} else {
// 			oldLogFile = fileDir + "/" + name + "." + timeStr + "_" + strconv.Itoa(int(time.Now().UnixNano())) + ext
// 		}
// 	}
//
// 	return oldLogFile
// }

func mkdirLog(dir string) (e error) {
	if zfile.DirExist(dir) {
		return
	}
	if err := os.MkdirAll(dir, 0775); err != nil && os.IsPermission(err) {
		e = err
	}
	return
}
