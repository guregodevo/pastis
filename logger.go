package pastis

import (
	"log"
	"fmt"
	"os"
	"io"
)


type Logger struct {
  Level int
  info *log.Logger
  debug *log.Logger
  fatal *log.Logger
  warn *log.Logger
  err *log.Logger
}

const (
	INFO = 1
	DEBUG = 2
	WARN = 3
	ERROR = 4
	FATAL = 5
	OFF = 6	
)

func LevelInt(level string) int {
	switch {
    case "INFO" == level :
        return INFO
    case "DEBUG" == level :
        return DEBUG
    case "WARN" == level :
        return WARN
    case "ERROR" == level :
        return ERROR
    case "FATAL" == level :
        return FATAL
    case "OFF" == level :
        return OFF
	}
    return OFF	
}

func (logger *Logger) Switch(level int) *log.Logger {
	switch {
    case INFO == level :
        return logger.info
    case DEBUG == level :
        return logger.debug
    case WARN == level :
        return logger.warn
    case ERROR == level :
        return logger.err
    case FATAL == level :
        return logger.fatal
    default :
        return nil
	}
}


// Retrieve a logger having the given level.
// The level defines the minimum set of levels recognized by the system, that is OFF, FATAL, ERROR, WARN, INFO, DEBUG and ALL.
func GetLogger(level string) *Logger {
 	levelInt := LevelInt(level)
 	pastisLogger := &Logger{ Level: levelInt }
 	pastisLogger.info = nativeLogger(os.Stdout, "INFO", log.Ltime)
 	pastisLogger.debug = nativeLogger(os.Stdout, "DEBUG", log.Ltime)
	pastisLogger.warn = nativeLogger(os.Stdout, "WARN", log.Ltime)
	pastisLogger.err = nativeLogger(os.Stdout, "ERROR", log.Ltime)
	pastisLogger.fatal = nativeLogger(os.Stdout, "FATAL", log.Ltime)
 	return pastisLogger
}

func nativeLogger(w io.Writer, prefix string, flag int) *log.Logger {
	return log.New(w, fmt.Sprintf("%s: ", prefix), flag)
}


//SetOutput sets the output destination for the standard logger.
func (logger *Logger) SetOuput(level string, w io.Writer, flag int) {
	levelNum := LevelInt(level)
	switch {
    case INFO == levelNum :
        logger.info = nativeLogger(w, level, flag)
    case DEBUG == levelNum :
        logger.debug = nativeLogger(w, level, flag)
    case WARN == levelNum :
        logger.warn = nativeLogger(w, level, flag)
    case ERROR == levelNum :
        logger.err = nativeLogger(w, level, flag)
    case FATAL == levelNum :
        logger.fatal = nativeLogger(w, level, flag)
    default :
	}		
}

//Log a message object with the given level. Arguments are handled in the manner of fmt.Println.
func (logger *Logger) Trace(level int, s string) {
	if (level <= logger.Level) {
		logger.Switch(level).Print(s)		
	}
}

//Log a message object with the given level. Arguments are handled in the manner of fmt.Printf.
func (logger *Logger) TraceF(level int, format string, v ...interface{} ) {
	if (level <= logger.Level) {
		logger.Switch(level).Printf(format, v...)		
	}
}

//Log a message object with the INFO level. Arguments are handled in the manner of fmt.Print.
func (logger *Logger) Info(s string) {
	logger.Trace(INFO, s)
}

//Log a message object with the DEBUG level. Arguments are handled in the manner of fmt.Print.
func (logger *Logger) Debug(s string ) {
	logger.Trace(DEBUG, s)
}

//Log a message object with the ERROR level. Arguments are handled in the manner of fmt.Print.
func (logger *Logger) Error(s string ) {
	logger.Trace(ERROR, s)
}

//Log a message object with the FATAL level. Arguments are handled in the manner of fmt.Print.
func (logger *Logger) Fatal(s string ) {
	logger.Trace(FATAL, s)
}

//Log a message object with the WARN level. Arguments are handled in the manner of fmt.Print.
func (logger *Logger) Warn(s string) {
	logger.Trace(WARN, s)
}

//Log a message object with the INFO level. Arguments are handled in the manner of fmt.Printf.
func (logger *Logger) Infof(format string, v ...interface{} ) {
	logger.TraceF(INFO, format, v...)
}

//Log a message object with the DEBUG level. Arguments are handled in the manner of fmt.Printf.
func (logger *Logger) Debugf(format string, v ...interface{} ) {
	logger.TraceF(DEBUG, format, v...)
}

//Log a message object with the ERROR level. Arguments are handled in the manner of fmt.Printf.
func (logger *Logger) Errorf(format string, v ...interface{} ) {
	logger.TraceF(ERROR, format, v...)
}

//Log a message object with the FATAL level. Arguments are handled in the manner of fmt.Printf.
func (logger *Logger) Fatalf(format string, v ...interface{} ) {
	logger.TraceF(FATAL, format, v...)
}

//Log a message object with the WARN level. Arguments are handled in the manner of fmt.Printf.
func (logger *Logger) Warnf(format string, v ...interface{} ) {
	logger.TraceF(WARN, format, v...)
}
