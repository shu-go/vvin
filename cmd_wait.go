package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type waitCmd struct {
	_ struct{} `help:"[--close] {Title}"`

	Closed   bool          `help:"wait until the window is closed"`
	Interval time.Duration `cli:"interval,i=DURATION" default:"1s"`
	Timeout  time.Duration `cli:"timeout=DURATION" default:"0s" help:"zero value means ininite"`
}

func (c waitCmd) Run(args []string) error {
	if len(args) != 1 {
		return errors.New("not one target")
	}

	an := ancestors()
	t := strings.ToLower(args[0])

	var ctx context.Context
	var cancel func()

	if c.Timeout == 0 {
		ctx = context.Background()
		cancel = func() {}
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), c.Timeout)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

	fmt.Println("Press Ctrl+C to cancel.")

waitLoop:
	for {
		wins, err := listAllWindows()
		if err != nil {
			cancel()
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
		case <-signalChan:
			fmt.Fprintln(os.Stderr, "cancelled")
			break waitLoop
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "cancelled")
			break waitLoop
		default:
			//nop
		}

		time.Sleep(c.Interval)
	}
	cancel()

	return nil
}
