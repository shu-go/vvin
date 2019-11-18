package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shu-go/gli"
)

type waitCmd struct {
	_ struct{} `help:"[--close] {Title}"`

	Closed    bool         `help:"wait until the window is closed"`
	Intervals gli.Duration `cli:"intervals,i=DURATION" default:"1s"`
	Timeout   gli.Duration `cli:"timeout=DURATION" default:"0s" help:"zelo value means ininite"`
}

func (c waitCmd) Run(args []string) error {
	if len(args) != 1 {
		return errors.New("not one target")
	}

	an := ancestors()
	t := strings.ToLower(args[0])

	var ctx context.Context
	if c.Timeout == 0 {
		ctx = context.Background()
	} else {
		ctx, _ = context.WithTimeout(context.Background(), c.Timeout.Duration())
	}

waitLoop:
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

		select {
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "cancelled")
			break waitLoop
		default:
			//nop
		}

		time.Sleep(c.Intervals.Duration())
	}

	return nil
}
