package zlog

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zfile"
)

func openFile(filepa string) (file *os.File, fileName, fileDir string, err error) {
	fullPath := zfile.RealPath(filepa)
	fileDir = filepath.Dir(filepa)+"/"
	fileName = filepath.Base(filepa)
	_ = mkdirLog(fileDir)
	if zfile.FileExist(fullPath) {
		file, err = os.OpenFile(fullPath, os.O_APPEND|os.O_RDWR, 0644)
	} else {
		file, err = os.OpenFile(fullPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	}
	if err != nil {
		return nil, "", "", err
	}

	return
}

// SetLogFile Setting log file output
func (log *Logger) SetLogFile(filepath string) {
	fileObj, fileName, fileDir, _ := openFile(filepath)
	log.DisableConsoleColor()
	log.mu.Lock()
	defer log.mu.Unlock()

	log.CloseFile()
	log.file = fileObj
	log.out = fileObj
	log.FileMaxSize = 0
	log.fileDir = fileDir
	log.fileName = fileName
}

func (log *Logger) SetSaveLogFile(filepath string) {
	log.SetLogFile(filepath)
	log.fileAndStdout = true
	log.out = io.MultiWriter(log.file, os.Stdout)
}

func (log *Logger) CloseFile() {
	if log.file != nil {
		_ = log.file.Close()
		log.file = nil
		log.out = os.Stderr
	}
}

func oldLogFile(fileDir, fileName string) string {
	ext := path.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)
	timeStr := time.Now().Format("2006-01-02")
	oldLogFile := fileDir + "/" + name + "." + timeStr + ext
judge:
	for {
		if !zfile.FileExist(oldLogFile) {
			break judge
		} else {
			oldLogFile = fileDir + "/" + name + "." + timeStr + "_" + strconv.Itoa(int(time.Now().UnixNano())) + ext
		}
	}

	return oldLogFile
}

func mkdirLog(dir string) (e error) {
	if zfile.DirExist(dir) {
		return
	}
	if err := os.MkdirAll(dir, 0775); err != nil && os.IsPermission(err) {
		e = err
	}
	return
}
