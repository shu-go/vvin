package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/mitchellh/go-ps"
	"github.com/shu-go/gli"
	"github.com/shu-go/rog"
)

// Version is app version
var Version string

func init() {
	if Version == "" {
		Version = "dev-" + time.Now().Format("20060102")
	}
}

type globalCmd struct {
	Target string `cli:"target,t=WINDOW_TITLE" help:"default to current window"`
	Debug  bool   `help:"output debug info"`

	Minimize minCmd     `cli:"minimize,min" help:"minimize/restore"`
	Maximize maxCmd     `cli:"maximize,max" help:"maximize/restore"`
	Resize   resizeCmd  `cli:"resize,move,mv" help:"resize/move"`
	Alpha    alphaCmd   `cli:"alpha"   help:"set alpha 0%(transparent) - 100%(opaque)"`
	Topmost  topmostCmd `cli:"topmost" help:"set always on top/restore"`

	targetHandle syscall.Handle

	scrWidth, scrHeight     int
	frameWidth, frameHeight int
}

func (c *globalCmd) Before() error {
	wins, err := listAllWindows()
	if err != nil {
		return err
	}

	an := ancestors()
	t := strings.ToLower(c.Target)

	for _, w := range wins {
		ancestor := false
		for _, p := range an {
			if w.PID == p {
				ancestor = true
				break
			}
		}

		if c.Debug {
			rog.Printf("win: %#v (ancestor? %v)", w, ancestor)
		}
		if t != "" && !ancestor {
			wt := strings.ToLower(w.Title)

			if strings.Contains(wt, t) {
				c.targetHandle = w.Handle
				break
			}
		} else if t == "" && ancestor {
			c.targetHandle = w.Handle
			break
		}
	}

	if c.targetHandle == 0 {
		return errors.New("no target")
	}

	w, _, _ := getSystemMetrics.Call(SM_CXVIRTUALSCREEN)
	h, _, _ := getSystemMetrics.Call(SM_CYVIRTUALSCREEN)
	c.scrWidth = int(w)
	c.scrHeight = int(h)

	return nil
}

func main() {
	app := gli.NewWith(&globalCmd{})
	app.Name = "vvin"
	app.Desc = "Change window properties for Windows"
	app.Version = Version
	app.Usage = ``
	app.Copyright = "(C) 2019 Shuhei Kubota"
	app.SuppressErrorOutput = true
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

var (
	user32                   = syscall.NewLazyDLL("user32.dll")
	enumWindows              = user32.NewProc("EnumWindows")
	getWindowText            = user32.NewProc("GetWindowTextW")
	getWindowTextLength      = user32.NewProc("GetWindowTextLengthW")
	getWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	isWindow                 = user32.NewProc("IsWindow")
	isWindowVisible          = user32.NewProc("IsWindowVisible")
	showWindow               = user32.NewProc("ShowWindow")
	setWindowPos             = user32.NewProc("SetWindowPos")
	getWindowRect            = user32.NewProc("GetWindowRect")
	getSystemMetrics         = user32.NewProc("GetSystemMetrics")

	setLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	getWindowLong              = user32.NewProc("GetWindowLongW")
	setWindowLong              = user32.NewProc("SetWindowLongW")
)

const (
	SW_MAXIMIZE = 3
	SW_MINIMIZE = 6
	SW_RESTORE  = 9
	SW_HIDE     = 0
	SW_SHOWNA   = 8

	SWP_NOSIZE     = 0x0001
	SWP_NOMOVE     = 0x0002
	SWP_NOZORDER   = 0x0004
	SWP_NOACTIVATE = 0x0010
	SWP_SHOWWINDOW = 0x0040

	SM_CXVIRTUALSCREEN = 78
	SM_CYVIRTUALSCREEN = 79
	SM_CXSIZEFRAME     = 32
	SM_CYSIZEFRAME     = 33

	GWL_EXSTYLE      = 0xFFFFFFEC
	WS_EX_TOOLWINDOW = 0x00000080
	WS_EX_LAYERED    = 0x80000

	LWA_ALPHA = 0x2

	HWND_TOPMOST   = ^uintptr(0)
	HWND_NOTOPMOST = ^uintptr(1)
)

type (
	Window struct {
		Title  string
		Handle syscall.Handle
		PID    int
	}

	RECT struct {
		Left, Top, Right, Bottom int32
	}
)

func listAllWindows() (wins []*Window, err error) {
	cb := syscall.NewCallback(func(hwnd syscall.Handle, lparam uintptr) uintptr {
		b, _, _ := isWindow.Call(uintptr(hwnd))
		if b == 0 {
			return 1
		}

		b, _, _ = isWindowVisible.Call(uintptr(hwnd))
		if b == 0 {
			return 1
		}

		title := ""
		tlen, _, _ := getWindowTextLength.Call(uintptr(hwnd))
		if tlen != 0 {
			tlen++
			buff := make([]uint16, tlen)
			getWindowText.Call(
				uintptr(hwnd),
				uintptr(unsafe.Pointer(&buff[0])),
				uintptr(tlen),
			)
			title = syscall.UTF16ToString(buff)
		}

		var processID uintptr
		getWindowThreadProcessId.Call(
			uintptr(hwnd),
			uintptr(unsafe.Pointer(&processID)),
		)

		win := &Window{
			Title:  title,
			Handle: hwnd,
			PID:    int(processID),
		}
		wins = append(wins, win)

		return 1
	})

	a, _, _ := enumWindows.Call(cb, 0)
	if a == 0 {
		return nil, fmt.Errorf("USER32.EnumWindows returned FALSE")
	}

	return wins, nil
}

func ancestors() []int {
	curr := os.Getpid()

	an := []int{curr}

	for {
		p, err := ps.FindProcess(curr)
		if p == nil || err != nil {
			break
		}

		curr = p.PPid()
		an = append(an, curr)
	}

	return an
}

func toInt(s string, max int) int32 {
	if strings.HasSuffix(s, "%") {
		i, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0
		}
		if i > 100 {
			i = 100
		}
		return int32(math.Trunc(float64(max*i) / 100))
	} else {
		i, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return int32(i)
	}
}
