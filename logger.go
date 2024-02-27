package raft

import "log"

type LoggerLevelType uint8

const (
	LoggerLevelDebug LoggerLevelType = iota
	LoggerLevelInfo
	LoggerLevelWarning
	LoggerLevelTest
	LoggerLevelError
	LoggerLevelPanic
	LoggerLevelExtra
)

type LoggerIFace interface {
	//Debug(v ...interface{})
	Debugf(format string, v ...interface{})

	//Info(v ...interface{})
	Infof(format string, v ...interface{})

	//Warning(v ...interface{})
	Warningf(format string, v ...interface{})

	//Error(v ...interface{})
	Errorf(format string, v ...interface{})

	//Panic(v ...interface{})
	Panicf(format string, v ...interface{})

	//Test(v ...interface{})
	Testf(format string, v ...interface{})

	//Extra(v ...interface{})
	Extraf(format string, v ...interface{})
}

type DefaultLogger struct {
	prefix string
}

func NewDefaultLogger(prefix string) *DefaultLogger {
	return &DefaultLogger{prefix: prefix}
}
func (d *DefaultLogger) doLog(level string, v ...interface{}) {

	log.Print(level, v)
}

func (d *DefaultLogger) doLogf(level string, format string, v ...interface{}) {
	format = level + " " + format
	log.Printf(format, v...)
}

//func (d *DefaultLogger) Debug(v ...interface{}) {
//	//TODO implement me
//	panic("implement me")
//}

func (d *DefaultLogger) Debugf(format string, v ...interface{}) {
	if LoggerLevelDebug >= LoggerLevel {
		format = d.prefix + " [Debug] " + format
		log.Printf(format, v...)
	}
}

//func (d *DefaultLogger) Info(v ...interface{}) {
//	//TODO implement me
//	panic("implement me")
//}

func (d *DefaultLogger) Infof(format string, v ...interface{}) {
	if LoggerLevelInfo >= LoggerLevel {
		format = d.prefix + " [Info] " + format
		log.Printf(format, v...)
	}
}

//func (d *DefaultLogger) Warning(v ...interface{}) {
//	//TODO implement me
//	panic("implement me")
//}

func (d *DefaultLogger) Warningf(format string, v ...interface{}) {
	if LoggerLevelWarning >= LoggerLevel {
		format = d.prefix + " [Warning] " + format
		log.Printf(format, v...)
	}
}

//func (d *DefaultLogger) Test(v ...interface{}) {
//	//TODO implement me
//	panic("implement me")
//}

func (d *DefaultLogger) Testf(format string, v ...interface{}) {

	if LoggerLevelTest >= LoggerLevel {
		format = d.prefix + " [Test] " + format
		log.Printf(format, v...)
	}
}

//func (d *DefaultLogger) Error(v ...interface{}) {
//	//TODO implement me
//	panic("implement me")
//}

func (d *DefaultLogger) Errorf(format string, v ...interface{}) {
	if LoggerLevelError >= LoggerLevel {
		format = d.prefix + " [Error] " + format
		log.Printf(format, v...)
	}
}

//func (d *DefaultLogger) Panic(v ...interface{}) {
//	//TODO implement me
//	panic("implement me")
//}

func (d *DefaultLogger) Panicf(format string, v ...interface{}) {

	if LoggerLevelPanic >= LoggerLevel {
		format = d.prefix + " [Panic] " + format
		log.Printf(format, v...)
	}
}

func (d *DefaultLogger) Extraf(format string, v ...interface{}) {

	if LoggerLevelExtra >= LoggerLevel {
		format = d.prefix + " [Extra] " + format
		log.Printf(format, v...)
	}
}
