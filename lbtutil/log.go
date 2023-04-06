package lbtutil

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	_ = iota
	Ldebug
	Linfo
	Lwarn
	Lerror
)

var (
	logLevel = int32(Ldebug)
	logOut io.Writer = os.Stdout
	logLenLimit int = 1000
)

func SetLogLevel(level int) { atomic.StoreInt32(&logLevel, int32(level)) }
func SetLogOut(out io.Writer) { logOut = out }

type Logger interface {
	Debug(format string, params ...interface{})
	Info(format string, params ...interface{})
	Warn(format string, params ...interface{})
	Error(format string, params ...interface{})
}

type logger struct {
	prefix string
	tag string
}

var log = NewLogger(strconv.Itoa(os.Getpid()), "lbtutil")

func NewLogger(prefix string, tag string) Logger { return &logger{prefix: fmt.Sprintf("%s|%s", prefix, tag), tag: tag} }

func (l *logger) Debug(format string, params ...interface{}) {
	logByLevel(Ldebug, "DEBUG", l.prefix, format, params...)
}

func (l *logger) Info(format string, params ...interface{}) {
	logByLevel(Linfo, "INFO", l.prefix, format, params...)
}

func (l *logger) Warn(format string, params ...interface{}) {
	logByLevel(Lwarn, "WARN", l.prefix, format, params...)
}

func (l *logger) Error(format string, params ...interface{}) {
	logByLevel(Lerror, "ERROR", l.prefix, format, params...)
}


func logByLevel(level int, lvstr, prefix, format string, params ...interface{}) {
	// check level
	if level < int(atomic.LoadInt32(&logLevel)) { return }
	// time str
	now := time.Now()
	tstr := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d.%03d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/1e6)
	// format
	body := fmt.Sprintf(format, params...)
	if len(body) > logLenLimit {
		body = body[:logLenLimit]
	}
	// build str
	var sepStr = "] ["
	pathInfo := ""
	n := 1 + len(tstr) + 3 + len(lvstr) + 3 + len(prefix) + 3 + len(body) + 3 + len(pathInfo) + 2
	b := make([]byte, n, n)

	b[0] = '['
	pos := 1
	copy(b[pos:], tstr)
	pos += len(tstr)
	copy(b[pos:], sepStr)
	pos += len(sepStr)
	copy(b[pos:], lvstr)
	pos += len(lvstr)
	copy(b[pos:], sepStr)
	pos += len(sepStr)
	copy(b[pos:], prefix)
	pos += len(prefix)
	copy(b[pos:], sepStr)
	pos += len(sepStr)
	copy(b[pos:], body)
	pos += len(body)
	copy(b[pos:], sepStr)
	pos += len(sepStr)
	copy(b[pos:], pathInfo)
	pos += len(pathInfo)
	copy(b[pos:], "]\n")
	logOut.Write(b)
}
