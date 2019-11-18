package main

import (
	"errors"

	"github.com/shu-go/rog"
)

type alphaCmd struct {
}

func (c alphaCmd) Run(args []string, g globalCmd) error {
	if len(args) != 1 {
		return errors.New("an argument is required")
	}

	alpha := toInt(args[0], 255)
	if g.Debug {
		rog.Printf("alpha = %v -> %v", args[0], alpha)
	}

	style, _, _ := getWindowLong.Call(uintptr(g.targetHandle), GWL_EXSTYLE)
	setWindowLong.Call(uintptr(g.targetHandle), GWL_EXSTYLE, style|WS_EX_LAYERED)

	setLayeredWindowAttributes.Call(uintptr(g.targetHandle), 0, uintptr(alpha), LWA_ALPHA)
	if alpha == 255 {
		setWindowLong.Call(uintptr(g.targetHandle), GWL_EXSTYLE, style&^WS_EX_LAYERED)
	}

	return nil
}
