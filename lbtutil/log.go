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

func NewLogger(prefix string, tag string) Logger { return &logger{prefix: prefix, tag: tag} }

func (l *logger) Debug(format string, params ...interface{}) {
	logByLevel(Ldebug, "DEBUG", l.prefix, l.tag, format, params...)
}

func (l *logger) Info(format string, params ...interface{}) {
	logByLevel(Linfo, "INFO", l.prefix, l.tag, format, params...)
}

func (l *logger) Warn(format string, params ...interface{}) {
	logByLevel(Lwarn, "WARN", l.prefix, l.tag, format, params...)
}

func (l *logger) Error(format string, params ...interface{}) {
	logByLevel(Lerror, "ERROR", l.prefix, l.tag, format, params...)
}


func logByLevel(level int, lvstr, prefix, tag, format string, params ...interface{}) {
	// check level
	if level < int(atomic.LoadInt32(&logLevel)) { return }
	// time str
	now := time.Now()
	tstr := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
			now.Year(), now.Month(), now.Day(),
			now.Hour(), now.Minute(), now.Second())
	// format
	body := fmt.Sprintf(format, params...)
	// build str
	n := len(tstr) + 1 + len(prefix) + 1 + len(lvstr) + 1 + len(tag) + 1 + len(body) + 1
	b := make([]byte, n, n)
	pos := 0
	var sep byte = ' '
	copy(b[pos:], tstr)
	pos += len(tstr)
	b[pos] = sep
	pos += 1
	copy(b[pos:], prefix)
	pos += len(prefix)
	b[pos] = sep
	pos += 1
	copy(b[pos:], lvstr)
	pos += len(lvstr)
	b[pos] = sep
	pos += 1
	copy(b[pos:], tag)
	pos += len(tag)
	b[pos] = sep
	pos += 1
	copy(b[pos:], body)
	pos += len(body)
	b[pos] = '\n'
	/*
	sb := strings.Builder{}
	sb.Grow(len(prefix) + 1 + len(tstr) + 1 + len(tag) + 1 + len(body) + 1)
	sep := " "
	sb.WriteString(tstr)
	sb.WriteString(sep)
	sb.WriteString(prefix)
	sb.WriteString(sep)
	sb.WriteString(lvstr)
	sb.WriteString(sep)
	sb.WriteString(tag)
	sb.WriteString(sep)
	sb.WriteString(body)
	sb.WriteString("\n")
	*/
	// write
	logOut.Write(b)
}