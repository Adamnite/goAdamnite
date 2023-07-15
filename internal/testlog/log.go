package testlog

import (
	"sync"
	"testing"
)

// Handler returns a log handler which logs to the unit test log of t.
func Handler(t *testing.T, level log15.Lvl) log15.Handler {
	return log15.LvlFilterHandler(level, &handler{t, log15.TerminalFormat()})
}

type handler struct {
	t   *testing.T
	fmt log15.Format
}

func (h *handler) Log(r *log15.Record) error {
	h.t.Logf("%s", h.fmt.Format(r))
	return nil
}

// logger implements log.Logger such that all output goes to the unit test log via
// t.Logf(). All methods in between logger.Trace, logger.Debug, etc. are marked as test
// helpers, so the file and line number in unit test output correspond to the call site
// which emitted the log message.
type logger struct {
	t  *testing.T
	l  log15.Logger
	mu *sync.Mutex
	h  *bufHandler
}

type bufHandler struct {
	buf []*log15.Record
	fmt log15.Format
}

func (h *bufHandler) Log(r *log15.Record) error {
	h.buf = append(h.buf, r)
	return nil
}

// Logger returns a logger which logs to the unit test log of t.
func Logger(t *testing.T, level log15.Lvl) log15.Logger {
	l := &logger{
		t:  t,
		l:  log15.New(),
		mu: new(sync.Mutex),
		h:  &bufHandler{fmt: log15.TerminalFormat()},
	}
	l.l.SetHandler(log15.LvlFilterHandler(level, l.h))
	return l
}

func (l *logger) Trace(msg string, ctx ...interface{}) {
	l.t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.Trace(msg, ctx...)
	l.flush()
}

func (l *logger) Debug(msg string, ctx ...interface{}) {
	l.t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.Debug(msg, ctx...)
	l.flush()
}

func (l *logger) Info(msg string, ctx ...interface{}) {
	l.t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.Info(msg, ctx...)
	l.flush()
}

func (l *logger) Warn(msg string, ctx ...interface{}) {
	l.t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.Warn(msg, ctx...)
	l.flush()
}

func (l *logger) Error(msg string, ctx ...interface{}) {
	l.t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.Error(msg, ctx...)
	l.flush()
}

func (l *logger) Crit(msg string, ctx ...interface{}) {
	l.t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.l.Crit(msg, ctx...)
	l.flush()
}

func (l *logger) New(ctx ...interface{}) log15.Logger {
	return &logger{l.t, l.l.New(ctx...), l.mu, l.h}
}

func (l *logger) GetHandler() log15.Handler {
	return l.l.GetHandler()
}

func (l *logger) SetHandler(h log15.Handler) {
	l.l.SetHandler(h)
}

// flush writes all buffered messages and clears the buffer.
func (l *logger) flush() {
	l.t.Helper()
	for _, r := range l.h.buf {
		l.t.Logf("%s", l.h.fmt.Format(r))
	}
	l.h.buf = nil
}
