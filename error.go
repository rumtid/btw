package btw

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"strings"
)

var (
	MaxStackDepth  int = 256
	MaxColumnWidth int = 25
)

type Error struct {
	Err     error
	Stack   []Frame
	Context []Layer
}

type Frame struct {
	Func string
	File string
	Line int
}

type Layer struct {
	Func   string
	Values []string
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func Trace(err error) error {
	if err == nil {
		return nil
	}
	e := &Error{Err: err}

	pc := make([]uintptr, MaxStackDepth)
	n := runtime.Callers(2, pc)
	if n == 0 {
		return e
	}

	frames := runtime.CallersFrames(pc[:n])
	for {
		f, more := frames.Next()
		e.Stack = append(e.Stack, Frame{
			Func: f.Function,
			File: f.File,
			Line: f.Line,
		})
		if !more {
			break
		}
	}

	return e
}

func Attach(err error, ctx ...interface{}) error {
	if err == nil {
		return nil
	}
	e, ok := err.(*Error)
	if !ok {
		e = &Error{Err: err}
	}

	ctx = ctx[:len(ctx)/2*2]
	if len(ctx) == 0 {
		return e
	}
	var l Layer
	if pc, _, _, ok := runtime.Caller(1); ok {
		l.Func = runtime.FuncForPC(pc).Name()
	} else {
		l.Func = "???"
	}
	l.Values = make([]string, len(ctx))
	for i, v := range ctx {
		l.Values[i] = fmt.Sprintf("%+v", v)
	}
	e.Context = append(e.Context, l)

	return e
}

func Format(err error) string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "error: %v", err)

	level := 1
	for err != nil {
		e, _ := err.(*Error)
		if e != nil {
			formatStack(level, e, b)
			formatContext(level, e, b)
			level++
		}
		err = errors.Unwrap(err)
	}

	return b.String()
}

func formatStack(l int, e *Error, b *strings.Builder) {
	var count int
	var prev string
	for i := 0; i < len(e.Stack); i++ {
		curr := fmt.Sprintf("\n%sfrom %s() at %s:%d",
			strings.Repeat("\t", l),
			path.Base(e.Stack[i].Func),
			e.Stack[i].File,
			e.Stack[i].Line,
		)
		if prev == curr {
			count++
			continue
		}
		if count != 0 {
			fmt.Fprintf(b, "\n%s     ... * %d",
				strings.Repeat("\t", l), count+1,
			)
			count = 0
		}
		b.WriteString(curr)
		prev = curr
	}
}

func formatContext(l int, e *Error, b *strings.Builder) {
	var maxWidth [2]int
	for _, layer := range e.Context {
		for i, item := range layer.Values {
			if n := len(item); n <= MaxColumnWidth {
				maxWidth[i%2] = max(maxWidth[i%2], n)
			}
		}
	}

	tags := make(map[string]struct{})
	for _, layer := range e.Context {
		for i := 0; i < len(layer.Values)-1; i += 2 {
			if _, ok := tags[layer.Values[i]]; ok {
				continue
			}
			tags[layer.Values[i]] = struct{}{}

			fmt.Fprintf(b, "\n%s[ %-[2]*[3]s ]",
				strings.Repeat("\t", l),
				maxWidth[0],
				layer.Values[i],
			)
			fmt.Fprintf(b, "    %[1]*[2]s",
				maxWidth[1],
				layer.Values[i+1],
			)
			if i == 0 {
				fmt.Fprintf(b, "    by %s()",
					path.Base(layer.Func),
				)
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
