package main

type topmostCmd struct {
	Restore bool `cli:"restore,r"`
}

func (c topmostCmd) Run(g globalCmd) {
	hwndInsertAfter := HWND_TOPMOST
	if c.Restore {
		hwndInsertAfter = HWND_NOTOPMOST
	}

	setWindowPos.Call(
		uintptr(g.targetHandle),
		hwndInsertAfter,
		0,
		0,
		0,
		0,
		SWP_NOSIZE|SWP_NOMOVE)
}
