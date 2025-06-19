# Suplog [![GoDoc](https://godoc.org/github.com/InjectiveLabs/suplog?status.svg)](https://godoc.org/github.com/InjectiveLabs/suplog)

A supercharged logging framework based upon [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus), which enables structured, leveled logging with hooks support. Hooks are middleware modules for logging that can augment message being logged or even send it to a remote server.

The key feature is an integrated stack tracing for any wrapped error (either from Go stdlib, or [github.com/pkg/errors](http://github.com/pkg/errors)).

<img alt="Screenshot of suplog" src="screenshot.png" width="600px" />

## Init options

```go
NewLogger(wr io.Writer, formatter Formatter, hooks ...Hook) Logger
```

Available formatters:
* `suplog.JSONFormatter` — suplogs all log entries as JSON objects
* `suplog.TextFormatter` — suplogs log entries as text lines for TTY or without TTY colors.

Available hooks:
* [github.com/InjectiveLabs/suplog/hooks/debug](https://github.com/InjectiveLabs/suplog/blob/master/hooks/debug/hook.go#L14)
* [github.com/InjectiveLabs/suplog/hooks/blob](https://github.com/InjectiveLabs/suplog/blob/master/hooks/blob/hook.go#L14)
* [github.com/InjectiveLabs/suplog/hooks/bugsnag](https://github.com/InjectiveLabs/suplog/blob/master/hooks/bugsnag/hook.go#L13)

## Leveled Logging

Suplog supports 7 levels: `Trace`, `Debug`, `Info`, `Warning`, `Error`, `Fatal` and `Panic`.

```go
log.Trace("Something very low level.")
log.Debug("Useful debugging information.")
log.Info("Something noteworthy happened!")
log.Warn("You should probably take a look at this.")
log.Error("Something failed but I'm not quitting.")
// Calls os.Exit(1) after logging
log.Fatal("Bye.")
// Calls panic() after logging
log.Panic("I'm bailing.")
```

You can set the logging level on an Logger, then it will only log entries with that severity or anything above it:

```go
// Will log anything that is info or above (warn, error, fatal, panic). Default.
log.SetLevel(suplog.InfoLevel)
```

Different levels will produce log lines of different colors. Also, some hooks will trigger on specific levels. For example, a debug hook will add infomation about line for `Debug` log entries. Another hook that enables Bugsnag support will report all errors and warnings to an external service.

## Structured Logging

In addition to log leveling, the new suplog package enables providing additional fields without altering the original message. By using this feature a developer can provide additional debug context.

```go
log.WithFields(suplog.Fields{
    "module": "accounts",
    "email":  "john@doe.com"
}).Error("account check failed")
```

In this case `email` field will be logged as a separate column that can be parsed and used as a filter. When using it with external services like Bugsnag, these fields will be reported in metadata tab and can be used for filtering too.

Suplog fields can be joined and chained:

```go
func init() {
    out = log.WithField("module", "fooer")
}

func runningFoo() {
    fooOut := log.WithField("action", "foo")

    for _, itemName := range items {
        itemOut := fooOut.WithField("item", itemName)
        itemOut.Info("processing item!")
    }
}
```

You should use chaining to avoid duplication of field context in sub-routines!

An example of issuing a warning without changing the original error:

```go
log.WithError(err).Warnln("something wrong happened")
```

## Hooks

During suplog initialisation it is possible to specify suplog hooks. Hooks are plugins that will pre-process log entries and do something useful. Below are several examples that are available to suplog users.

### Debug

Debug hook adds information about caller fn name and position is source code. By default applies only to `Debug` and `Trace` entries, but can be extended to any level.

```go
import debugHook github.com/InjectiveLabs/suplog/hooks/debug
```

Options:

```go
type HookOptions struct {
    // AppVersion specifies version of the app currently running.
    AppVersion string
    // Levels enables this hook for all listed levels.
    Levels []logrus.Level
    // PathSegmentsLimit allows to trim amount of source code file path segments.
    // Untrimmed: /Users/xlab/Documents/dev/go/src/github.com/InjectiveLabs/suplog/default_test.go
    // Trimmed (3): xlab/suplog/default_test.go
    PathSegmentsLimit int
}
```

If not specified, AppVersion is set from **APP_VERSION** env variable. PathSegmentsLimit is set to 3 by default, which means the latest 3 path segments of the source path.

### Bugsnag

Bugsnag hook implements integration with [Bugsnag.com](https://app.bugsnag.com) service for error tracing and monitoring. It will send any entry above warning level, including its meta data and stack trace.

```go
import bugsnagHook github.com/InjectiveLabs/suplog/hooks/bugsnag
```

Options:

```go
type HookOptions struct {
    // Levels enables this hook for all listed levels.
    Levels       []logrus.Level

    Env               string
    AppVersion        string
    BugsnagAPIKey     string
    BugsnagEnabledEnv []string
    BugsnagPackages   []string
}
```

Be default reporting is enabled for all levels above `Warning`.

The hook can be enabled in default suplogger by setting OS ENV variables:

* APP_ENV (e.g. `test`, `staging` or `prod`)
* APP_VERSION
* LOG_BUGSNAG_KEY
* **LOG_BUGSNAG_ENABLED** — this option enables bugsnag in default suplogger for existing codebase.

### Blob Uploads

Blob hook allows to upload heavy blobs of data such as request and response HTML / JSON dumps into a remote log storage. This hook utilizes Amazon S3 interface, therefore is compatible with any S3-like API.

```go
import blobHook github.com/InjectiveLabs/suplog/hooks/blob
```

Options:

```go
type HookOptions struct {
    Env               string
    BlobStoreURL      string
    BlobStoreAccount  string
    BlobStoreKey      string
    BlobStoreEndpoint string
    BlobStoreRegion   string
    BlobStoreBucket   string
    BlobRetentionTTL  time.Duration
    BlobEnabledEnv    map[string]bool
}
```

Be default blob uploading is enabled for all levels.

The following OS ENV variables are mapped:

* APP_ENV
* LOG_BLOB_STORE_URL
* LOG_BLOB_STORE_ACCOUNT
* LOG_BLOB_STORE_KEY
* LOG_BLOB_STORE_ENDPOINT
* LOG_BLOB_STORE_REGION
* LOG_BLOB_STORE_BUCKET
* **LOG_BLOB_ENABLED** — this option enables blob in default suplogger for existing codebase.

How to use:

```go
log.WithField("blob", testBlob).Infoln("test is running, trying to submit blob")
```

Where field name should be exactly `blob` and `testBlob` should be `[]byte`.

# Conditional triggers
It will only log if the condition is met, otherwise it will return a `NoOp` logger.

## `OnCondition`

### Description
Returns a `ConditionLogger` if the provided condition is `true`. If the condition is `false`, it returns a `NoOp` logger. If a logger is provided in the variadic argument, only the first logger is used.

### Behavior
- If `cond` is `true`, the function returns the first logger provided (or the default logger if none is provided).
- If `cond` is `false`, it returns a `NoOp` logger.

### Usage
```go
log.OnCondition(x > 0).WithField("foo", "bar").Info("This will log because condition is true")
````

---

## `OnErr`
Logs only if the provided error is not `nil`.

### Description
Returns a `ConditionLogger` if the provided error is not `nil`. If the error is `nil`, it returns a `NoOp` logger. If a logger is provided in the variadic argument, only the first logger is used.

### Parameters
- `err error`: The error to check.
- `logger ...Logger`: Optional variadic argument for the logger to use.

### Behavior
- If `err` is not `nil`, the function returns a logger with the error attached.
- If `err` is `nil`, it returns a `NoOp` logger.
 
### Usage
```go
log.OnError(someErr).WithField("foo", "bar").Info("something happened")
```

---

## `OnTime`
Receives a ticker/timer, if if the time is met, it will log. This is useful for preventing excessive logging
when log frequency is too high, similar to sampling

### Description
Returns a `ConditionLogger` if the provided tick channel receives a tick. If no tick is received, it returns a `NoOp` logger. The function checks the channel non-blockingly.

### Behavior
- If a tick is received from the channel, the function returns the first logger provided (or the default logger if none is provided).
- If no tick is received, it returns a `NoOp` logger.

### Usage
```go
ticker := time.NewTicker(1 * time.Minute)
log.OnTime(ticker.C).WithField("foo", "bar").Info("This will log every minute at most")
```
