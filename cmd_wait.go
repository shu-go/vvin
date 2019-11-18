package main

import (
	"errors"
	"strings"
	"time"

	"github.com/shu-go/gli"
)

type waitCmd struct {
	_ struct{} `help:"[--close] {Title}"`

	Closed    bool         `help:"wait until the window is closed"`
	Intervals gli.Duration `cli:"intervals,i=DURATION" default:"1s"`
}

func (c waitCmd) Run(args []string) error {
	if len(args) != 1 {
		return errors.New("not one target")
	}

	an := ancestors()
	t := strings.ToLower(args[0])

	for {
		wins, err := listAllWindows()
		if err != nil {
			return err
		}

		win := findFirstTarget(t, wins, an)
		if c.Closed {
			if win == nil {
				break
			}
		} else {
			if win != nil {
				break
			}
		}

		time.Sleep(c.Intervals.Duration())
	}

	return nil
}
