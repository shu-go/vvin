package main

import (
	"errors"
	"unsafe"

	"github.com/shu-go/rog"
)

type resizeCmd struct {
	Left   string `cli:"left,x"`
	Top    string `cli:"top,y"`
	Width  string `cli:"width,w"`
	Height string `cli:"height,h"`

	Restore bool `cli:"restore,r"`

	NoRestorable bool `cli:"norestorable"`

	rect rect
}

func (c *resizeCmd) Before(g globalCmd) error {
	if c.Left == "" && c.Top == "" && c.Width == "" && c.Height == "" && !c.Restore {
		return errors.New("no options")
	}

	getWindowRect.Call(uintptr(g.targetHandle), uintptr(unsafe.Pointer(&c.rect)))

	oldrect := c.rect

	if g.Debug {
		rog.Print(oldrect)
	}
	if c.Left != "" {
		c.rect.Left = toInt(c.Left, g.scrWidth)
	}
	if c.Top != "" {
		c.rect.Top = toInt(c.Top, g.scrHeight)
	}
	if c.Width != "" {
		c.rect.Right = c.rect.Left + toInt(c.Width, g.scrWidth)
	} else {
		c.rect.Right = c.rect.Left + (oldrect.Right - oldrect.Left)
	}
	if c.Height != "" {
		c.rect.Bottom = c.rect.Top + toInt(c.Height, g.scrHeight)
	} else {
		c.rect.Bottom = c.rect.Top + (oldrect.Bottom - oldrect.Top)
	}
	if g.Debug {
		if c.Restore {
			rog.Print("restore")
		} else {
			rog.Print(c.rect)
		}
	}

	return nil
}

func (c resizeCmd) Run(g globalCmd) {
	if c.Restore {
		showWindow.Call(uintptr(g.targetHandle), swRestore)
		return
	}

	if !c.NoRestorable {
		showWindow.Call(uintptr(g.targetHandle), swHide)
		showWindow.Call(uintptr(g.targetHandle), swMaximize)
	}
	setWindowPos.Call(
		uintptr(g.targetHandle),
		0,
		uintptr(c.rect.Left),
		uintptr(c.rect.Top),
		uintptr(c.rect.Right-c.rect.Left),
		uintptr(c.rect.Bottom-c.rect.Top),
		swpNoActivate|swpNoZOrder)
	if !c.NoRestorable {
		showWindow.Call(uintptr(g.targetHandle), swShowNA)
	}
}
