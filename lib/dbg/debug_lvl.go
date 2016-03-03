package dbg

import (
	"fmt"
	"github.com/daviddengcn/go-colortext"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"sync"
)

// These are information-debugging levels that can be turned on or off.
// Every logging greater than 'DebugVisible' will be discarded. So you can
// Log at different levels and easily turn on or off the amount of logging
// generated by adjusting the 'DebugVisible' variable.
var debugVisible = 1
var debugMut sync.RWMutex

// The padding of functions to make a nice debug-output - this is automatically updated
// whenever there are longer functions and kept at that new maximum. If you prefer
// to have a fixed output and don't remember oversized names, put a negative value
// in here
var NamePadding = 40

// Padding of line-numbers for a nice debug-output - used in the same way as
// NamePadding
var LinePadding = 3

// Testing output has to be on fmt, it doesn't take into account log-outputs
// So for testing, set Testing = true, and instead of sending to log, it will
// output to fmt
var Testing = false

// If this variable is set, it will be outputted between the position and the message
var StaticMsg = ""

var regexpPaths, _ = regexp.Compile(".*/")

const (
	LvlPrint = iota - 10
	LvlWarning
	LvlError
	LvlFatal
	LvlPanic
)

// Needs two functions to keep the caller-depth the same and find who calls us
// Lvlf1 -> Lvlf -> Lvl
// or
// Lvl1 -> Lvld -> Lvl
func Lvld(lvl int, args ...interface{}) {
	Lvl(lvl, args...)
}
func Lvl(lvl int, args ...interface{}) {
	debugMut.Lock()
	defer debugMut.Unlock()

	if lvl > debugVisible {
		return
	}
	pc, _, line, _ := runtime.Caller(3)
	name := regexpPaths.ReplaceAllString(runtime.FuncForPC(pc).Name(), "")
	lineStr := fmt.Sprintf("%d", line)

	// For the testing-framework, we check the resulting string. So as not to
	// have the tests fail every time somebody moves the functions, we put
	// the line-# to 0
	if Testing {
		line = 0
	}

	if len(name) > NamePadding && NamePadding > 0 {
		NamePadding = len(name)
	}
	if len(lineStr) > LinePadding && LinePadding > 0 {
		LinePadding = len(name)
	}
	fmtstr := fmt.Sprintf("%%%ds: %%%dd", NamePadding, LinePadding)
	caller := fmt.Sprintf(fmtstr, name, line)
	if StaticMsg != "" {
		caller += "@" + StaticMsg
	}
	message := fmt.Sprintln(args...)
	var lvlStr string
	if lvl > 0 {
		lvlStr = strconv.Itoa(lvl)
	} else {
		lvlStr = "!" + strconv.Itoa(-lvl)
	}
	switch lvl {
	case LvlPrint:
		ct.Foreground(ct.White, true)
		lvlStr = "I"
	case LvlWarning:
		ct.Foreground(ct.Green, true)
		lvlStr = "W"
	case LvlError:
		ct.Foreground(ct.Red, false)
		lvlStr = "E"
	case LvlFatal:
		ct.Foreground(ct.Red, true)
		lvlStr = "F"
	case LvlPanic:
		ct.Foreground(ct.Red, true)
		lvlStr = "P"
	default:
		if lvl != 0 {
			bright := lvl < 0
			lvlAbs := lvl * (lvl / lvl)
			if lvlAbs <= 5 {
				colors := []ct.Color{ct.Yellow, ct.Cyan, ct.Blue, ct.Green, ct.Cyan}
				ct.Foreground(colors[lvlAbs-1], bright)
			}
		}
	}
	fmt.Printf("%-2s: (%s) - %s", lvlStr, caller, message)
	ct.ResetColor()
}

func Lvlf(lvl int, f string, args ...interface{}) {
	Lvl(lvl, fmt.Sprintf(f, args...))
}

func Print(args ...interface{}) {
	Lvld(LvlPrint, args...)
}

func Printf(f string, args ...interface{}) {
	Lvlf(LvlPrint, f, args...)
}

func Lvl1(args ...interface{}) {
	Lvld(1, args...)
}

func Lvl2(args ...interface{}) {
	Lvld(2, args...)
}

func Lvl3(args ...interface{}) {
	Lvld(3, args...)
}

func Lvl4(args ...interface{}) {
	Lvld(4, args...)
}

func Lvl5(args ...interface{}) {
	Lvld(5, args...)
}

func Error(args ...interface{}) {
	Lvld(LvlError, args...)
}

func Warn(args ...interface{}) {
	args = append([]interface{}{"WARN:"}, args...)
	Lvld(LvlWarning, args...)
}

func Fatal(args ...interface{}) {
	Lvld(LvlFatal, args...)
	os.Exit(1)
}

func Panic(args ...interface{}) {
	Lvld(LvlPanic, args...)
	panic(args)
}

func Lvlf1(f string, args ...interface{}) {
	Lvlf(1, f, args...)
}

func Lvlf2(f string, args ...interface{}) {
	Lvlf(2, f, args...)
}

func Lvlf3(f string, args ...interface{}) {
	Lvlf(3, f, args...)
}

func Lvlf4(f string, args ...interface{}) {
	Lvlf(4, f, args...)
}

func Lvlf5(f string, args ...interface{}) {
	Lvlf(5, f, args...)
}

func Fatalf(f string, args ...interface{}) {
	Lvlf(LvlFatal, f, args...)
	os.Exit(1)
}

func Errorf(f string, args ...interface{}) {
	Lvlf(LvlError, f, args...)
}

func Warnf(f string, args ...interface{}) {
	Lvlf(LvlWarning, "WARN: "+f, args...)
}

func Panicf(f string, args ...interface{}) {
	Lvlf(LvlPanic, f, args...)
	panic(args)
}

// SetDebugVisible set the global debug output level in a go-rountine-safe way
func SetDebugVisible(lvl int) {
	debugMut.Lock()
	defer debugMut.Unlock()
	debugVisible = lvl
}

func DebugVisible() int {
	debugMut.RLock()
	defer debugMut.RUnlock()
	return debugVisible
}

// TestOutput sets the DebugVisible to 0 if 'show'
// is false, else it will set DebugVisible to 'level'
//
// Usage: TestOutput( test.Verbose(), 2 )
func TestOutput(show bool, level int) {
	debugMut.Lock()
	defer debugMut.Unlock()

	if show {
		debugVisible = level
	} else {
		debugVisible = 0
	}
}

// To easy print a debug-message anyway without discarding the level
// Just add an additional "L" in front, and remove it later:
// - easy hack to turn on other debug-messages
// - easy removable by searching/replacing 'LLvl' with 'Lvl'
func LLvl1(args ...interface{})            { Lvld(-1, args...) }
func LLvl2(args ...interface{})            { Lvld(-1, args...) }
func LLvl3(args ...interface{})            { Lvld(-1, args...) }
func LLvl4(args ...interface{})            { Lvld(-1, args...) }
func LLvl5(args ...interface{})            { Lvld(-1, args...) }
func LLvlf1(f string, args ...interface{}) { Lvlf(-1, f, args...) }
func LLvlf2(f string, args ...interface{}) { Lvlf(-1, f, args...) }
func LLvlf3(f string, args ...interface{}) { Lvlf(-1, f, args...) }
func LLvlf4(f string, args ...interface{}) { Lvlf(-1, f, args...) }
func LLvlf5(f string, args ...interface{}) { Lvlf(-1, f, args...) }
