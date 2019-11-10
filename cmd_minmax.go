package main

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
