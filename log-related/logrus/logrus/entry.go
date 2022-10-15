package logrus

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (

	// qualified package name, cached at first use
	logrusPackage string

	// Positions in the call stack when tracing to report the calling method
	minimumCallerDepth int

	// Used for caller information initialisation
	callerInitOnce sync.Once
)

const (
	maximumCallerDepth int = 25
	knownLogrusFrames  int = 4
)

func init() {
	// start at the bottom of the stack before the package-name cache is primed
	minimumCallerDepth = 1
}

// Defines the key when adding errors using WithError.
var ErrorKey = "error"

// An entry is the final or intermediate Logrus logging entry. It contains all
// the fields passed with WithField{,s}. It's finally logged when Trace, Debug,
// Info, Warn, Error, Fatal or Panic is called on it. These objects can be
// reused and passed around as much as you wish to avoid field duplication.
// Entry 是实际进行日志记录的实体，而且这个 Entry 实体是跟具体的 Fields 绑定的
// 为了提高性能，这个 Entry 是可以重复利用，甚至是 pass around 的
type Entry struct {
	// 因为 Entry 并没有 io.Writer, 所以还是要记录自己属于哪一个 logrus.Logger 的
	Logger *Logger

	// Contains all the fields set by the user.
	// 跟着这个 Entry 走的数据（相当于 session 的唯一标识，每次都把它们打印出来）
	Data Fields

	// Time at which the log entry was created
	Time time.Time

	// Level the log entry was logged at: Trace, Debug, Info, Warn, Error, Fatal or Panic
	// This field will be set on entry firing and the value will be equal to the one in Logger struct field.
	Level Level

	// Calling method, with package name
	Caller *runtime.Frame

	// Message passed to Trace, Debug, Info, Warn, Error, Fatal or Panic
	Message string

	// When formatter is called in entry.log(), a Buffer may be set to entry
	Buffer *bytes.Buffer

	// Contains the context set by the user. Useful for hook processing etc.
	// 使用者往 hook 传递 session 信息的手段
	Context context.Context

	// err may contain a field formatting error
	err string
}

func NewEntry(logger *Logger) *Entry {
	return &Entry{
		Logger: logger,
		// Default is three fields, plus one optional.  Give a little extra room.
		Data: make(Fields, 6),
	}
}

// deep copy for logrus.Entry
func (entry *Entry) Dup() *Entry {
	data := make(Fields, len(entry.Data))
	for k, v := range entry.Data {
		data[k] = v
	}
	return &Entry{Logger: entry.Logger, Data: data, Time: entry.Time, Context: entry.Context, err: entry.err}
}

// Returns the bytes representation of this entry from the formatter.
func (entry *Entry) Bytes() ([]byte, error) {
	return entry.Logger.Formatter.Format(entry)
}

// Returns the string representation from the reader and ultimately the
// formatter.
func (entry *Entry) String() (string, error) {
	serialized, err := entry.Bytes()
	if err != nil {
		return "", err
	}
	str := string(serialized)
	return str, nil
}

// Add an error as single field (using the key defined in ErrorKey) to the Entry.
func (entry *Entry) WithError(err error) *Entry {
	return entry.WithField(ErrorKey, err)
}

// Add a context to the Entry.
func (entry *Entry) WithContext(ctx context.Context) *Entry {
	dataCopy := make(Fields, len(entry.Data))
	for k, v := range entry.Data {
		dataCopy[k] = v
	}
	return &Entry{Logger: entry.Logger, Data: dataCopy, Time: entry.Time, err: entry.err, Context: ctx}
}

// Add a single field to the Entry.
func (entry *Entry) WithField(key string, value interface{}) *Entry {
	return entry.WithFields(Fields{key: value})
}

// Add a map of fields to the Entry.
// 填充每一个 Entry 的 Data 这个 map（做到 struct 化的第一步）
// 返回 *Entry 是为了能够进行链式调用，方便后续追加 Message 之类的
func (entry *Entry) WithFields(fields Fields) *Entry {
	// Q&A(DONE): 为什么要创建一个临时 logrus.Fields map ? 而且还做了一个 deep copy
	// 因为这个函数返回的是一个新的 Entry，这个新的 Entry 是基于原始 receiver 的 Entry 的基础上
	// 再补充新的 param 而构成的
	// 所以整个函数的基调是：在原有的基础上，追加 fields param 带来的新变化
	// Q&A(DONE): 为什么要采用这种如此慢的方式每次都构建新的 Entry ？
	// 这其实就涉及到了，Entry 这玩意的定位问题:
	// 1. Entry 的存在，很大程度上是为了减少 reflact 的使用（因为 reflact 存在相当大的性能问题），
	//    而使用的一个中间层概念
	// 2. 鉴于 Go 擅长的是后端，基本上是基于 request-response 的模型构建的后端服务，
	//    那么每一次 request 只进行一次 reflact 解析，这个 request scope 内重复利用这个 Entry，
	//    就可以最大程度的规避 reflact 的性能问题了（类似于每一个 HttpHandler 就只解析一次 Context）
	// 3. 记录每个 session、request scope 必须要有的标识、跟踪信息（session-id、trace-id、ip-port、mac-address 之类的）
	//    这是在查看日志时，按 session 查看的必要信息。
	// Q&A(DONE): 为什么每次 data 都是重新分配且 deep copy？
	// 1. 因为 Entry 可能会被重复利用，但是又没有相应的 reset 操作，所以要重新创建 Fields map
	// 2. 避免 rehash，毕竟改动 Entry.data 的可能性还是很低的（显然一次性更新更好）
	// 3. 真正的原因是这个函数的旧注释：https://github.com/sirupsen/logrus/pull/1229
	//    This function is not declared with a pointer value because otherwise
	//    race conditions will occur when using multiple goroutines
	//    这个函数本来的 receiver 是变量，而不是指针，因为 logrus 采用 COW 的方案来避免并发问题！
	//    所以你才会看到总是分配新的 logrus.Fields, logrus.Entry
	data := make(Fields, len(entry.Data)+len(fields)) // cap = 旧的 + 新的
	for k, v := range entry.Data {
		// NOTE: copy on write to avoid race conditions
		// deep copy
		data[k] = v
	}

	/* check type and generate key and value as field */
	fieldErr := entry.err
	for k, v := range fields {
		isErrField := false
		if t := reflect.TypeOf(v); t != nil {
			// 不支持函数对象，也不支持函数指针
			// 因为这两个 printf 出来也是没有意义的
			switch {
			case t.Kind() == reflect.Func, t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Func:
				isErrField = true
			}
		}
		if isErrField {
			tmp := fmt.Sprintf("can not add field %q", k)
			if fieldErr != "" {
				fieldErr = entry.err + ", " + tmp
			} else {
				fieldErr = tmp
			}
		} else {
			// 其实甚至可以更粗暴，把这个 data 构建成 string
			// 然后 Entry 直接缓存 string 数据
			data[k] = v
		}
	}

	// 返回的是一个新的 Entry
	// 个人感觉这里创建的 Entry 不被 pool buf 起来的话，
	// 其实也就是暗示我们：不要满天飞的调用 WithFields() 函数，
	// 一个 request scope 调用一两次就好了
	// NOTE: 往现有 Entry 加字段的时候，要像 slice 那样，总是接住新的 Entry 来用
	return &Entry{Logger: entry.Logger, Data: data, Time: entry.Time, err: fieldErr, Context: entry.Context}
}

// Overrides the time of the Entry.
func (entry *Entry) WithTime(t time.Time) *Entry {
	dataCopy := make(Fields, len(entry.Data))
	for k, v := range entry.Data {
		dataCopy[k] = v
	}
	return &Entry{Logger: entry.Logger, Data: dataCopy, Time: t, err: entry.err, Context: entry.Context}
}

// getPackageName reduces a fully qualified function name to the package name
// There really ought to be to be a better way...
func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}

// getCaller retrieves the name of the first non-logrus calling function
func getCaller() *runtime.Frame {
	// cache this package's fully-qualified name
	callerInitOnce.Do(func() {
		pcs := make([]uintptr, maximumCallerDepth)
		_ = runtime.Callers(0, pcs)

		// dynamic get the package name and the minimum caller depth
		for i := 0; i < maximumCallerDepth; i++ {
			funcName := runtime.FuncForPC(pcs[i]).Name()
			if strings.Contains(funcName, "getCaller") {
				logrusPackage = getPackageName(funcName)
				break
			}
		}

		minimumCallerDepth = knownLogrusFrames
	})

	// Restrict the lookback frames to avoid runaway lookups
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := getPackageName(f.Function)

		// If the caller isn't part of this package, we're done
		if pkg != logrusPackage {
			return &f //nolint:scopelint
		}
	}

	// if we got here, we failed to find the caller's context
	return nil
}

// pass by value, 不带锁也是没有关系的？不算严格安全吧，但是一般不会动态改
// 就算改了，这里也是全部必要的字段都检查了，至少这时候是不会有问题的
// 实话说，logrus 里面只有 formatter 调用了这个函数，formatter 调用之前都是上锁的，所以没有并发问题
func (entry Entry) HasCaller() (has bool) {
	return entry.Logger != nil &&
		entry.Logger.ReportCaller &&
		entry.Caller != nil
}

// 正式输出
func (entry *Entry) log(level Level, msg string) {
	var buffer *bytes.Buffer

	// Q&A(DONE): 为何要 Dup() 出来？
	// 1. COW 确保并发安全（level 可能修改），而且 level 参数也是通过 atomic 读取出来的
	// 2. 下面读取 entry 字段的地方太多了，比起加锁，还不如直接 copy 出来
	// 3. 其实这时候 deep copy 出来一个 newEntry，更多是是利用这个 newEntry 来记录这一个瞬间的信息（时间、Message）
	//    同时这样可以尽可能减小加锁的范围，缓解并发导致的性能下降
	newEntry := entry.Dup()

	if newEntry.Time.IsZero() {
		newEntry.Time = time.Now()
	}

	newEntry.Level = level
	newEntry.Message = msg

	// Entry 本身并没有 fd 或者是输出方向，所以也不需要锁，而是使用 Logger 里面的锁
	newEntry.Logger.mu.Lock() // ReportCaller，bufPool 是上层 Logger 的公共资源，所以要加锁访问
	reportCaller := newEntry.Logger.ReportCaller
	bufPool := newEntry.getBufferPool()
	// 因为 bufPool 内部也有锁，所以可以解开 Logger.mu 这个大的资源锁
	// 另外，避免同时锁住多个不同的锁，是避免死锁的重要手段
	newEntry.Logger.mu.Unlock()

	if reportCaller {
		// 拿栈帧是会变慢的
		newEntry.Caller = getCaller()
	}

	newEntry.fireHooks() // 调用一下 callback 而已
	buffer = bufPool.Get()
	defer func() {
		newEntry.Buffer = nil
		buffer.Reset()      // 个人感觉跟下面的 buffer.Reset() 是重复的，有一个可以去掉
		bufPool.Put(buffer) // 减轻 GC 压力
	}()
	buffer.Reset()
	newEntry.Buffer = buffer

	newEntry.write()

	newEntry.Buffer = nil

	// To avoid Entry#log() returning a value that only would make sense for
	// panic() to use in Entry#Panic(), we avoid the allocation by checking
	// directly here.
	if level <= PanicLevel {
		panic(newEntry)
	}
}

func (entry *Entry) getBufferPool() (pool BufferPool) {
	if entry.Logger.BufferPool != nil {
		return entry.Logger.BufferPool
	}
	return bufferPool
}

func (entry *Entry) fireHooks() {
	var tmpHooks LevelHooks
	entry.Logger.mu.Lock() // 因为 Logger.Hooks 是共享资源，deep copy 出来就可以解锁了
	tmpHooks = make(LevelHooks, len(entry.Logger.Hooks))
	for k, v := range entry.Logger.Hooks {
		tmpHooks[k] = v
	}
	entry.Logger.mu.Unlock()

	err := tmpHooks.Fire(entry.Level, entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fire hook: %v\n", err)
	}
}

// 全程加锁输出日志
func (entry *Entry) write() {
	entry.Logger.mu.Lock()
	defer entry.Logger.mu.Unlock()
	// 格式化字符串
	// Q&A(DONE): 为什么这个也要加锁？
	// 因为 formatter 会访问 Logger 里面的资源，所以也得加锁。比如：entry.HasCaller()
	serialized, err := entry.Logger.Formatter.Format(entry) // JSONFormatter/TextFormatter
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to obtain reader, %v\n", err)
		return
	}
	// 正式输出（通过 io.Writer 输出内容而已，没有什么特别的）
	if _, err := entry.Logger.Out.Write(serialized); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log, %v\n", err)
	}
}

// Log will log a message at the level given as parameter.
// Warning: using Log at Panic or Fatal level will not respectively Panic nor Exit.
// For this behaviour Entry.Panic or Entry.Fatal should be used instead.
func (entry *Entry) Log(level Level, args ...interface{}) {
	if entry.Logger.IsLevelEnabled(level) { // 简单的显示开关
		entry.log(level, fmt.Sprint(args...)) // 先利用 fmt.Sprint 或者是 fmt.Sprintf，跟标准库的 log 思路是一样的
	}
}

func (entry *Entry) Trace(args ...interface{}) {
	entry.Log(TraceLevel, args...)
}

func (entry *Entry) Debug(args ...interface{}) {
	entry.Log(DebugLevel, args...)
}

func (entry *Entry) Print(args ...interface{}) {
	entry.Info(args...)
}

func (entry *Entry) Info(args ...interface{}) {
	entry.Log(InfoLevel, args...)
}

func (entry *Entry) Warn(args ...interface{}) {
	entry.Log(WarnLevel, args...)
}

func (entry *Entry) Warning(args ...interface{}) {
	entry.Warn(args...)
}

func (entry *Entry) Error(args ...interface{}) {
	entry.Log(ErrorLevel, args...)
}

func (entry *Entry) Fatal(args ...interface{}) {
	entry.Log(FatalLevel, args...)
	entry.Logger.Exit(1)
}

func (entry *Entry) Panic(args ...interface{}) {
	entry.Log(PanicLevel, args...)
}

// Entry Printf family functions

func (entry *Entry) Logf(level Level, format string, args ...interface{}) {
	if entry.Logger.IsLevelEnabled(level) {
		entry.Log(level, fmt.Sprintf(format, args...))
	}
}

func (entry *Entry) Tracef(format string, args ...interface{}) {
	entry.Logf(TraceLevel, format, args...)
}

func (entry *Entry) Debugf(format string, args ...interface{}) {
	entry.Logf(DebugLevel, format, args...)
}

func (entry *Entry) Infof(format string, args ...interface{}) {
	entry.Logf(InfoLevel, format, args...)
}

func (entry *Entry) Printf(format string, args ...interface{}) {
	entry.Infof(format, args...)
}

func (entry *Entry) Warnf(format string, args ...interface{}) {
	entry.Logf(WarnLevel, format, args...)
}

func (entry *Entry) Warningf(format string, args ...interface{}) {
	entry.Warnf(format, args...)
}

func (entry *Entry) Errorf(format string, args ...interface{}) {
	entry.Logf(ErrorLevel, format, args...)
}

func (entry *Entry) Fatalf(format string, args ...interface{}) {
	entry.Logf(FatalLevel, format, args...)
	entry.Logger.Exit(1)
}

func (entry *Entry) Panicf(format string, args ...interface{}) {
	entry.Logf(PanicLevel, format, args...)
}

// Entry Println family functions

func (entry *Entry) Logln(level Level, args ...interface{}) {
	if entry.Logger.IsLevelEnabled(level) {
		entry.Log(level, entry.sprintlnn(args...))
	}
}

func (entry *Entry) Traceln(args ...interface{}) {
	entry.Logln(TraceLevel, args...)
}

func (entry *Entry) Debugln(args ...interface{}) {
	entry.Logln(DebugLevel, args...)
}

func (entry *Entry) Infoln(args ...interface{}) {
	entry.Logln(InfoLevel, args...)
}

func (entry *Entry) Println(args ...interface{}) {
	entry.Infoln(args...)
}

func (entry *Entry) Warnln(args ...interface{}) {
	entry.Logln(WarnLevel, args...)
}

func (entry *Entry) Warningln(args ...interface{}) {
	entry.Warnln(args...)
}

func (entry *Entry) Errorln(args ...interface{}) {
	entry.Logln(ErrorLevel, args...)
}

func (entry *Entry) Fatalln(args ...interface{}) {
	entry.Logln(FatalLevel, args...)
	entry.Logger.Exit(1)
}

func (entry *Entry) Panicln(args ...interface{}) {
	entry.Logln(PanicLevel, args...)
}

// Sprintlnn => Sprint no newline. This is to get the behavior of how
// fmt.Sprintln where spaces are always added between operands, regardless of
// their type. Instead of vendoring the Sprintln implementation to spare a
// string allocation, we do the simplest thing.
func (entry *Entry) sprintlnn(args ...interface{}) string {
	msg := fmt.Sprintln(args...)
	return msg[:len(msg)-1]
}
