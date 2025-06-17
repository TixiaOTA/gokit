package logger

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/TixiaOTA/gokit/utils/env"
	"github.com/google/uuid"
)

type logger struct{}

var (
	Log  *logger
	once sync.Once
)

func init() {
	once.Do(func() {
		Log = &logger{}
	})
}

func (l *logger) Errorf(ctx context.Context, format string, args ...interface{}) {
	var (
		messages []LogMessage
		file     string
	)

	if ctx == nil {
		fmt.Printf("ERROR: %v (nil context)\n", fmt.Sprintf(format, args...))
		return
	}

	value, ok := extract(ctx)
	if !ok {
		fmt.Printf("ERROR: %v (logger not found in context)\n", fmt.Sprintf(format, args...))
		return
	}

	// for get filename and line when developer called this method
	_, fileName, line, _ := runtime.Caller(1)
	files := strings.Split(fileName, "/")
	if len(files) > 3 {
		file = fmt.Sprintf("%s:%d", strings.Join(files[len(files)-2:], "/"), line)
	} else {
		file = fmt.Sprintf("%s:%d", strings.Join(files, "/"), line)
	}

	tmp, ok := value.LoadAndDelete(_LogMessages)
	if ok && tmp != nil {
		existingMessages, ok := tmp.([]LogMessage)
		if ok {
			messages = existingMessages
		}
	}

	message := LogMessage{
		File:    file,
		Level:   err,
		Message: fmt.Sprintf(format, args...),
	}

	messages = append(messages, message)

	value.Set(_LogMessages, messages)
}

func (l *logger) Error(ctx context.Context, args ...interface{}) {
	var (
		messages []LogMessage
		file     string
	)

	if ctx == nil {
		fmt.Printf("ERROR: %v (nil context)\n", fmt.Sprint(args...))
		return
	}

	value, ok := extract(ctx)
	if !ok {
		fmt.Printf("ERROR: %v (logger not found in context)\n", fmt.Sprint(args...))
		return
	}

	// for get filename and line when developer called this method
	_, fileName, line, _ := runtime.Caller(1)
	files := strings.Split(fileName, "/")
	if len(files) > 3 {
		file = fmt.Sprintf("%s:%d", strings.Join(files[len(files)-2:], "/"), line)
	} else {
		file = fmt.Sprintf("%s:%d", strings.Join(files, "/"), line)
	}

	tmp, ok := value.LoadAndDelete(_LogMessages)
	if ok && tmp != nil {
		existingMessages, ok := tmp.([]LogMessage)
		if ok {
			messages = existingMessages
		}
	}

	message := LogMessage{
		File:    file,
		Level:   err,
		Message: fmt.Sprint(args...),
	}

	messages = append(messages, message)

	value.Set(_LogMessages, messages)
}

func (l *logger) DebugF(ctx context.Context, format string, args ...interface{}) {
	var (
		messages []LogMessage
		file     string
		appEnv   = strings.ToUpper(env.GetString("APP_ENV"))
	)

	// skip debug when app_env is production
	if !reflect.ValueOf(appEnv).IsZero() && appEnv == "PRODUCTION" {
		return
	}

	if ctx == nil {
		fmt.Printf("DEBUG: %v (nil context)\n", fmt.Sprintf(format, args...))
		return
	}

	value, ok := extract(ctx)
	if !ok {
		fmt.Printf("DEBUG: %v (logger not found in context)\n", fmt.Sprintf(format, args...))
		return
	}

	// for get filename and line when developer called this method
	_, fileName, line, _ := runtime.Caller(1)
	files := strings.Split(fileName, "/")
	if len(files) > 3 {
		file = fmt.Sprintf("%s:%d", strings.Join(files[len(files)-2:], "/"), line)
	} else {
		file = fmt.Sprintf("%s:%d", strings.Join(files, "/"), line)
	}

	tmp, ok := value.LoadAndDelete(_LogMessages)
	if ok && tmp != nil {
		existingMessages, ok := tmp.([]LogMessage)
		if ok {
			messages = existingMessages
		}
	}

	message := LogMessage{
		File:    file,
		Level:   debug,
		Message: fmt.Sprintf(format, args...),
	}

	messages = append(messages, message)

	value.Set(_LogMessages, messages)
}

func (l *logger) Debug(ctx context.Context, args ...interface{}) {
	var (
		messages []LogMessage
		file     string
		appEnv   = strings.ToUpper(env.GetString("APP_ENV"))
	)

	// skip debug when app_env is production
	if !reflect.ValueOf(appEnv).IsZero() && appEnv == "PRODUCTION" {
		return
	}

	if ctx == nil {
		fmt.Printf("DEBUG: %v (nil context)\n", fmt.Sprint(args...))
		return
	}

	value, ok := extract(ctx)
	if !ok {
		fmt.Printf("DEBUG: %v (logger not found in context)\n", fmt.Sprint(args...))
		return
	}

	// for get filename and line when developer called this method
	_, fileName, line, _ := runtime.Caller(1)
	files := strings.Split(fileName, "/")
	if len(files) > 3 {
		file = fmt.Sprintf("%s:%d", strings.Join(files[len(files)-2:], "/"), line)
	} else {
		file = fmt.Sprintf("%s:%d", strings.Join(files, "/"), line)
	}

	tmp, ok := value.LoadAndDelete(_LogMessages)
	if ok && tmp != nil {
		existingMessages, ok := tmp.([]LogMessage)
		if ok {
			messages = existingMessages
		}
	}

	message := LogMessage{
		File:    file,
		Level:   debug,
		Message: fmt.Sprint(args...),
	}

	messages = append(messages, message)

	value.Set(_LogMessages, messages)
}

func (l *logger) Printf(ctx context.Context, format string, args ...interface{}) {
	var (
		messages []LogMessage
		file     string
	)

	if ctx == nil {
		fmt.Printf("INFO: %v (nil context)\n", fmt.Sprintf(format, args...))
		return
	}

	value, ok := extract(ctx)
	if !ok {
		fmt.Printf("INFO: %v (logger not found in context)\n", fmt.Sprintf(format, args...))
		return
	}

	// for get filename and line when developer called this method
	_, fileName, line, _ := runtime.Caller(1)
	files := strings.Split(fileName, "/")
	if len(files) > 3 {
		file = fmt.Sprintf("%s:%d", strings.Join(files[len(files)-2:], "/"), line)
	} else {
		file = fmt.Sprintf("%s:%d", strings.Join(files, "/"), line)
	}

	tmp, ok := value.LoadAndDelete(_LogMessages)
	if ok && tmp != nil {
		existingMessages, ok := tmp.([]LogMessage)
		if ok {
			messages = existingMessages
		}
	}

	message := LogMessage{
		File:    file,
		Level:   print,
		Message: fmt.Sprintf(format, args...),
	}

	messages = append(messages, message)

	value.Set(_LogMessages, messages)
}

func (l *logger) Print(ctx context.Context, args ...interface{}) {
	var (
		messages []LogMessage
		file     string
	)

	if ctx == nil {
		fmt.Printf("INFO: %v (nil context)\n", fmt.Sprint(args...))
		return
	}

	value, ok := extract(ctx)
	if !ok {
		fmt.Printf("INFO: %v (logger not found in context)\n", fmt.Sprint(args...))
		return
	}

	// for get filename and line when developer called this method
	_, fileName, line, _ := runtime.Caller(1)
	files := strings.Split(fileName, "/")
	if len(files) > 3 {
		file = fmt.Sprintf("%s:%d", strings.Join(files[len(files)-2:], "/"), line)
	} else {
		file = fmt.Sprintf("%s:%d", strings.Join(files, "/"), line)
	}

	tmp, ok := value.LoadAndDelete(_LogMessages)
	if ok && tmp != nil {
		existingMessages, ok := tmp.([]LogMessage)
		if ok {
			messages = existingMessages
		}
	}

	message := LogMessage{
		File:    file,
		Level:   print,
		Message: fmt.Sprint(args...),
	}

	messages = append(messages, message)

	value.Set(_LogMessages, messages)
}

// GetRequestId getting request id log from context
func GetRequestId(ctx context.Context) string {
	if ctx == nil {
		return uuid.New().String()
	}

	value, ok := extract(ctx)
	if !ok {
		return uuid.New().String()
	}

	val, ok := value.Load(RequestId)
	if !ok || val == nil {
		return uuid.New().String()
	}

	v, ok := val.(string)
	if ok && v != "" {
		return v
	}

	return uuid.New().String()
}

func SetSaltKey(ctx context.Context, val string) {
	if ctx == nil {
		return
	}

	value, ok := extract(ctx)
	if !ok {
		return
	}

	if reflect.ValueOf(val).IsZero() {
		val = env.GetString("NEW_SALT_KEY")
	}

	value.Set(_SaltKey, val)
}

func GetSaltKey(ctx context.Context) string {
	defaultSaltKey := env.GetString("NEW_SALT_KEY")

	if ctx == nil {
		return defaultSaltKey
	}

	value, ok := extract(ctx)
	if !ok {
		return defaultSaltKey
	}

	val, ok := value.Load(_SaltKey)
	if !ok || val == nil {
		return defaultSaltKey
	}

	v, ok := val.(string)
	if ok && v != "" {
		return v
	}

	return defaultSaltKey
}

func Red(val interface{}) {
	fmt.Printf("\x1b[31;5m%v\x1b[0m\n", val)
}

func RedBold(val interface{}) {
	fmt.Printf("\x1b[31;1m%v\x1b[0m\n", val)
}

func Green(val interface{}) {
	fmt.Printf("\x1b[32;5m%v\x1b[0m\n", val)
}

func GreenBold(val interface{}) {
	fmt.Printf("\x1b[32;1m%v\x1b[0m\n", val)
}

func Yellow(val interface{}) {
	fmt.Printf("\x1b[33;5m%v\x1b[0m\n", val)
}

func YellowBold(val interface{}) {
	fmt.Printf("\x1b[33;1m%v\x1b[0m\n", val)
}

func Purple(val interface{}) {
	fmt.Printf("\x1b[35;5m%v\x1b[0m\n", val)
}

func PurpleBold(val interface{}) {
	fmt.Printf("\x1b[35;1m%v\x1b[0m\n", val)
}

func Blue(val interface{}) {
	fmt.Printf("\x1b[36;5m%v\x1b[0m\n", val)
}

func BlueBold(val interface{}) {
	fmt.Printf("\x1b[36;1m%v\x1b[0m\n", val)
}
