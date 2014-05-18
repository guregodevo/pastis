package pastis

import (
	"fmt"
	"io"
	"log"
	"os"
)

//An API Logger whose level is configurable
type Logger struct {
	level int
	info  *log.Logger
	debug *log.Logger
	fatal *log.Logger
	warn  *log.Logger
	err   *log.Logger
}

const (
	DEBUG = 1
	INFO  = 2
	WARN  = 3
	ERROR = 4
	FATAL = 5
	OFF   = 6
)

//Level converts a strong value of a level into into its integer value
func LevelInt(level string) int {
	switch {
	case "INFO" == level:
		return INFO
	case "DEBUG" == level:
		return DEBUG
	case "WARN" == level:
		return WARN
	case "ERROR" == level:
		return ERROR
	case "FATAL" == level:
		return FATAL
	}
	return OFF  //default level
}

// GetLogger retrieves a logger having the given level.
// The level defines the minimum set of levels recognized by the system, that is OFF, FATAL, ERROR, WARN, INFO, DEBUG and ALL.
func GetLogger(level string) *Logger {
	levelInt := LevelInt(level)
	pastisLogger := &Logger{level: levelInt}
	pastisLogger.info = nativeLogger(os.Stdout, "INFO", log.Ldate|log.Ltime)
	pastisLogger.debug = nativeLogger(os.Stdout, "DEBUG", log.Ldate|log.Ltime)
	pastisLogger.warn = nativeLogger(os.Stdout, "WARN", log.Ldate|log.Ltime)
	pastisLogger.err = nativeLogger(os.Stdout, "ERROR", log.Ldate|log.Ltime)
	pastisLogger.fatal = nativeLogger(os.Stdout, "FATAL", log.Ldate|log.Ltime)
	return pastisLogger
}

func nativeLogger(w io.Writer, prefix string, flag int) *log.Logger {
	return log.New(w, fmt.Sprintf("%s: ", prefix), flag)
}

// Set the minimum set of levels recognized by the system, that is OFF, FATAL, ERROR, WARN, INFO, DEBUG and ALL.
func (logger *Logger) SetLevel(level string) {
	logger.level = LevelInt(level)
}

//SetOutput sets the output destination for the logger having the given level.
func (logger *Logger) SetOuputs(w io.Writer, flag int, levels ...string) {
	for _, l := range levels {
		logger.SetOuput(l, w, flag)
	}
}


//SetOutput sets the output destination for the logger having the given level.
func (logger *Logger) SetOuput(level string, w io.Writer, flag int) {
	levelNum := LevelInt(level)
	switch {
	case INFO == levelNum:
		logger.info = nativeLogger(w, level, flag)
	case DEBUG == levelNum:
		logger.debug = nativeLogger(w, level, flag)
	case WARN == levelNum:
		logger.warn = nativeLogger(w, level, flag)
	case ERROR == levelNum:
		logger.err = nativeLogger(w, level, flag)
	case FATAL == levelNum:
		logger.fatal = nativeLogger(w, level, flag)
	default:
	}
}

//Log a message object with the INFO level. Arguments are handled in the manner of fmt.Print.
func (logger *Logger) Info(s string) {
	if INFO >= logger.level {
		logger.info.Print(s)
	}
}

//Log a message object with the DEBUG level. Arguments are handled in the manner of fmt.Print.
func (logger *Logger) Debug(s string) {
	if DEBUG >= logger.level {
		logger.debug.Print(s)
	}
}

//Log a message object with the ERROR level. Arguments are handled in the manner of fmt.Print.
func (logger *Logger) Error(s string) {
	if ERROR >= logger.level {
		logger.err.Print(s)
	}
}

//Log a message object with the FATAL level. Arguments are handled in the manner of fmt.Print.
func (logger *Logger) Fatal(s string) {
	if FATAL >= logger.level {
		logger.fatal.Print(s)
	}
}

//Log a message object with the WARN level. Arguments are handled in the manner of fmt.Print.
func (logger *Logger) Warn(s string) {
	if WARN >= logger.level {
		logger.warn.Print(s)
	}
}

//Log a message object with the INFO level. Arguments are handled in the manner of fmt.Printf.
func (logger *Logger) Infof(format string, v ...interface{}) {
	if INFO >= logger.level {
		logger.info.Printf(format, v...)
	}	
}

//Log a message object with the DEBUG level. Arguments are handled in the manner of fmt.Printf.
func (logger *Logger) Debugf(format string, v ...interface{}) {
	if DEBUG >= logger.level {
		logger.debug.Printf(format, v...)
	}	
}

//Log a message object with the ERROR level. Arguments are handled in the manner of fmt.Printf.
func (logger *Logger) Errorf(format string, v ...interface{}) {
	if ERROR >= logger.level {
		logger.err.Printf(format, v...)
	}	
}

//Log a message object with the FATAL level. Arguments are handled in the manner of fmt.Printf.
func (logger *Logger) Fatalf(format string, v ...interface{}) {
	if FATAL >= logger.level {
		logger.fatal.Printf(format, v...)
	}
}

//Log a message object with the WARN level. Arguments are handled in the manner of fmt.Printf.
func (logger *Logger) Warnf(format string, v ...interface{}) {
	if WARN >= logger.level {
		logger.warn.Printf(format, v...)
	}
}
