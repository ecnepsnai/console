// Package console provides a simple interface for logging things to stdout & a log file
package console

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/ecnepsnai/zip"
	"github.com/fatih/color"
)

const (
	// LevelDebug debug level includes debugging information and is very verbose
	LevelDebug = 3
	// LevelInfo informational messages for normal operation of the application
	LevelInfo = 2
	// LevelWarn warning messages for potential issues
	LevelWarn = 1
	// LevelError error messages for problems
	LevelError = 0
)

// Console describes a log object
type Console struct {
	Level    int
	filePath string
	file     *os.File
	mutex    *sync.Mutex
}

// NewMemory create a new in-memory console instance that does not write to disk
func NewMemory(level int) Console {
	return Console{
		mutex: &sync.Mutex{},
		Level: level,
	}
}

// New creates a new logging instance
func New(logPath string, level int) (*Console, error) {
	logFile, err := newFile(logPath)
	if err != nil {
		return nil, err
	}

	l := Console{
		filePath: logPath,
		file:     logFile,
		mutex:    &sync.Mutex{},
		Level:    level,
	}
	return &l, nil
}

func newFile(logPath string) (*os.File, error) {
	return os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
}

// Close close the log file
func (l *Console) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

// Rotate retire the current log file into a gzipped file with todays date
func (l *Console) Rotate(destinationDir string) error {
	if l.file == nil {
		return nil
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	destFileName := destinationDir + "/log." + time.Now().Format("2006-01-02") + ".zip"
	l.Close()
	l.file = nil

	data, err := ioutil.ReadFile(l.filePath)
	if err != nil {
		fmt.Printf("%s\n", debug.Stack())
		return err
	}

	archive, err := zip.NewZipFile(destFileName)
	defer archive.Close()
	if err != nil {
		fmt.Printf("%s\n", debug.Stack())
		return err
	}

	archive.AddFile("console.log", data)

	os.Remove(l.filePath)
	newFile, err := newFile(l.filePath)
	if err != nil {
		fmt.Printf("%s\n", debug.Stack())
		return err
	}
	l.file = newFile

	return nil
}

func (l *Console) write(message string) {
	if l.file == nil {
		return
	}

	l.mutex.Lock()
	if l.file != nil {
		_, err := l.file.WriteString(time.Now().Format(time.RFC3339) + " " + message + "\n")
		if err != nil {
			// Try opening the file again
			l.file.Close()
			newFile, err := newFile(l.filePath)
			if err != nil {
				fmt.Printf("Error writing to log: %s", err.Error())
			} else {
				l.file = newFile
			}
		}
	}
	l.mutex.Unlock()
}

// Debug print debug information to the console if verbose logging is enabled
// Safe to call with sensitive data, but verbose logging should not be enabled on production instances
func (l *Console) Debug(format string, a ...interface{}) {
	if l.Level >= LevelDebug {
		fmt.Printf("%s %s\n", color.HiBlackString("[DEBUG]"), fmt.Sprintf(format, a...))
		l.write("[DEBUG] " + fmt.Sprintf(format, a...))
	}
}

// Info print informational message to the console
func (l *Console) Info(format string, a ...interface{}) {
	if l.Level >= LevelInfo {
		fmt.Printf("%s %s\n", color.BlueString("[INFO] "), fmt.Sprintf(format, a...))
		l.write("[INFO]  " + fmt.Sprintf(format, a...))
	}
}

// Warn print warning information to the console
func (l *Console) Warn(format string, a ...interface{}) {
	if l.Level >= LevelWarn {
		fmt.Printf("%s %s\n", color.YellowString("[WARN] "), fmt.Sprintf(format, a...))
		l.write("[WARN]  " + fmt.Sprintf(format, a...))
	}
}

// Error print error information to the console
func (l *Console) Error(format string, a ...interface{}) {
	if l.Level >= LevelError {
		stack := string(debug.Stack())
		fmt.Printf("%s %s\n%s\n", color.RedString("[ERROR]"), fmt.Sprintf(format, a...), stack)
		l.write(fmt.Sprintf("[ERROR] %s\n%s", fmt.Sprintf(format, a...), stack))
	}
}

// ErrorDesc print an error object with description
func (l *Console) ErrorDesc(desc string, err error) {
	l.Error("%s: %s", desc, err.Error())
}

// Fatal print fatal error and exit the app
func (l *Console) Fatal(format string, a ...interface{}) {
	fmt.Printf("%s\n", color.RedString("[FATAL] "+fmt.Sprintf(format, a...)))
	l.write("[FATAL] " + fmt.Sprintf(format, a...))
	os.Exit(1)
}
