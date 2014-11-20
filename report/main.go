package report
/**
 * Reporting lib.
 * Simple abstraction for logging.
 */
import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
)

var (
	msg *log.Logger
	err *log.Logger

	logFd *os.File
	errFd *os.File

	IsVerbose bool
)

// https://github.com/graarh/golang/tree/master/trace
func backtrace(skip int) string {
	var stack string

	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		Func := runtime.FuncForPC(pc)
		stack += fmt.Sprintf("\n  at %s(%s:%v)", Func.Name(), file, line)
		if Func.Name() == "main.main" {
			break
		}
	}

	return stack
}

func Err(e error) {
	err.Println(fmt.Sprintf("%s %s", e.Error(), backtrace(1)))
}

func Msg(s string) {
	msg.Println(s)
}

func Debug(s string) {
	if IsVerbose {
		msg.Println(s)
	}
}

// Same as Msg except his reports abuse
func Abuse(s string) {
	msg.Println(s)
}

func open(path string) (*os.File, error) {
	r, e := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if e != nil {
		return nil, e
	}
	return r, nil
}

func Init(prefix string, path string, isVerbose bool) error {
	var (
		e error
	)
	IsVerbose = isVerbose
	logFd, e = open(path + "/activity.log")
	if e != nil {
		return e
	}
	errFd, e = open(path + "/error.log")
	if e != nil {
		logFd.Close()
		return e
	}

	if isVerbose {
		msg = log.New(io.MultiWriter(os.Stdout, logFd), prefix + " ", log.Ldate|log.Ltime)
		err = log.New(io.MultiWriter(os.Stderr, errFd), prefix + " ", log.Ldate|log.Ltime)
	} else {
		msg = log.New(io.Writer(logFd), prefix + " ", log.Ldate|log.Ltime)
		err = log.New(io.Writer(errFd), prefix + " ", log.Ldate|log.Ltime)
	}
	return nil
}

func Close() {
	if logFd != nil {
		logFd.Close()
	}
	if errFd != nil {
		errFd.Close()
	}
}
