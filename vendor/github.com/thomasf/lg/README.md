# lg
--
    import "github.com/thomasf/lg"

Go support for leveled logs

This is a fork of https://github.com/golang/glog with changes:

• Added support for colored levels/package highlighting.

• Added simple in memory logging.

• Added package pkg/promethues which has a prometheus line numbers per level
metric counters.

• Added support package pkg/lgexpire for expiering old lg/glog logs.

Copyright 2013 Google Inc. All Rights Reserved.

Package lg implements logging analogous to the Google-internal C++ INFO/ERROR/V
setup. It provides functions Info, Warning, Error, Fatal, plus formatting
variants such as Infof. It also provides V-style logging controlled by the -v
and -vmodule=file=2 flags.

Basic examples:

    lg.Info("Prepare to repel boarders")

    lg.Fatalf("Initialization failed: %s", err)

See the documentation for the V function for an explanation of these examples:

    if lg.V(2) {
    	lg.Info("Starting transaction...")
    }

    lg.V(2).Infoln("Processed", nItems, "elements")

## Usage

```go
var MaxSize uint64 = 1024 * 1024 * 1800
```
MaxSize is the maximum size of a log file in bytes.

```go
var Stats struct {
	Info, Warning, Error OutputStats
}
```
Stats tracks the number of lines of output and number of bytes per severity
level. Values must be read with atomic.LoadInt64.

#### func  CopyLoggerTo

```go
func CopyLoggerTo(name string, logger *stdLog.Logger)
```
CopyLoggerTo arranges for messages to be written to any log.Logger, see
CopyStandardLogTo for details on behaviour and details

#### func  CopyStandardLogTo

```go
func CopyStandardLogTo(name string)
```
CopyStandardLogTo arranges for messages written to the Go "log" package's
default logs to also appear in the Google logs for the named and lower
severities. Subsequent changes to the standard log's default output location or
format may break this behavior.

Valid names are "INFO", "WARNING", "ERROR", and "FATAL". If the name is not
recognized, CopyStandardLogTo panics.

#### func  Error

```go
func Error(args ...interface{})
```
Error logs to the ERROR, WARNING, and INFO logs. Arguments are handled in the
manner of fmt.Print; a newline is appended if missing.

#### func  ErrorDepth

```go
func ErrorDepth(depth int, args ...interface{})
```
ErrorDepth acts as Error but uses depth to determine which call frame to log.
ErrorDepth(0, "msg") is the same as Error("msg").

#### func  Errorf

```go
func Errorf(format string, args ...interface{})
```
Errorf logs to the ERROR, WARNING, and INFO logs. Arguments are handled in the
manner of fmt.Printf; a newline is appended if missing.

#### func  Errorln

```go
func Errorln(args ...interface{})
```
Errorln logs to the ERROR, WARNING, and INFO logs. Arguments are handled in the
manner of fmt.Println; a newline is appended if missing.

#### func  Exit

```go
func Exit(args ...interface{})
```
Exit logs to the FATAL, ERROR, WARNING, and INFO logs, then calls os.Exit(1).
Arguments are handled in the manner of fmt.Print; a newline is appended if
missing.

#### func  ExitDepth

```go
func ExitDepth(depth int, args ...interface{})
```
ExitDepth acts as Exit but uses depth to determine which call frame to log.
ExitDepth(0, "msg") is the same as Exit("msg").

#### func  Exitf

```go
func Exitf(format string, args ...interface{})
```
Exitf logs to the FATAL, ERROR, WARNING, and INFO logs, then calls os.Exit(1).
Arguments are handled in the manner of fmt.Printf; a newline is appended if
missing.

#### func  Exitln

```go
func Exitln(args ...interface{})
```
Exitln logs to the FATAL, ERROR, WARNING, and INFO logs, then calls os.Exit(1).

#### func  Fatal

```go
func Fatal(args ...interface{})
```
Fatal logs to the FATAL, ERROR, WARNING, and INFO logs, including a stack trace
of all running goroutines, then calls os.Exit(255). Arguments are handled in the
manner of fmt.Print; a newline is appended if missing.

#### func  FatalDepth

```go
func FatalDepth(depth int, args ...interface{})
```
FatalDepth acts as Fatal but uses depth to determine which call frame to log.
FatalDepth(0, "msg") is the same as Fatal("msg").

#### func  Fatalf

```go
func Fatalf(format string, args ...interface{})
```
Fatalf logs to the FATAL, ERROR, WARNING, and INFO logs, including a stack trace
of all running goroutines, then calls os.Exit(255). Arguments are handled in the
manner of fmt.Printf; a newline is appended if missing.

#### func  Fatalln

```go
func Fatalln(args ...interface{})
```
Fatalln logs to the FATAL, ERROR, WARNING, and INFO logs, including a stack
trace of all running goroutines, then calls os.Exit(255). Arguments are handled
in the manner of fmt.Println; a newline is appended if missing.

#### func  Flush

```go
func Flush()
```
Flush flushes all pending log I/O.

#### func  Info

```go
func Info(args ...interface{})
```
Info logs to the INFO log. Arguments are handled in the manner of fmt.Print; a
newline is appended if missing.

#### func  InfoDepth

```go
func InfoDepth(depth int, args ...interface{})
```
InfoDepth acts as Info but uses depth to determine which call frame to log.
InfoDepth(0, "msg") is the same as Info("msg").

#### func  Infof

```go
func Infof(format string, args ...interface{})
```
Infof logs to the INFO log. Arguments are handled in the manner of fmt.Printf; a
newline is appended if missing.

#### func  Infoln

```go
func Infoln(args ...interface{})
```
Infoln logs to the INFO log. Arguments are handled in the manner of fmt.Println;
a newline is appended if missing.

#### func  Memlog

```go
func Memlog() []string
```
Memlog returns the in memory log file

#### func  SetSrcHighlight

```go
func SetSrcHighlight(paths ...string)
```
SetSrcHighlight prepares colored stack trace highliting if -logtostderr and
-logcolor are enabled.

#### func  Warning

```go
func Warning(args ...interface{})
```
Warning logs to the WARNING and INFO logs. Arguments are handled in the manner
of fmt.Print; a newline is appended if missing.

#### func  WarningDepth

```go
func WarningDepth(depth int, args ...interface{})
```
WarningDepth acts as Warning but uses depth to determine which call frame to
log. WarningDepth(0, "msg") is the same as Warning("msg").

#### func  Warningf

```go
func Warningf(format string, args ...interface{})
```
Warningf logs to the WARNING and INFO logs. Arguments are handled in the manner
of fmt.Printf; a newline is appended if missing.

#### func  Warningln

```go
func Warningln(args ...interface{})
```
Warningln logs to the WARNING and INFO logs. Arguments are handled in the manner
of fmt.Println; a newline is appended if missing.

#### type Level

```go
type Level int32
```

Level specifies a level of verbosity for V logs. *Level implements flag.Value;
the -v flag is of type Level and should be modified only through the flag.Value
interface.

#### func  Verbosity

```go
func Verbosity() Level
```
Verbosity returns the current verbosity level.

#### func (*Level) Get

```go
func (l *Level) Get() interface{}
```
Get is part of the flag.Value interface.

#### func (*Level) Set

```go
func (l *Level) Set(value string) error
```
Set is part of the flag.Value interface.

#### func (*Level) String

```go
func (l *Level) String() string
```
String is part of the flag.Value interface.

#### type OutputStats

```go
type OutputStats struct {
}
```

OutputStats tracks the number of output lines and bytes written.

#### func (*OutputStats) Bytes

```go
func (s *OutputStats) Bytes() int64
```
Bytes returns the number of bytes written.

#### func (*OutputStats) Lines

```go
func (s *OutputStats) Lines() int64
```
Lines returns the number of lines written.

#### type Verbose

```go
type Verbose bool
```

Verbose is a boolean type that implements Infof (like Printf) etc. See the
documentation of V for more information.

#### func  V

```go
func V(level Level) Verbose
```
V reports whether verbosity at the call site is at least the requested level.
The returned value is a boolean of type Verbose, which implements Info, Infoln
and Infof. These methods will write to the Info log if called. Thus, one may
write either

    if lg.V(2) { lg.Info("log this") }

or

    lg.V(2).Info("log this")

The second form is shorter but the first is cheaper if logging is off because it
does not evaluate its arguments.

Whether an individual call to V generates a log record depends on the setting of
the -v and --vmodule flags; both are off by default. If the level in the call to
V is at least the value of -v, or of -vmodule for the source file containing the
call, the V call will log.

#### func (Verbose) Info

```go
func (v Verbose) Info(args ...interface{})
```
Info is equivalent to the global Info function, guarded by the value of v. See
the documentation of V for usage.

#### func (Verbose) Infof

```go
func (v Verbose) Infof(format string, args ...interface{})
```
Infof is equivalent to the global Infof function, guarded by the value of v. See
the documentation of V for usage.

#### func (Verbose) Infoln

```go
func (v Verbose) Infoln(args ...interface{})
```
Infoln is equivalent to the global Infoln function, guarded by the value of v.
See the documentation of V for usage.
