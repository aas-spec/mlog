// Logger with File Rotation 4.0
// Function:
// LogX, PanicX - Простая запись в лог
// LLogX - запись в лог с указанием в первом аргументе
// уровня логирования(Level)
// PrintX - аналог LogX, для совместимости и горячей замены log
// OutX - аналог предыдущих функций,
// но в качестве первого аргумента принимают LoggerID
// LOutx - второй аргумент Level (уровень логирования)
// SetLogLevel     - устанавливает для логгера уровень логирования
// SetStoreDays    - устанавливает для логгера кол-во дней ротации
// SetLogUseOwnDir - устанавливает для логгера параметр записи в свою папку

package mlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	DefStoreDays = 5     // Кол-во хранения файлов по дефолту
	DefLoggerID  = ""    // Идентификатор дефолтного логгера
	DefLevel     = 5     // Уровень логирования по дефолту
	DefUseOwnDir = false // У каждого логгера своя папка
)

// Описывает логгер
type TLogger struct {
	ID           string
	BaseFileName string
	log          *log.Logger
	lastFileName string
	StoreDays    int
	Level        int
	UseOwnDir    bool
}

func (ldata TLogger) getLogFileNameStr(s string) string {
	curPath := ldata.BaseFileName
	curPath = filepath.Dir(curPath) + string(os.PathSeparator) + filepath.Base(curPath)
	logfile := curPath[:len(curPath)-len(filepath.Ext(curPath))]
	logfile = logfile + "_" + s + ".log"
	return logfile
}

func (ldata TLogger) getLogFileName(tm time.Time) string {
	return ldata.getLogFileNameStr(tm.Format("2006-01-02"))
}

func (ldata *TLogger) checkLogRotation() (res string) {
	curLogFileName := ldata.getLogFileName(time.Now())
	if ldata.lastFileName != curLogFileName {
		curPath := ldata.BaseFileName
		curDir := filepath.Dir(curPath)
		curPath = curDir + string(os.PathSeparator) + filepath.Base(curPath)
		_, err := os.Stat(filepath.Dir(curPath))
		if err != nil && os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(curPath), 0777)
		}

		logFile, err := os.OpenFile(curLogFileName,
			os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			log.Panicf("Unable to open file %v : %s", curLogFileName, err)
		}

		ldata.lastFileName = curLogFileName
		ldata.log.SetFlags(0)

		if os.Getenv("DisableStdout") != "1" {
			ldata.log.SetOutput(io.MultiWriter(os.Stdout, logFile))
		} else {
			ldata.log.SetOutput(logFile)
		}

		// Очищаю старые файлы при смене имени лог файла
		files, err := filepath.Glob(ldata.getLogFileNameStr("????-??-??"))
		if err != nil {
			res += fmt.Sprintf("%s unable to expand mask: %s \n", GetTimeStamp(), err)
		}
		sort.Strings(files)
		for len(files) > ldata.StoreDays {
			fileToDelete := files[0]
			files = files[1:]
			res += fmt.Sprintf("%s delete: %s \n", GetTimeStamp(), fileToDelete)
			if err := os.Remove(fileToDelete); err != nil {
				res += fmt.Sprintf("%s unable to delete file: %s \n", GetTimeStamp(), err)
			}
		}
	}
	return res
}

func GetTimeStamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (L *TLogger) Logln(level int, v ...interface{}) {
	if level > L.Level {
		return
	}
	s := L.checkLogRotation()
	s += GetTimeStamp() + " " + fmt.Sprintln(v...)
	L.log.Print(s)
}

func (L *TLogger) Log(level int, v ...interface{}) {
	if level > L.Level {
		return
	}
	s := L.checkLogRotation()
	s += GetTimeStamp() + " " + fmt.Sprint(v...)
	L.log.Print(s)
}

func (L *TLogger) Logf(level int, format string, v ...interface{}) {
	if level > L.Level {
		return
	}
	s := L.checkLogRotation()
	s += GetTimeStamp() + " " + fmt.Sprintf(format, v...)
	L.log.Print(s)
}

func (L *TLogger) Panic(v ...interface{}) {
	s := GetTimeStamp() + " " + fmt.Sprint(v...)
	L.log.Panic(s)
}

// Возвращает базовое имя файла для логгера.
// Например /home/user/myapp → /home/user/logs/myapp-LogID.log
func getDefLoggerFileName(logID string, useOwnDir bool) string {
	curPath := os.Args[0]
	curDir := filepath.Dir(curPath)
	newPath := curDir + string(os.PathSeparator) + "logs" + string(os.PathSeparator)
	if useOwnDir && logID != "" {
		newPath += logID + string(os.PathSeparator)
	}
	curPath = newPath + filepath.Base(curPath)
	logfile := curPath[:len(curPath)-len(filepath.Ext(curPath))]
	if logID != "" && !useOwnDir {
		logfile += "-" + logID
	}
	logfile += ".log"
	return logfile
}

var defLogger = newLogger(DefLoggerID, getDefLoggerFileName(DefLoggerID, DefUseOwnDir), DefStoreDays, DefLevel, DefUseOwnDir)

// Тип для списка логгеров
type TLoggers struct {
	Sync  sync.Mutex
	Items map[string]TLogger
}

// Список логгеров
var loggers = TLoggers{
	Sync:  sync.Mutex{},
	Items: make(map[string]TLogger),
}

func newLogger(ID string, baseFileName string, storeDays int, level int, useOwnDir bool) TLogger {
	loggers.Sync.Lock()
	defer loggers.Sync.Unlock()
	ldata := TLogger{
		ID:           ID,
		BaseFileName: baseFileName,
		log:          log.New(os.Stdout, "", log.LstdFlags),
		StoreDays:    storeDays,
		Level:        level,
		UseOwnDir:    useOwnDir,
	}
	loggers.Items[ID] = ldata
	return ldata
}

func SetStoreDays(LoggerID string, StoreDays int) {
	loggers.Sync.Lock()
	logger, found := loggers.Items[LoggerID]
	if !found {
		loggers.Sync.Unlock()
		logger = newLogger(LoggerID, getDefLoggerFileName(LoggerID, DefUseOwnDir), StoreDays, DefLevel, DefUseOwnDir)
		loggers.Sync.Lock()
	} else {
		logger.StoreDays = StoreDays
	}
	loggers.Items[LoggerID] = logger
	if LoggerID == DefLoggerID {
		defLogger = logger
	}
	loggers.Sync.Unlock()
}

func SetLogLevel(LoggerID string, Level int) {
	loggers.Sync.Lock()
	logger, found := loggers.Items[LoggerID]
	if !found {
		loggers.Sync.Unlock()
		logger = newLogger(LoggerID, getDefLoggerFileName(LoggerID, DefUseOwnDir), DefStoreDays, Level, DefUseOwnDir)
		loggers.Sync.Lock()
	} else {
		logger.Level = Level
	}
	loggers.Items[LoggerID] = logger
	if LoggerID == DefLoggerID {
		defLogger = logger
	}
	loggers.Sync.Unlock()
}

func SetLogUseOwnDir(LoggerID string, useOwnDir bool) {
	loggers.Sync.Lock()
	logger, found := loggers.Items[LoggerID]
	if !found {
		loggers.Sync.Unlock()
		logger = newLogger(LoggerID, getDefLoggerFileName(LoggerID, useOwnDir), DefStoreDays, DefLevel, useOwnDir)
		loggers.Sync.Lock()
	} else {
		logger.UseOwnDir = useOwnDir
		logger.BaseFileName = getDefLoggerFileName(LoggerID, useOwnDir)
	}
	loggers.Items[LoggerID] = logger
	if LoggerID == DefLoggerID {
		defLogger = logger
	}
	loggers.Sync.Unlock()
}

func getLogger(loggerID string) TLogger {
	logger, found := loggers.Items[loggerID]
	if !found {
		logger = newLogger(loggerID, getDefLoggerFileName(loggerID, DefUseOwnDir), DefStoreDays, DefLevel, DefUseOwnDir)
	}
	return logger
}

///////////////////////////////////////////////////
// Функции для логирования в указанный LoggerID
///////////////////////////////////////////////////

func Outln(LoggerID string, v ...interface{}) {
	logger := getLogger(LoggerID)
	logger.Logln(0, v...)
}

func Out(LoggerID string, v ...interface{}) {
	logger := getLogger(LoggerID)
	logger.Log(0, v...)
}

func Outf(LoggerID string, format string, v ...interface{}) {
	logger := getLogger(LoggerID)
	logger.Logf(0, format, v...)
}

func LOutln(LoggerID string, Level int, v ...interface{}) {
	logger := getLogger(LoggerID)
	logger.Logln(Level, v...)
}

func LOut(LoggerID string, Level int, v ...interface{}) {
	logger := getLogger(LoggerID)
	logger.Log(Level, v...)
}

func LOutf(LoggerID string, Level int, format string, v ...interface{}) {
	logger := getLogger(LoggerID)
	logger.Logf(Level, format, v...)
}

///////////////////////////////////////////////////
// Функции используются с логгером по-умолчанию
///////////////////////////////////////////////////

func Logln(v ...interface{}) {
	defLogger.Logln(0, v...)
}

func Log(v ...interface{}) {
	defLogger.Log(0, v...)
}

func Logf(format string, v ...interface{}) {
	defLogger.Logf(0, format, v...)
}

func Panic(v ...interface{}) {
	defLogger.Panic(v...)
}

func LLogln(level int, v ...interface{}) {
	defLogger.Logln(level, v...)
}

func LLog(level int, v ...interface{}) {
	defLogger.Log(level, v...)
}

func LLogf(level int, format string, v ...interface{}) {
	defLogger.Logf(level, format, v...)
}

func Println(v ...interface{}) {
	defLogger.Logln(0, v...)
}

func Print(v ...interface{}) {
	defLogger.Log(0, v...)
}

func Printf(format string, v ...interface{}) {
	defLogger.Logf(0, format, v...)
}

func LPrintln(level int, v ...interface{}) {
	defLogger.Logln(level, v...)
}

func LPrint(level int, v ...interface{}) {
	defLogger.Log(level, v...)
}

func LPrintf(level int, format string, v ...interface{}) {
	defLogger.Logf(level, format, v...)
}

func StdPrintf(format string, v ...interface{}) {
	s := GetTimeStamp() + " " + fmt.Sprintf(format, v...)
	fmt.Println(s)
}
