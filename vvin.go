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
	Wait     waitCmd    `cli:"wait" help:"wait for a window is opened/closed ([--close] {Title})"`

	targetHandle syscall.Handle

	scrWidth, scrHeight     int
	frameWidth, frameHeight int
}

func (c *globalCmd) Before() error {
	wins, err := listAllWindows()
	if err != nil {
		return err
	}

	win := findFirstTarget(c.Target, wins, ancestors())
	if win == nil {
		return errors.New("no target")
	}
	c.targetHandle = win.Handle

	w, _, _ := getSystemMetrics.Call(smCXVirtualScreen)
	h, _, _ := getSystemMetrics.Call(smCYVirtualScreen)
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

	systemParametersInfo = user32.NewProc("SystemParametersInfoW")
)

const (
	swMaximize = 3
	smMinimize = 6
	swRestore  = 9
	swHide     = 0
	swShowNA   = 8

	swpNoSize     = 0x0001
	swpNoMove     = 0x0002
	swpNoZOrder   = 0x0004
	swpNoActivate = 0x0010
	swpShowWindow = 0x0040

	smCXVirtualScreen = 78
	smCYVirtualScreen = 79
	smCXSizeFrame     = 32
	smCYSizeFrame     = 33

	gwlEXStyle     = 0xFFFFFFEC
	wsEXToolWindow = 0x00000080
	wsEXLayered    = 0x80000

	lwaAlpha = 0x2

	hwndTopmost   = ^uintptr(0)
	hwndNoTopmost = ^uintptr(1)

	spiGetAnimation = 0x0048
	spiSetAnimation = 0x0049
)

type (
	window struct {
		Title  string
		Handle syscall.Handle
		PID    int
	}

	rect struct {
		Left, Top, Right, Bottom int32
	}

	anmationinfo struct {
		Size    uint32
		Animate int32
	}
)

func listAllWindows() (wins []*window, err error) {
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

		win := &window{
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
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return int32(i)
}

func findFirstTarget(title string, wins []*window, ancestors []int) *window {
	if title == "" {
		for _, p := range ancestors {
			for _, w := range wins {
				if w.PID == p {
					return w
				}
			}
		}
	} else {
		t := strings.ToLower(title)

		for _, w := range wins {
			ancestor := false
			for _, p := range ancestors {
				if w.PID == p {
					ancestor = true
					break
				}
			}

			if t != "" && !ancestor {
				wt := strings.ToLower(w.Title)

				if strings.Contains(wt, t) {
					return w
				}
			} else if t == "" && ancestor {
				return w
			}
		}
	}

	return nil
}
