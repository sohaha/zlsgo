package zlog

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/ztime"
)

var LogMaxDurationDate = 15

// SetFileErrorHandler sets the handler for file setup and close/flush errors.
// A nil handler reports errors to os.Stderr. Handlers run without the logger
// lock and must be safe for concurrent calls. File replacement proceeds even
// when closing the old file fails: MemoryFile.Close has already closed it.
// This handler does not cover asynchronous MemoryFile auto-flush errors.
func (log *Logger) SetFileErrorHandler(handler func(error)) {
	log.mu.Lock()
	log.fileErrorHandler = handler
	log.mu.Unlock()
}

func (log *Logger) reportFileError(err error) {
	if err == nil {
		return
	}
	log.mu.RLock()
	handler := log.fileErrorHandler
	log.mu.RUnlock()
	if handler != nil {
		handler(err)
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "zlog:", err)
	}
}

func closeLogFile(file *zfile.MemoryFile) error {
	if file == nil {
		return nil
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close log file %q: %w", file.Name(), err)
	}
	return nil
}

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

							if LogMaxDurationDate > 0 {
								date, err := ztime.Parse(strings.TrimSuffix(filepath.Base(path), ext), "Y-m-d")
								if err == nil && date.AddDate(0, 0, LogMaxDurationDate).Before(now) {
									_ = os.Remove(path)
								}
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

func (log *Logger) SetLevelFile(level int, filepath string, archive ...bool) {
	log.DisableConsoleColor()
	logArchive := len(archive) > 0 && archive[0]
	log.setLevelFile(level, filepath, logArchive, false)
}

func (log *Logger) SetLevelSaveFile(level int, filepath string, archive ...bool) {
	log.DisableConsoleColor()
	logArchive := len(archive) > 0 && archive[0]
	log.setLevelFile(level, filepath, logArchive, true)
}

func (log *Logger) setLogfile(filepath string, archive bool, save ...bool) {
	fileObj, fileName, fileDir, err := openFile(filepath, archive)
	if err != nil || fileObj == nil {
		// Keep the existing output instead of installing a nil writer.
		log.reportFileError(err)
		return
	}
	log.mu.Lock()
	err = log.closeFileLocked()
	if len(save) > 0 && save[0] {
		log.fileAndStdout = true
	}
	log.file = fileObj
	log.fileDir = fileDir
	log.fileName = fileName
	if log.fileAndStdout {
		log.Out = io.MultiWriter(fileObj, os.Stdout)
	} else {
		log.Out = fileObj
	}
	log.mu.Unlock()
	log.reportFileError(err)
}

func (log *Logger) setLevelFile(level int, filepath string, archive bool, andStdout bool) {
	fileObj, _, _, err := openFile(filepath, archive)
	if err != nil || fileObj == nil {
		// Keep the existing output instead of installing a nil writer.
		log.reportFileError(err)
		return
	}
	log.mu.Lock()
	if log.levelFiles == nil {
		log.levelFiles = map[int]*levelFile{}
	}
	if old, ok := log.levelFiles[level]; ok && old != nil && old.file != nil {
		err = closeLogFile(old.file)
	}
	var out io.Writer = fileObj
	if andStdout {
		out = io.MultiWriter(fileObj, os.Stdout)
	}
	log.levelFiles[level] = &levelFile{file: fileObj, out: out}
	log.mu.Unlock()
	log.reportFileError(err)
}

func (log *Logger) Discard() {
	log.mu.Lock()
	err := errors.Join(log.closeFileLocked(), log.closeLevelFilesLocked())
	log.Out = ioutil.Discard
	log.level.Store(LogNot)
	log.mu.Unlock()
	log.reportFileError(err)
}

func (log *Logger) SetSaveFile(filepath string, archive ...bool) {
	log.DisableConsoleColor()
	log.setLogfile(filepath, len(archive) > 0 && archive[0], true)
}

func (log *Logger) CloseLevelFiles() {
	log.mu.Lock()
	err := log.closeLevelFilesLocked()
	log.mu.Unlock()
	log.reportFileError(err)
}

func (log *Logger) closeLevelFilesLocked() error {
	var errs []error
	if log.levelFiles != nil {
		for _, lf := range log.levelFiles {
			if lf != nil && lf.file != nil {
				errs = append(errs, closeLogFile(lf.file))
			}
		}
		log.levelFiles = nil
	}
	return errors.Join(errs...)
}

func (log *Logger) CloseFile() {
	log.mu.Lock()
	err := log.closeFileLocked()
	log.mu.Unlock()
	log.reportFileError(err)
}

func (log *Logger) closeFileLocked() error {
	var err error
	if log.file != nil {
		err = closeLogFile(log.file)
		log.file = nil
		log.Out = os.Stdout
	}
	return err
}
