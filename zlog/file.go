package zlog

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/ztime"
)

var LogMaxDurationDate = 15

func openFile(filepa string, archive bool) (file *zfile.MemoryFile, fileName, fileDir string, err error) {
	fullPath := zfile.RealPath(filepa)
	fileDir, fileName = filepath.Split(fullPath)
	opt := []zfile.MemoryFileOption{zfile.MemoryFileAutoFlush(1)}
	if archive {
		ext := filepath.Ext(fileName)
		base := strings.TrimSuffix(fileName, ext)
		fileDir = zfile.RealPathMkdir(fileDir+base, true)
		fullPath = fileDir + fileName
		lastArchiveName := ""
		opt = append(opt, zfile.MemoryFileFlushBefore(func(f *zfile.MemoryFile) error {
			archiveName := ztime.Now("Y-m-d")
			if lastArchiveName != archiveName {
				if lastArchiveName != "" {
					// Delete the log file that is too old
					go func() {
						now := ztime.UnixMicro(ztime.Clock())
						_ = filepath.Walk(fileDir, func(path string, info os.FileInfo, err error) error {
							if err != nil {
								return err
							}
							if info.IsDir() {
								return nil
							}

							// The log may have been modified, so the modification time of the file is no longer used here.
							// if info.ModTime().AddDate(0, 0, LogMaxDurationDate).Before(now) {
							// 	_ = os.Remove(path)
							// }

							date, err := ztime.Parse(strings.TrimSuffix(filepath.Base(path), ext), "Y-m-d")
							if err == nil && date.AddDate(0, 0, LogMaxDurationDate).Before(now) {
								_ = os.Remove(path)
							}
							return nil

						})
					}()
				}
				lastArchiveName = archiveName
			}
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
