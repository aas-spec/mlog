// Logger with File Rotation 3.0
// Function:
// LogX, PanicX - Простая запись в лог
// LLogX - запись в лог с указанием в первом аргументе
// уровня логирования(Level)
// PrintX - аналог LogX, для совместимости и горячей замены log
// OutX - аналог предыдущих функций,
// но в качестве первого аргумента принимают LoggerID
// LOutx - второй аргумент Level (уровень логирования)
// SetLogLevel  - устанавливает для логгера уровень логирования
// SetStoreDays - устанавливает для логгера кол-во дней ротации

package mlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	DefStoreDays = 5  // Кол-во хранения файлов по дефолту
	DefLoggerID  = "" // Идентификатор дефолтного логгера
	DefLevel     = 5  // Уровень логирования по дефолту
)

// Описывает логгер
type TLogger struct {
	ID           string
	BaseFileName string
	log          *log.Logger
	lastFileName string
	StoreDays    int
	Level        int
}

func (ldata TLogger) getLogFileName(tm time.Time, Mask bool) string {
	curpath := ldata.BaseFileName
	if Mask {
		curpath = getDefLoggerFileName("")
	}
	curpath = filepath.Dir(curpath) + string(os.PathSeparator) + filepath.Base(curpath)
	logfile := curpath[:len(curpath)-len(filepath.Ext(curpath))]
	if Mask {
		logfile += "*"
	}
	logfile = logfile + "-" + tm.Format("2006-01-02") + ".log"

	return logfile
}

func (ldata *TLogger) checkLogRotation() {
	CurLogFileName := ldata.getLogFileName(time.Now(), false)
	if ldata.lastFileName != CurLogFileName { // Переоткрыть новый файл
		// Проверяю и создаю каталог для записи
		curpath := ldata.BaseFileName
		curdir := filepath.Dir(curpath)
		curpath = curdir + string(os.PathSeparator) + filepath.Base(curpath)
		_, err := os.Stat(filepath.Dir(curpath))
		if err != nil && os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(curpath), 0777)
		}
		// Очищаю старые фалйы при смене имени лог файла
		for i := 0; i < 7; i++ {
			logfilename := ldata.getLogFileName(time.Now().AddDate(0, 0,
				-ldata.StoreDays-i), true)
			files, err := filepath.Glob(logfilename)
			if err != nil {
				Logln(err)
			}
			for _, f := range files {
				log.Println("Delete: " + f)
				if err := os.Remove(f); err != nil {
					Logln(err)
				}
			}
		}
		// Открываю Log
		logFile, err := os.OpenFile(CurLogFileName,
			os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			log.Panicf("Unable to open file %v : %s", CurLogFileName, err)
		}
		ldata.lastFileName = CurLogFileName
		mw := io.MultiWriter(os.Stdout, logFile)
		ldata.log.SetFlags(0)
		ldata.log.SetOutput(mw)
	}
}

func getTimeStamp() string {
	tm := time.Now()
	res := tm.Format("2006-01-02 15:04:05")
	return res
}

func (L *TLogger) Logln(Level int, v ...interface{}) {

	if Level > L.Level {
		return
	}
	L.checkLogRotation()
	s := getTimeStamp() + " " + fmt.Sprintln(v...)
	L.log.Print(s)
}

func (L *TLogger) Log(Level int, v ...interface{}) {
	if Level > L.Level {
		return
	}
	L.checkLogRotation()
	s := getTimeStamp() + " " + fmt.Sprint(v...)
	L.log.Print(s)
}

func (L *TLogger) Logf(Level int, format string, v ...interface{}) {
	if Level > L.Level {
		return
	}
	L.checkLogRotation()
	s := getTimeStamp() + " " + fmt.Sprintf(format, v...)
	L.log.Print(s)
}

func (L *TLogger) Panic(v ...interface{}) {
	L.checkLogRotation()
	s := getTimeStamp() + " " + fmt.Sprint(v...)
	L.log.Panic(s)
}

// Возвращает базовое имя файла для логгера по дефолту
// Например c:\Blabla\My\My.exe
// Лог будет c:\Blabla\My\logs\My-LogID.log
func getDefLoggerFileName(LogID string) string {
	curpath := os.Args[0]
	curdir := filepath.Dir(curpath)
	curpath = curdir + string(os.PathSeparator) + "logs" + string(os.PathSeparator) + filepath.Base(curpath)
	logfile := curpath[:len(curpath)-len(filepath.Ext(curpath))]
	if LogID != "" {
		logfile += "-" + LogID
	}
	logfile += ".log"
	return logfile
}

var defLogger = newLogger(DefLoggerID, getDefLoggerFileName(DefLoggerID), DefStoreDays, DefLevel)

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

///////////////////////////////////////////////////
// Добавляет новый логгер
///////////////////////////////////////////////////
func newLogger(ID string, BaseFileName string, StoreDays int, Level int) TLogger {
	loggers.Sync.Lock()
	defer func() {
		loggers.Sync.Unlock()
	}()
	ldata := TLogger{
		ID:           ID,
		BaseFileName: BaseFileName,
		log:          log.New(os.Stdout, "", log.LstdFlags),
		StoreDays:    StoreDays,
		Level:        Level,
	}
	loggers.Items[ID] = ldata
	return ldata
}

func SetStoreDays(LoggerID string, StoreDays int) {
	loggers.Sync.Lock()
	defer func() {
		loggers.Sync.Unlock()
	}()
	logger, found := loggers.Items[LoggerID]
	if !found {
		logger = newLogger(LoggerID, getDefLoggerFileName(LoggerID), StoreDays, DefLevel)
	} else {
		logger.StoreDays = StoreDays
	}
	loggers.Items[LoggerID] = logger
	if LoggerID == DefLoggerID {
		defLogger = logger
	}
}

func SetLogLevel(LoggerID string, Level int) {
	loggers.Sync.Lock()
	defer func() {
		loggers.Sync.Unlock()
	}()
	logger, found := loggers.Items[LoggerID]
	if !found {
		logger = newLogger(LoggerID, getDefLoggerFileName(LoggerID), DefStoreDays, Level)
	} else {
		logger.Level = Level
	}
	loggers.Items[LoggerID] = logger
	if LoggerID == DefLoggerID {
		defLogger = logger
	}
}

///////////////////////////////////////////////////
// Функции для логирования в указанный LoggerID
///////////////////////////////////////////////////

func Outln(LoggerID string, v ...interface{}) {
	logger, found := loggers.Items[LoggerID]
	if !found {
		logger = newLogger(LoggerID, getDefLoggerFileName(LoggerID), DefStoreDays, DefLevel)
	}
	logger.Logln(0, v...)
}

func Out(LoggerID string, v ...interface{}) {
	logger, found := loggers.Items[LoggerID]
	if !found {
		logger = newLogger(LoggerID, getDefLoggerFileName(LoggerID), DefStoreDays, DefLevel)
	}
	logger.Log(0, v...)
}

func Outf(LoggerID string, fmt string, v ...interface{}) {
	logger, found := loggers.Items[LoggerID]
	if !found {
		logger = newLogger(LoggerID, getDefLoggerFileName(LoggerID), DefStoreDays, DefLevel)
	}
	logger.Logf(0, fmt, v...)
}

func LOutln(LoggerID string, Level int, v ...interface{}) {
	logger, found := loggers.Items[LoggerID]
	if !found {
		logger = newLogger(LoggerID, getDefLoggerFileName(LoggerID), DefStoreDays, DefLevel)
	}
	logger.Logln(Level, v...)
}

func LOut(LoggerID string, Level int, v ...interface{}) {
	logger, found := loggers.Items[LoggerID]
	if !found {
		logger = newLogger(LoggerID, getDefLoggerFileName(LoggerID), DefStoreDays, DefLevel)
	}
	logger.Log(Level, v...)
}

func LOutf(LoggerID string, Level int, fmt string, v ...interface{}) {
	logger, found := loggers.Items[LoggerID]
	if !found {
		logger = newLogger(LoggerID, getDefLoggerFileName(LoggerID), DefStoreDays, DefLevel)
	}
	logger.Logf(Level, fmt, v...)
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
func Logf(fmt string, v ...interface{}) {
	defLogger.Logf(0, fmt, v...)
}

func Panic(v ...interface{}) {
	defLogger.Panic(v)
}
func LLogln(level int, v ...interface{}) {
	defLogger.Logln(level, v...)
}

func LLog(level int, v ...interface{}) {
	defLogger.Log(level, v...)
}
func LLogf(level int, fmt string, v ...interface{}) {
	defLogger.Logf(level, fmt, v...)
}

func Println(v ...interface{}) {
	defLogger.Logln(0, v...)
}

func Print(v ...interface{}) {
	defLogger.Log(0, v...)
}
func Printf(fmt string, v ...interface{}) {
	defLogger.Logf(0, fmt, v...)
}

func LPrintln(level int, v ...interface{}) {
	defLogger.Logln(level, v...)
}

func LPrint(level int, v ...interface{}) {
	defLogger.Log(level, v...)
}
func LPrintf(level int, fmt string, v ...interface{}) {
	defLogger.Logf(level, fmt, v...)
}
