package main

import (
	"errors"
	"os"

	"github.com/shu-go/nmfmt"
)

type alphaCmd struct {
}

func (c alphaCmd) Run(args []string, g globalCmd) error {
	if len(args) != 1 {
		return errors.New("an argument (opacity; 0%-100% or 0-255) is required")
	}

	opacity := toInt(args[0], 255)

	g.debug(os.Stderr, "opacity $=arg:q -> $opacity/255\n",
		nmfmt.M{
			"arg":     args[0],
			"opacity": opacity,
		})

	style, _, _ := getWindowLong.Call(uintptr(g.targetHandle), gwlEXStyle)
	setWindowLong.Call(uintptr(g.targetHandle), gwlEXStyle, style|wsEXLayered)

	setLayeredWindowAttributes.Call(uintptr(g.targetHandle), 0, uintptr(opacity), lwaAlpha)
	if opacity == 255 {
		setWindowLong.Call(uintptr(g.targetHandle), gwlEXStyle, style&^wsEXLayered)
	}

	return nil
}
