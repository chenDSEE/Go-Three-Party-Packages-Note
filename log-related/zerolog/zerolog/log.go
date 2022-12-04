// Package zerolog provides a lightweight logging library dedicated to JSON logging.
//
// A global Logger can be use for simple logging:
//
//     import "github.com/rs/zerolog/log"
//
//     log.Info().Msg("hello world")
//     // Output: {"time":1494567715,"level":"info","message":"hello world"}
//
// NOTE: To import the global logger, import the "log" subpackage "github.com/rs/zerolog/log".
//
// Fields can be added to log messages:
//
//     log.Info().Str("foo", "bar").Msg("hello world")
//     // Output: {"time":1494567715,"level":"info","message":"hello world","foo":"bar"}
//
// Create logger instance to manage different outputs:
//
//     logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
//     logger.Info().
//            Str("foo", "bar").
//            Msg("hello world")
//     // Output: {"time":1494567715,"level":"info","message":"hello world","foo":"bar"}
//
// Sub-loggers let you chain loggers with additional context:
//
//     sublogger := log.With().Str("component": "foo").Logger()
//     sublogger.Info().Msg("hello world")
//     // Output: {"time":1494567715,"level":"info","message":"hello world","component":"foo"}
//
// Level logging
//
//     zerolog.SetGlobalLevel(zerolog.InfoLevel)
//
//     log.Debug().Msg("filtered out message")
//     log.Info().Msg("routed message")
//
//     if e := log.Debug(); e.Enabled() {
//         // Compute log output only if enabled.
//         value := compute()
//         e.Str("foo": value).Msg("some debug message")
//     }
//     // Output: {"level":"info","time":1494567715,"routed message"}
//
// Customize automatic field names:
//
//     log.TimestampFieldName = "t"
//     log.LevelFieldName = "p"
//     log.MessageFieldName = "m"
//
//     log.Info().Msg("hello world")
//     // Output: {"t":1494567715,"p":"info","m":"hello world"}
//
// Log with no level and message:
//
//     log.Log().Str("foo","bar").Msg("")
//     // Output: {"time":1494567715,"foo":"bar"}
//
// Add contextual fields to global Logger:
//
//     log.Logger = log.With().Str("foo", "bar").Logger()
//
// Sample logs:
//
//     sampled := log.Sample(&zerolog.BasicSampler{N: 10})
//     sampled.Info().Msg("will be logged every 10 messages")
//
// Log with contextual hooks:
//
//     // Create the hook:
//     type SeverityHook struct{}
//
//     func (h SeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
//          if level != zerolog.NoLevel {
//              e.Str("severity", level.String())
//          }
//     }
//
//     // And use it:
//     var h SeverityHook
//     log := zerolog.New(os.Stdout).Hook(h)
//     log.Warn().Msg("")
//     // Output: {"level":"warn","severity":"warn"}
//
//
// Caveats
//
// There is no fields deduplication out-of-the-box.
// Using the same key multiple times creates new key in final JSON each time.
//
//     logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
//     logger.Info().
//            Timestamp().
//            Msg("dup")
//     // Output: {"level":"info","time":1494567715,"time":1494567715,"message":"dup"}
//
// In this case, many consumers will take the last value,
// but this is not guaranteed; check yours if in doubt.
package zerolog

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

// Level defines log levels.
type Level int8

const (
	// DebugLevel defines debug log level.
	DebugLevel Level = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel
	// NoLevel defines an absent log level.
	NoLevel
	// Disabled disables the logger.
	Disabled

	// TraceLevel defines trace log level.
	TraceLevel Level = -1
	// Values less than TraceLevel are handled as numbers.
)

func (l Level) String() string {
	switch l {
	case TraceLevel:
		return LevelTraceValue
	case DebugLevel:
		return LevelDebugValue
	case InfoLevel:
		return LevelInfoValue
	case WarnLevel:
		return LevelWarnValue
	case ErrorLevel:
		return LevelErrorValue
	case FatalLevel:
		return LevelFatalValue
	case PanicLevel:
		return LevelPanicValue
	case Disabled:
		return "disabled"
	case NoLevel:
		return ""
	}
	return strconv.Itoa(int(l))
}

// ParseLevel converts a level string into a zerolog Level value.
// returns an error if the input string does not match known values.
func ParseLevel(levelStr string) (Level, error) {
	switch levelStr {
	case LevelFieldMarshalFunc(TraceLevel):
		return TraceLevel, nil
	case LevelFieldMarshalFunc(DebugLevel):
		return DebugLevel, nil
	case LevelFieldMarshalFunc(InfoLevel):
		return InfoLevel, nil
	case LevelFieldMarshalFunc(WarnLevel):
		return WarnLevel, nil
	case LevelFieldMarshalFunc(ErrorLevel):
		return ErrorLevel, nil
	case LevelFieldMarshalFunc(FatalLevel):
		return FatalLevel, nil
	case LevelFieldMarshalFunc(PanicLevel):
		return PanicLevel, nil
	case LevelFieldMarshalFunc(Disabled):
		return Disabled, nil
	case LevelFieldMarshalFunc(NoLevel):
		return NoLevel, nil
	}
	i, err := strconv.Atoi(levelStr)
	if err != nil {
		return NoLevel, fmt.Errorf("Unknown Level String: '%s', defaulting to NoLevel", levelStr)
	}
	if i > 127 || i < -128 {
		return NoLevel, fmt.Errorf("Out-Of-Bounds Level: '%d', defaulting to NoLevel", i)
	}
	return Level(i), nil
}

// UnmarshalText implements encoding.TextUnmarshaler to allow for easy reading from toml/yaml/json formats
func (l *Level) UnmarshalText(text []byte) error {
	if l == nil {
		return errors.New("can't unmarshal a nil *Level")
	}
	var err error
	*l, err = ParseLevel(string(text))
	return err
}

// MarshalText implements encoding.TextMarshaler to allow for easy writing into toml/yaml/json formats
func (l Level) MarshalText() ([]byte, error) {
	return []byte(LevelFieldMarshalFunc(l)), nil
}

// A Logger represents an active logging object that generates lines
// of JSON output to an io.Writer. Each logging operation makes a single
// call to the Writer's Write method. 每一个 logging 动作，都会调用一次底层 io.Writer 的 Write() 一次
// There is no guarantee on access serialization to the Writer. If your Writer is not thread safe,
// you may consider a sync wrapper.
type Logger struct {
	w       LevelWriter // 日志输出
	level   Level
	sampler Sampler
	context []byte // 打包成数据数据的上下文信息(注意是 []byte)，并不是 Context package 的 Context
	hooks   []Hook
	stack   bool
}

// New creates a root logger with given output writer. If the output writer implements
// the LevelWriter interface, the WriteLevel method will be called instead of the Write
// one.
//
// 大部分通过 fd 来进行文件读写的其实都不需要担心这个并发问题，因为文件的读写本身的并发安全的
// 但是走 TCP 的 remote log 保存，就需要叠加一层 TCP socket buffer 确保 half-write 时的并发问题了
// Each logging operation makes a single call to the Writer's Write method. There is no
// guarantee on access serialization to the Writer. If your Writer is not thread safe,
// you may consider using sync wrapper.
// 将一个 io.Writer 裹成一个 Logger struct
func New(w io.Writer) Logger {
	if w == nil {
		// Discard is an io.Writer on which all Write calls succeed
		// without doing anything. 其实是直接 drop 的
		w = ioutil.Discard
	}
	lw, ok := w.(LevelWriter)
	if !ok {
		lw = levelWriterAdapter{w}
	}
	return Logger{w: lw, level: TraceLevel}
}

// Nop returns a disabled logger for which all operation are no-op.
func Nop() Logger {
	return New(nil).Level(Disabled)
}

// Output duplicates the current logger and sets w as its output.
func (l Logger) Output(w io.Writer) Logger {
	l2 := New(w)
	l2.level = l.level
	l2.sampler = l.sampler
	l2.stack = l.stack
	if len(l.hooks) > 0 {
		l2.hooks = append(l2.hooks, l.hooks...)
	}
	if l.context != nil {
		l2.context = make([]byte, len(l.context), cap(l.context))
		copy(l2.context, l.context)
	}
	return l2
}

// With creates a child logger with the field added to its context.
// 通过 zerolog.Context struct 对原本的 Logger 进行一次 deep copy
// 注意！这个函数采用的是 logger receiver，每次调用都影响原本的 Logger，原 Logger 总是 read-only 的
func (l Logger) With() Context {
	context := l.context
	l.context = make([]byte, 0, 500) // 包含 500 bytes 的一块内存 buffer
	if context != nil {
		l.context = append(l.context, context...)
	} else {
		// 第一次初始化
		// This is needed for AppendKey to not check len of input
		// thus making it inlinable
		l.context = enc.AppendBeginMarker(l.context)
	}
	return Context{l} // 新的 Logger，是原 Logger 的 deep copy 诞生出来的
}

// UpdateContext updates the internal logger's context.
//
// Use this method with caution. If unsure, prefer the With method.
func (l *Logger) UpdateContext(update func(c Context) Context) {
	if l == disabledLogger {
		return
	}
	if cap(l.context) == 0 {
		l.context = make([]byte, 0, 500)
	}
	if len(l.context) == 0 {
		l.context = enc.AppendBeginMarker(l.context)
	}
	c := update(Context{*l})
	l.context = c.l.context
}

// Level creates a child logger with the minimum accepted level set to level.
func (l Logger) Level(lvl Level) Logger {
	l.level = lvl
	return l
}

// GetLevel returns the current Level of l.
func (l Logger) GetLevel() Level {
	return l.level
}

// Sample returns a logger with the s sampler.
func (l Logger) Sample(s Sampler) Logger {
	l.sampler = s
	return l
}

// Hook returns a logger with the h Hook.
func (l Logger) Hook(h Hook) Logger {
	l.hooks = append(l.hooks, h)
	return l
}

// Trace starts a new message with trace level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Trace() *Event {
	return l.newEvent(TraceLevel, nil)
}

// Debug starts a new message with debug level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Debug() *Event {
	return l.newEvent(DebugLevel, nil)
}

// Info starts a new message with info level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Info() *Event {
	return l.newEvent(InfoLevel, nil)
}

// Warn starts a new message with warn level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Warn() *Event {
	return l.newEvent(WarnLevel, nil)
}

// Error starts a new message with error level.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Error() *Event {
	return l.newEvent(ErrorLevel, nil)
}

// Err starts a new message with error level with err as a field if not nil or
// with info level if err is nil.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Err(err error) *Event {
	if err != nil {
		return l.Error().Err(err)
	}

	return l.Info()
}

// Fatal starts a new message with fatal level. The os.Exit(1) function
// is called by the Msg method, which terminates the program immediately.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Fatal() *Event {
	return l.newEvent(FatalLevel, func(msg string) { os.Exit(1) })
}

// Panic starts a new message with panic level. The panic() function
// is called by the Msg method, which stops the ordinary flow of a goroutine.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Panic() *Event {
	return l.newEvent(PanicLevel, func(msg string) { panic(msg) })
}

// WithLevel starts a new message with level. Unlike Fatal and Panic
// methods, WithLevel does not terminate the program or stop the ordinary
// flow of a goroutine when used with their respective levels.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) WithLevel(level Level) *Event {
	switch level {
	case TraceLevel:
		return l.Trace()
	case DebugLevel:
		return l.Debug()
	case InfoLevel:
		return l.Info()
	case WarnLevel:
		return l.Warn()
	case ErrorLevel:
		return l.Error()
	case FatalLevel:
		return l.newEvent(FatalLevel, nil)
	case PanicLevel:
		return l.newEvent(PanicLevel, nil)
	case NoLevel:
		return l.Log()
	case Disabled:
		return nil
	default:
		return l.newEvent(level, nil)
	}
}

// Log starts a new message with no level. Setting GlobalLevel to Disabled
// will still disable events produced by this method.
//
// You must call Msg on the returned event in order to send the event.
func (l *Logger) Log() *Event {
	return l.newEvent(NoLevel, nil)
}

// Print sends a log event using debug level and no extra field.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Print(v ...interface{}) {
	if e := l.Debug(); e.Enabled() {
		e.CallerSkipFrame(1).Msg(fmt.Sprint(v...))
	}
}

// Printf sends a log event using debug level and no extra field.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, v ...interface{}) {
	if e := l.Debug(); e.Enabled() {
		e.CallerSkipFrame(1).Msg(fmt.Sprintf(format, v...))
	}
}

// Write implements the io.Writer interface. This is useful to set as a writer
// for the standard library log.
func (l Logger) Write(p []byte) (n int, err error) {
	n = len(p)
	if n > 0 && p[n-1] == '\n' {
		// Trim CR added by stdlog.
		p = p[0 : n-1]
	}
	l.Log().CallerSkipFrame(1).Msg(string(p))
	return
}

// Event 实际上是针对当前 Logger 进行了一次 snapshot
func (l *Logger) newEvent(level Level, done func(string)) *Event {
	enabled := l.should(level)
	if !enabled {
		// 该 level 不需要进行日志，或 log 功能是 disable 的
		if done != nil {
			done("")
		}
		return nil
	}
	e := newEvent(l.w, level)
	e.done = done
	e.ch = l.hooks
	if level != NoLevel && LevelFieldName != "" {
		e.Str(LevelFieldName, LevelFieldMarshalFunc(level))
	}
	if l.context != nil && len(l.context) > 1 {
		// 在 Event.buf 里面保存 Logger 中的 context 内容
		e.buf = enc.AppendObjectData(e.buf, l.context)
	}
	if l.stack {
		e.Stack()
	}
	return e
}

// should returns true if the log event should be logged.
func (l *Logger) should(lvl Level) bool {
	if lvl < l.level || lvl < GlobalLevel() {
		return false
	}
	if l.sampler != nil && !samplingDisabled() {
		return l.sampler.Sample(lvl)
	}
	return true
}
