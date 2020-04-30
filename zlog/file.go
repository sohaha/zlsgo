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
	"github.com/sohaha/zlsgo/ztime/cron"
)

func openFile(filepa string, archive bool) (file *os.File, fileName, fileDir string, err error) {
	fullPath := zfile.RealPath(filepa)
	fileDir, fileName = filepath.Split(fullPath)
	if archive {
		archiveName := ztime.FormatTime(time.Now(), "Y-m-d")
		ext := filepath.Ext(fileName)
		base := strings.TrimSuffix(fileName, ext)
		fileDir = zfile.RealPathMkdir(fileDir+base, true)
		fileName = archiveName + ext
		fullPath = fileDir + fileName
	}
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

// SetFile Setting log file output
func (log *Logger) SetFile(filepath string, archive ...bool) {
	log.DisableConsoleColor()
	logArchive := len(archive) > 0 && archive[0]
	if logArchive {
		c := cron.New()
		_, _ = c.Add("0 0 * * *", func() {
			log.setLogfile(filepath, logArchive)
		})
		c.Run()
	}
	log.setLogfile(filepath, logArchive)
}

func (log *Logger) setLogfile(filepath string, archive bool) {
	fileObj, fileName, fileDir, _ := openFile(filepath, archive)
	log.mu.Lock()
	log.CloseFile()
	log.file = fileObj
	log.fileDir = fileDir
	log.fileName = fileName
	if log.fileAndStdout {
		log.out = io.MultiWriter(log.file, os.Stdout)
	} else {
		log.out = fileObj
	}
	log.mu.Unlock()
}

func (log *Logger) Discard() {
	log.mu.Lock()
	log.out = ioutil.Discard
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
