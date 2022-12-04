package zerolog

import (
	"bytes"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// LevelWriter defines as interface a writer may implement in order
// to receive level information with payload.
// Q&A(DONE): WriteLevel() 跟 zerolog.LevelWriter interface 是想干什么的？
// 强制约束：每一个 Writer 都得是按 level 进行输出的
// 首先你要注意，zerolog.Event 中的 w 就是一个 LevelWriter
// 这其实就要求了：通过 Event 触发的 log 输出，必须要判断 level 是否满足
//
// 如果你使用了 zerolog 提供的 API 进行日志输出，创建 Event 的时候，就已经判断过 level 了，
// 也就不需要在 WriteLevel() 判断一次 level 了
// 这也是为什么 zerolog.New() 函数中，直接将 io.Writer 利用 levelWriterAdapter 裹一下就完事了
// 而且 levelWriterAdapter.WriteLevel() 还什么事都不干。
//
// 但是 syslogWriter 不一样，syslog 的 level 跟 zerolog 的 level 是不一致的，
// 这时候就要利用这个 LevelWriter.WriteLevel() 进行 level 的对接
type LevelWriter interface {
	io.Writer
	WriteLevel(level Level, p []byte) (n int, err error)
}

// 辅助其他没有 impl zerolog.LevelWriter interface 的 io.Writer
type levelWriterAdapter struct {
	io.Writer
}

func (lw levelWriterAdapter) WriteLevel(l Level, p []byte) (n int, err error) {
	return lw.Write(p)
}

type syncWriter struct {
	mu sync.Mutex
	lw LevelWriter
}

// SyncWriter wraps w so that each call to Write is synchronized with a mutex.
// This syncer can be used to wrap the call to writer's Write method if it is
// not thread safe. Note that you do not need this wrapper for os.File Write
// operations on POSIX and Windows systems as they are already thread-safe.
func SyncWriter(w io.Writer) io.Writer {
	if lw, ok := w.(LevelWriter); ok {
		return &syncWriter{lw: lw}
	}
	return &syncWriter{lw: levelWriterAdapter{w}}
}

// Write implements the io.Writer interface.
func (s *syncWriter) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lw.Write(p)
}

// WriteLevel implements the LevelWriter interface.
func (s *syncWriter) WriteLevel(l Level, p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lw.WriteLevel(l, p)
}

type multiLevelWriter struct {
	writers []LevelWriter
}

func (t multiLevelWriter) Write(p []byte) (n int, err error) {
	for _, w := range t.writers {
		if _n, _err := w.Write(p); err == nil {
			n = _n
			if _err != nil {
				err = _err
			} else if _n != len(p) {
				err = io.ErrShortWrite
			}
		}
	}
	return n, err
}

func (t multiLevelWriter) WriteLevel(l Level, p []byte) (n int, err error) {
	for _, w := range t.writers {
		if _n, _err := w.WriteLevel(l, p); err == nil {
			n = _n
			if _err != nil {
				err = _err
			} else if _n != len(p) {
				err = io.ErrShortWrite
			}
		}
	}
	return n, err
}

// MultiLevelWriter creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command. If some writers
// implement LevelWriter, their WriteLevel method will be used instead of Write.
func MultiLevelWriter(writers ...io.Writer) LevelWriter {
	lwriters := make([]LevelWriter, 0, len(writers))
	for _, w := range writers {
		if lw, ok := w.(LevelWriter); ok {
			lwriters = append(lwriters, lw)
		} else {
			lwriters = append(lwriters, levelWriterAdapter{w})
		}
	}
	return multiLevelWriter{lwriters}
}

// TestingLog is the logging interface of testing.TB.
type TestingLog interface {
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Helper()
}

// TestWriter is a writer that writes to testing.TB.
type TestWriter struct {
	T TestingLog

	// Frame skips caller frames to capture the original file and line numbers.
	Frame int
}

// NewTestWriter creates a writer that logs to the testing.TB.
func NewTestWriter(t TestingLog) TestWriter {
	return TestWriter{T: t}
}

// Write to testing.TB.
func (t TestWriter) Write(p []byte) (n int, err error) {
	t.T.Helper()

	n = len(p)

	// Strip trailing newline because t.Log always adds one.
	p = bytes.TrimRight(p, "\n")

	// Try to correct the log file and line number to the caller.
	if t.Frame > 0 {
		_, origFile, origLine, _ := runtime.Caller(1)
		_, frameFile, frameLine, ok := runtime.Caller(1 + t.Frame)
		if ok {
			erase := strings.Repeat("\b", len(path.Base(origFile))+len(strconv.Itoa(origLine))+3)
			t.T.Logf("%s%s:%d: %s", erase, path.Base(frameFile), frameLine, p)
			return n, err
		}
	}
	t.T.Log(string(p))

	return n, err
}

// ConsoleTestWriter creates an option that correctly sets the file frame depth for testing.TB log.
func ConsoleTestWriter(t TestingLog) func(w *ConsoleWriter) {
	return func(w *ConsoleWriter) {
		w.Out = TestWriter{T: t, Frame: 6}
	}
}
