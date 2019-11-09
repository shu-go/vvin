package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/mitchellh/go-ps"
	"github.com/shu-go/gli"
	"github.com/shu-go/rog"
)

type globalCmd struct {
	Target string `cli:"target,t=WINDOW_TITLE" help:"default to current window"`
	Debug  bool

	Minimize minCmd    `cli:"minimize,min"`
	Maximize maxCmd    `cli:"maximize,max"`
	Resize   resizeCmd `cli:"resize"`
	Move     moveCmd   `cli:"move,mv"`

	targetHandle syscall.Handle

	scrWidth, scrHeight int
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

type minCmd struct {
	Restore bool `cli:"restore,r"`
}

func (c minCmd) Run(g globalCmd) {
	if c.Restore {
		showWindow.Call(uintptr(g.targetHandle), SW_RESTORE)
	} else {
		showWindow.Call(uintptr(g.targetHandle), SW_MINIMIZE)
	}
}

type maxCmd struct {
	Restore bool `cli:"restore,r"`
}

func (c maxCmd) Run(g globalCmd) {
	if c.Restore {
		showWindow.Call(uintptr(g.targetHandle), SW_RESTORE)
	} else {
		showWindow.Call(uintptr(g.targetHandle), SW_MAXIMIZE)
	}
}

type resizeCmd struct {
	Left   int `cli:"left,x"`
	Top    int `cli:"top,y"`
	Width  int `cli:"width,w"`
	Height int `cli:"height,h"`

	NoRestorable bool `cli:"norestorable"`
}

func (c resizeCmd) Run(g globalCmd) {
	if !c.NoRestorable {
		showWindow.Call(uintptr(g.targetHandle), SW_MAXIMIZE)
	}
	setWindowPos.Call(
		uintptr(g.targetHandle),
		0,
		uintptr(c.Left),
		uintptr(c.Top),
		uintptr(c.Width),
		uintptr(c.Height),
		SWP_NOACTIVATE|SWP_NOZORDER)
}

type moveCmd struct {
	Left int `cli:"left,x"`
	Top  int `cli:"top,y"`

	NoRestorable bool `cli:"norestorable"`
}

func (c moveCmd) Run(g globalCmd) {
	rect := struct {
		Left, Top, Right, Bottom int32
	}{}

	if !c.NoRestorable {
		getWindowRect.Call(uintptr(g.targetHandle), uintptr(unsafe.Pointer(&rect)))
		showWindow.Call(uintptr(g.targetHandle), SW_MAXIMIZE)
		setWindowPos.Call(
			uintptr(g.targetHandle),
			0,
			uintptr(c.Left),
			uintptr(c.Top),
			uintptr(rect.Right-rect.Left),
			uintptr(rect.Bottom-rect.Top),
			SWP_NOACTIVATE|SWP_NOZORDER)
	} else {
		setWindowPos.Call(
			uintptr(g.targetHandle),
			0,
			uintptr(c.Left),
			uintptr(c.Top),
			0,
			0,
			SWP_NOACTIVATE|SWP_NOZORDER|SWP_NOSIZE)
	}
}

func main() {
	app := gli.NewWith(&globalCmd{})
	app.Name = "vvin"
	app.Desc = ""
	app.Version = "0.0.0"
	app.Usage = ``
	app.Copyright = "(C) 2019 Shuhei Kubota"
	app.Run(os.Args)
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
)

const (
	SW_MAXIMIZE = 3
	SW_MINIMIZE = 6
	SW_RESTORE  = 9

	SWP_NOACTIVATE = 0x0010
	SWP_NOSIZE     = 0x0001
	SWP_NOZORDER   = 0x0004

	SM_CXVIRTUALSCREEN = 78
	SM_CYVIRTUALSCREEN = 79
)

type (
	Window struct {
		Title  string
		Handle syscall.Handle
		PID    int
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
