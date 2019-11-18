package main

type topmostCmd struct {
	Restore bool `cli:"restore,r"`
}

func (c topmostCmd) Run(g globalCmd) {
	hwndInsertAfter := hwndTopmost
	if c.Restore {
		hwndInsertAfter = hwndNoTopmost
	}

	setWindowPos.Call(
		uintptr(g.targetHandle),
		hwndInsertAfter,
		0,
		0,
		0,
		0,
		swpNoSize|swpNoMove)
}
