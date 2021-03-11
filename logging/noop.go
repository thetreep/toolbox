package logging

type Noop struct{}

func NoopLogger() LibLogger                             { return &Noop{} }
func (*Noop) Debugf(format string, args ...interface{}) {}
func (*Noop) Infof(format string, args ...interface{})  {}
func (*Noop) Printf(format string, args ...interface{}) {}
func (*Noop) Warnf(format string, args ...interface{})  {}
func (*Noop) Errorf(format string, args ...interface{}) {}
func (*Noop) Fatalf(format string, args ...interface{}) {}
func (*Noop) Panicf(format string, args ...interface{}) {}
func (*Noop) Debug(args ...interface{})                 {}
func (*Noop) Info(args ...interface{})                  {}
func (*Noop) Print(args ...interface{})                 {}
func (*Noop) Warn(args ...interface{})                  {}
func (*Noop) Warning(args ...interface{})               {}
func (*Noop) Error(args ...interface{})                 {}
func (*Noop) Fatal(args ...interface{})                 {}
func (*Noop) Panic(args ...interface{})                 {}
func (*Noop) Debugln(args ...interface{})               {}
