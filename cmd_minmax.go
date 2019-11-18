package main

type minCmd struct {
	Restore bool `cli:"restore,r"`
}

func (c minCmd) Run(g globalCmd) {
	if c.Restore {
		showWindow.Call(uintptr(g.targetHandle), swRestore)
	} else {
		showWindow.Call(uintptr(g.targetHandle), smMinimize)
	}
}

type maxCmd struct {
	Restore bool `cli:"restore,r"`
}

func (c maxCmd) Run(g globalCmd) {
	if c.Restore {
		showWindow.Call(uintptr(g.targetHandle), swRestore)
	} else {
		showWindow.Call(uintptr(g.targetHandle), swMaximize)
	}
}
