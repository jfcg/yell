## yell [![Go Report Card](https://goreportcard.com/badge/github.com/jfcg/yell)](https://goreportcard.com/report/github.com/jfcg/yell) [![go.dev ref](https://raw.githubusercontent.com/jfcg/.github/main/godev.svg)](https://pkg.go.dev/github.com/jfcg/yell)
yell is yet another minimalist logging library. It comes with:
- four severity levels (info, warn, error, fatal)
- simple API
- [`io.Writer`](https://pkg.go.dev/io#Writer) & [`sync.Locker`](https://pkg.go.dev/sync#Locker) support
- package-specific loggers
- customizations (severity names, time format, local or UTC time)
- easy, granular request location (file.go:line) logging
- [semantic](https://semver.org) versioning

### Example
mypkg.go:
```go
package mypkg

import (
	"os"
	"github.com/jfcg/yell"
)

// log to stdout with warn or higher severity (for example).
var Logger = yell.New(": mypkg:", os.Stdout, yell.Swarn)

// Info tries to log message list with info severity
func Info(msg ...interface{}) error {
	return Logger.Log(yell.Sinfo, msg...)
}

// Warn tries to log message list with warn severity
func Warn(msg ...interface{}) error {
	return Logger.Log(yell.Swarn, msg...)
}

// Error tries to log message list with error severity
func Error(msg ...interface{}) (err error) {
	err = Logger.Log(yell.Serror, msg...)
	// extra stuff for error severity
	return
}

// Fatal tries to log message list with fatal severity and panics
func Fatal(msg ...interface{}) (err error) {
	err = Logger.Log(yell.Sfatal, msg...)
	pm := Logger.Name() + yell.Sname[yell.Sfatal]
	if err != nil {
		pm += err.Error()
	}
	// probably panic or os.Exit(1) in a fatal situation
	panic(pm)
}
```
myApp.go:
```go
package main

import (
	"fmt"
	"mypkg"
	"github.com/jfcg/yell"
)

func log() {
	defer func() {
		fmt.Println("recovering:", recover())
	}()

	// uses mypkg.Logger. yell records calling line in file.go:line format
	mypkg.Info("some info:", 1, "more")

	// uses yell.Default Logger, minimum severity is warning by default
	yell.Warn("some warning:", "few details")

	// record log() caller instead of this line
	mypkg.Error(yell.Caller(1), "bad error", 3.5, "data")

	// Fatal() logs & panicks
	yell.Fatal("fatal mistake", 2, "hard to recover")
}

func main() {
	// minimum severity for mypkg.Logger is warning, so ignored
	mypkg.Info("some info:", 3, "more")

	// set min severity level to info
	mypkg.Logger.SetLevel(yell.Sinfo)
	log()

	// yell library uses local time by default, to get coordinated universal time
	yell.UTC = true
	log()

	// change time format
	yell.TimeFormat = yell.TimeFormat[:19]
	log()

	// customized severity names (increasing severity)
	yell.Sname = [...]string{"信息:", "警告:", "错误:", "致命的:"}
	yell.UTC = false
	log()

	// disable logging for yell.Default
	yell.Default.SetLevel(yell.Snolog)
	log()
}
```
output:
```
2021-03-28 21:48:53.591948: mypkg:info: myApp.go:15: some info: 1 more
2021-03-28 21:48:53.592051: myApp:warn: myApp.go:18: some warning: few details
2021-03-28 21:48:53.592063: mypkg:error: myApp.go:33: bad error 3.5 data
2021-03-28 21:48:53.592082: myApp:fatal: myApp.go:24: fatal mistake 2 hard to recover
recovering: myApp:fatal:
2021-03-28 18:48:53.592100: mypkg:info: myApp.go:15: some info: 1 more
2021-03-28 18:48:53.592110: myApp:warn: myApp.go:18: some warning: few details
2021-03-28 18:48:53.592118: mypkg:error: myApp.go:37: bad error 3.5 data
2021-03-28 18:48:53.592126: myApp:fatal: myApp.go:24: fatal mistake 2 hard to recover
recovering: myApp:fatal:
2021-03-28 18:48:53: mypkg:info: myApp.go:15: some info: 1 more
2021-03-28 18:48:53: myApp:warn: myApp.go:18: some warning: few details
2021-03-28 18:48:53: mypkg:error: myApp.go:41: bad error 3.5 data
2021-03-28 18:48:53: myApp:fatal: myApp.go:24: fatal mistake 2 hard to recover
recovering: myApp:fatal:
2021-03-28 21:48:53: mypkg:信息: myApp.go:15: some info: 1 more
2021-03-28 21:48:53: myApp:警告: myApp.go:18: some warning: few details
2021-03-28 21:48:53: mypkg:错误: myApp.go:46: bad error 3.5 data
2021-03-28 21:48:53: myApp:致命的: myApp.go:24: fatal mistake 2 hard to recover
recovering: myApp:致命的:
2021-03-28 21:48:53: mypkg:信息: myApp.go:15: some info: 1 more
2021-03-28 21:48:53: mypkg:错误: myApp.go:50: bad error 3.5 data
recovering: myApp:致命的:
```

### Support
If you use yell and like it, please support via ETH:`0x464B840ee70bBe7962b90bD727Aac172Fa8B9C15`
