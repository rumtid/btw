# BTW

BTW is an error library in Go. It can carry backtrace of the stack and customized context in errors.

## Features

* Wrap a backtrace of the stack, fold recursive functions
* Attach customized context, shadow outers' duplicated keys
* Format human friendly error information with stack and context
* Make life easier :)

## Install

Run following command:

```bash
go get github.com/rumtid/btw
```

## Usage

```go
import "github.com/rumtid/btw"

// btw.Trace() wraps the stack trace information.
// It is common to trace errors from other modules.
err := otherMod.Func()
if err != nil {
    return btw.Trace(err)
}

// btw.Attach() records additional details.
// For example, the i variable is attached, and labelled as "retries".
if err != nil {
    err = btw.Attach(err,
        "retries", i,
    )
    return err
}

// btw.Format() outputs the error in a human-readable fashion.
if err != nil {
    fmt.Println(btw.Format(err))
}
// Output:
// error: something wrong
//         from main.Dial() at github.com/rumtid/example.go:10
//         [ retries ]    2    by main.Dial()

```
