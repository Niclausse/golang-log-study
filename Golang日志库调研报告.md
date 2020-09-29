# Golang业务系统日志库调研报告



| 调研模块             | 调研时间/作者     | 更新时间/作者     |
| -------------------- | ----------------- | ----------------- |
| built-in log package | 2020-09-28 / 林鹏 | 2020-09-28 / 林鹏 |
| glog                 | 2020-09-28 / 林鹏 | 2020-09-28 / 林鹏 |
| logrus               | 2020-09-28 / 林鹏 | 2020-09-28 / 林鹏 |
| zap                  | 2020-09-28 / 林鹏 | 2020-09-28 / 林鹏 |
| golog                | 2020-09-28 / 林鹏 | 2020-09-28 / 林鹏 |



## 一、Golang Built-In log package

Go标准库log提供了最基础的日志功能。没有提供日志级别，但是它为用户提供了构建日志策略所需的最基本的属性。

### 简单使用

```go
package main

import "log"

func main() {
  // 向标准输出打印，日志信息包含日期和时间（可以通过日期过滤日志信息）
  log.Println("Hello world!")
}

// 2020/09/28 10:36:32 Hello world!  控制台打印信息
```



默认情况下，log package将日志信息打印到标准输出中，但是可以配置输出到任意实现了`io.Writer`接口的位置中，例如输出到文件中。



### 日志输出到文件中

```go
package main

import (
	"log"
  "os"
)

func main() {
  // if the file doesn't exist, create it or append to the file
  file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
  if err != nil {
    log.Fatal(err)
  }
  
  log.SetOutput(file) // 将日志信息输出到文件中
  
  log.Println("Hello world!")
}

// 2020/09/28 10:36:32 Hello world!  输出到文件中的日志信息
```



### 基于log标准库，构建自定义logger

```go
package main

import (
	"log"
	"os"
)

var (
  // 定义不同级别的日志类型
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func init() {
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

  // 为每种日志类型定义输出格式：输出位置，日志前缀信息，提示信息：日期 ｜ 时间 ｜ src文件位置及行号
  // 报告文件位置及行号，调用runtime.Caller()性能消耗大
	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Llongfile)  // 可以将不同级别的日志输出不同文件中
	WarningLogger = log.New(file, "WARNING", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	InfoLogger.Println("Starting the application...")
	InfoLogger.Println("Something noteworthy happened")
	WarningLogger.Println("There is something you should know about")
	ErrorLogger.Println("Something went wrong")
}
```



新构建的Info，warning，error三种级别的日志类型；对日志格式设置了前缀信息，时间，行号等附加信息。以上代码的日志输出如下：

```zsh
INFO: 2020/09/28 10:36:32 main.go:26: Starting the application...
INFO: 2020/09/28 10:36:32 main.go:27: Something noteworthy happened
WARNING2020/09/28 10:36:32 main.go:28: There is something you should know about
ERROR2020/09/28 10:36:32 main.go:29: Something went wrong
```



### 性能消耗

runtime.Caller()能够拿到当前执行的文件名和行号，这个方法几乎所有的日志组件都有使用。

```go
// Caller reports file and line number information about function invocations on
// the calling goroutine's stack. The argument skip is the number of stack frames
// to ascend, with 0 identifying the caller of Caller.  (For historical reasons the
// meaning of skip differs between Caller and Callers.) The return values report the
// program counter, file name, and line number within the file of the corresponding
// call. The boolean ok is false if it was not possible to recover the information.
func Caller(skip int) (pc uintptr, file string, line int, ok bool) {
	rpc := make([]uintptr, 1)
	n := callers(skip+1, rpc[:])
	if n < 1 {
		return
	}
	frame, _ := CallersFrames(rpc).Next()
	return frame.PC, frame.File, frame.Line, frame.PC != 0
}
```

这个函数开销很大。主要通过不停的迭代来跟踪到执行文件及行号，迭代过程单次消耗时间可以忽略不计，但是对于日志量巨大的服务而言影响还是很大的。



## 二、Google/Glog

Golfing/glog是C++版本google/glog的Go版本实现，实现了分级执行日志。 在kubernetes中，glog是默认的日志库。



### Overview

* glog将日志级别分为四种，分别是：INFO、WARNING、ERROR、FATAL，打印完日志后程序将会推出（os.Exit()）。

* 每个日志等级对应一个日志文件，低等级的日志文件除了包含该等级的日志，还会包含高等级的日志。

* 日志文件可以根据大小切割，但不能根据日期切割。

* 日志文件名称格式：program.host.username.log.log_level.date-time.pid，不可自定义。

* 固定日志输出格式：Lmmdd hh:mm:ss.uuuuuu pid file:line] msg，不可自定义。



### 简单使用

```go
package main

import (
	"flag"
	"github.com/golang/glog"
)

func main() {
	flag.Parse()
	defer glog.Flush() // flush all pending log I/O

	glog.Infoln("This is info message")
	glog.Infof("This is info message: %+v", 12345)

	glog.Warningln("This is warning message")
	glog.Warningf("This is warning message: %+v", 12345)

	glog.Errorln("This is error message")
	glog.Errorf("This is error message: %+v", 12345)

	glog.Fatal("This is fatal error")
}
```

运行程序：

```zsh
$ mkdir -p log && go run main.go -log_dir=log -alsologtostderr
```

以上打印日志将会同时打印在`log/`目录和标准错误输出中（`-alsologtostderr`）



## 三、结构化日志框架Logrus

### Overview

* 完全兼容Golang标准库日志模块，因此对于使用`log`标准库的项目，可以使用`log "github.com/sirupsen/logrus"`直接替换即可。
* logrus拥有七种日志级别：trace、debug、info、warn、error、fatal、panic，相当于标准库`log`的超集。
* 可扩展的Hook机制“允许使用者通过hook的方式将日志分发到任意地方，如本地文件系统、标准输出、logstash、elasticsearch或者mq等；或者通过hook定义日志内容和格式等。
* 可选的日志输出格式：logrus内置了两种日志格式。JSONFormatter和TextFormatter，如果这两个格式都不满足要求，可以通过实现Formatter接口，自定义日志格式。目前logrus支持的第三方日志格式也比较多，如FluentdFormatter、GELF、logstash等。
* Field机制：logrus鼓励通过Field机制进行精细化的、结构化的日志记录，而不是通过冗长的消息来记录日志。



### 简单使用

```go
package main

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{}) // 日志格式

	log.SetOutput(os.Stdout) // 重定向输出

	log.SetLevel(log.TraceLevel) // 日志显示级别

	// 报告被调用位置（文件及行号）!!! 开启此功能对性能开销大
	// 考虑通过Hook的方式定制需要此功能的日志级别；比如对ERROR级别及以上的日志报告文件名及行号
	log.SetReportCaller(true)
}


func main() {

	log.Traceln("Trace message")

	log.Debugln("Debug message")

	log.Println("Info message by Print func")
	log.Infoln("Info message by Info func")

	log.Warningln("Warning message")

	log.Errorln("Error message")

}
```

TextFormatter和JSONFormatter标准错误输出分别如下：

```zsh
TRAC[0000]/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-logrus/main.go:12 main.main() Trace message
DEBU[0000]/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-logrus/main.go:14 main.main() Debug message
INFO[0000]/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-logrus/main.go:16 main.main() Info message by Print func
INFO[0000]/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-logrus/main.go:17 main.main() Info message by Info func
WARN[0000]/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-logrus/main.go:19 main.main() Warning message
ERRO[0000]/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-logrus/main.go:21 main.main() Error message
```

```json
{"file":"/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-logrus/main.go:13","func":"main.main","level":"trace","msg":"Trace message","time":"2020-09-28T15:22:39+08:00"}
{"file":"/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-logrus/main.go:15","func":"main.main","level":"debug","msg":"Debug message","time":"2020-09-28T15:22:39+08:00"}
{"file":"/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-logrus/main.go:17","func":"main.main","level":"info","msg":"Info message by Print func","time":"2020-09-28T15:22:39+08:00"}
{"file":"/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-logrus/main.go:18","func":"main.main","level":"info","msg":"Info message by Info func","time":"2020-09-28T15:22:39+08:00"}
{"file":"/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-logrus/main.go:20","func":"main.main","level":"warning","msg":"Warning message","time":"2020-09-28T15:22:39+08:00"}
{"file":"/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-logrus/main.go:22","func":"main.main","level":"error","msg":"Error message","time":"2020-09-28T15:22:39+08:00"}

```



### FieldS用法

logrus不推荐使用冗长的消息来记录运行信息，它推荐使用Fields来进行精细化的、结构化的信息记录。

例如下面的记录日志的方式：

```go
log.Errorf("Failed to send event %s to topic %s with key %d", event, topic, key)
```

在logrus中鼓励使用以下方式替代：

```go
	log.WithFields(log.Fields{
		"event": event,
		"topic": topic,
		"key": key,
	}).Errorf("Failed to send event")
```



通常,在一个应用中、或者应用的一部分中，都有一些固定的Field。比如在处理用户http请求时，上下文中，所有的日志都会有request_id和user_ip。为了避免每次记录日志都要使用log.WithFields(log.Fields{“request_id”: request_id, “user_ip”: user_ip})，我们可以创建一个logrus.Entry实例,为这个实例设置默认Fields,在上下文中使用这个logrus.Entry实例记录日志即可。

```go
	logEntry := log.WithFields(log.Fields{"request_id": requestID, "user_ip": userIP})
	logEntry.Infoln("Something happened on that request")
```



## 四、结构化日志框架Uber/Zap

### Overview

Blazing fast, structure, leveled logging in Go

根据Uber-go Zap的文档，它的性能比类似的结构化日志库更好，也比标准库更快。可参看<a href="https://github.com/uber-go/zap">基准测试的对比表</a>。

### 简单使用

```go
package main

import (
	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
  // zap提供了三种默认配置，分别对应NewExample() / NewDevelopment() / NewProduction()三个方法
	log = zap.NewExample()
	// log, _ = zap.NewDevelopment()
	// log, _ = zap.NewProduction()
	defer log.Sync() // flush buffers
}

func main() {

	log.Debug("Debug message")
	log.Info("Info message")
	log.Warn("Warning message")
	log.Error("Error message")

	//log.Panic("Panic message")
	//log.Fatal("Fatal message")
}

```

标准输出日志如下：

```zsh
# NewExample()对应配置的输出
{"level":"debug","msg":"Debug message"}
{"level":"info","msg":"Info message"}
{"level":"warn","msg":"Warning message"}
{"level":"error","msg":"Error message"}
```

```zsh
# NewDevelopment()对应配置的输出
2020-09-28T17:49:26.098+0800    DEBUG   go-zap/main.go:19       Debug message
2020-09-28T17:49:26.099+0800    INFO    go-zap/main.go:20       Info message
2020-09-28T17:49:26.099+0800    WARN    go-zap/main.go:21       Warning message
main.main
        /Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-zap/main.go:21
runtime.main
        /usr/local/go/src/runtime/proc.go:204
2020-09-28T17:49:26.099+0800    ERROR   go-zap/main.go:22       Error message
main.main
        /Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-zap/main.go:22
runtime.main
        /usr/local/go/src/runtime/proc.go:204
```

```zsh
# NewProduction()对应配置的输出
{"level":"info","ts":1601286676.22206,"caller":"go-zap/main.go:20","msg":"Info message"}
{"level":"warn","ts":1601286676.22211,"caller":"go-zap/main.go:21","msg":"Warning message"}
{"level":"error","ts":1601286676.222116,"caller":"go-zap/main.go:22","msg":"Error message","stacktrace":"main.main\n\t/Users/penglin/gopath/src/github.com/Niclausse/golang-log-study/go-zap/main.go:22\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:204"}
```

除了使用zap提供的默认配置，用户也可以自定义配置，定制Logger。



### Fields用法

由于`fmt.Printf`之类的方法大量使用`interface{}`和反射，会有不少性能损失，并且增加了内存分配的频次。`zap`为了提高性能、减少内存分配次数，没有使用反射，而且默认的`Logger`只支持强类型的、结构化的日志。必须使用`zap`提供的方法记录字段。`zap`为 Go 语言中所有的基本类型和其他常见类型都提供了方法。也有任意类型的字段：`zap.Any()`、`zap.Binary()`等。

```go
	log.Error("Failed to send event",
		zap.String("request_id", requestID),
		zap.String("user_ip", userIP),
		zap.Int64("index", index),
		zap.Duration("request_time", time.Second),
	)
```

为每个字段都用方法包一层用起来比较繁琐。zap也提供了便捷的方法SugarLogger，可以使用`printf`格式符的方式。调用`logger.Sugar()`创建`SugaredLogger`

。`SugaredLogger`的使用比`Logger`简单，只是性能要低50%左右，可以用在**非热点函数**中。



### 记录层级关系

```go
	// 记录层级关系
	// 方式1
	log.Info("tracked some metrics",
		zap.Namespace("metrics"),
		zap.Int64("counter", 1),
		zap.String("name", "m1"),
	)

	// 方式2
	logger := log.With(
		zap.Namespace("metrics"),
		zap.Int("counter", 1),
		zap.String("name", "m2"),
	)
	logger.Info("tracked some metrics")
```

日志输出如下：

```zsh
{"level":"info","msg":"tracked some metrics","metrics":{"counter":1,"name":"m1"}}
{"level":"info","msg":"tracked some metrics","metrics":{"counter":1,"name":"m2"}}
```



## 五、Iris框架中使用的日志库 →golog

与glog类似，golog也是实现了日志分级的简单高效的日志库，不需要任何依赖。支持的日志级别如下：

| Name  | Method                | Text    |
| ----- | --------------------- | ------- |
| debug | Debug, Debugf         | [DEBUG] |
| info  | Info, Infof           | [INFO]  |
| warn  | Warn, Warnf, Warningf | [WARN]  |
| error | Error, Errorf         | [ERROR] |
| fatal | Fatal, Fatalf         | [FATAL] |

### 简单使用

```go
package main

import (
	log "github.com/kataras/golog"
)

func main() {
	log.SetLevel("debug")

	log.Println("Println message, no levels, no colors") // 不会打印日志级别
	log.Debug("Debug message")
	log.Info("Info message")
	log.Warn("Warning message")
	log.Error("Error message")
	log.Fatal("Fatal message")
}
```

```zsh
2020/09/28 19:22 Println message, no levels, no colors
[DBUG] 2020/09/28 19:22 Debug message
[INFO] 2020/09/28 19:22 Info message
[WARN] 2020/09/28 19:22 Warning message
[ERRO] 2020/09/28 19:22 Error message
[FTAL] 2020/09/28 19:22 Fatal message
exit status 1
```



### 集成Level-based标准Logger

使用`Install`和`InstallStd`方法可以将`golog`包裹在其它日志库logger之上。（任何实现了<a href="https://pkg.go.dev/github.com/kataras/golog#ExternalLogger">ExternalLogger</a>接口的Logger都可以适配）

例如：

```go
	// Simulate a logrus logger preparation
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	log.Install(logrus.StandardLogger())

	log.Debugf("Failed to send message: %s", "hello world!")
	log.Infof("Failed to send message: %s", "hello world!")
```

```zsh
{"file":"/Users/penglin/gopath/pkg/mod/github.com/kataras/golog@v0.0.10/integration.go:28","func":"github.com/kataras/golog.integrateExternalLogger.func1","level":"info","msg":"Failed to send message: hello world!","time":"2020-09-28T19:56:26+08:00"}
```



## 六、特性对比

|               | 日志分级        | ReportCaller报告文件及行号 | 输出格式                                                     | 日志切割                                   | 典型应用               |
| ------------- | --------------- | -------------------------- | ------------------------------------------------------------ | ------------------------------------------ | ---------------------- |
| log           | 不支持          | 不支持                     | text                                                         | 不支持，需自定义                           |                        |
| kataras/golog | 支持            | 不支持                     | text                                                         | 不支持，需自定义                           | Iris                   |
| google/glog   | 支持（无DEBUG） | 支持                       | text                                                         | 只能根据日志文件大小切割，不能根据日期切割 | k8s                    |
| logrus        | 支持            | 支持                       | 支持结构化输出。支持第三方日志格式（如Fluentd, logstash, elastic search、mq等），也可通过Hook自定义logging formaater | 不支持，需配合第三方库，如file-rotatelogs  | Docker, Prometheus     |
| Uber/zap      | 支持            | 支持                       | 支持结构化输出                                               | 不支持，需配合第三方库，如lumberjack       | Uber，Bilibili部分应用 |

