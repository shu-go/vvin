keep waking Windows up

[![Go Report Card](https://goreportcard.com/badge/github.com/shu-go/vvin)](https://goreportcard.com/report/github.com/shu-go/vvin)
![MIT License](https://img.shields.io/badge/License-MIT-blue)

# Usage

```
Sub commands:
  minimize, min     minimize/restore
  maximize, max     maximize/restore
  resize, move, mv  resize/move
  alpha             set alpha 0%(transparent) - 100%(opaque)
  topmost           set always on top/restore
  wait              wait for a window is opened/closed ([--close] {Title})

Options:
  -t, --target WINDOW_TITLE  default to current window
```

## minimize/restore

```
> ./vvin -t Notepad min

> ./vvin -t Notepad min --restore
> ./vvin -t Notepad min -r
```

## resize

```
> ./vvin -t notepad resize -x 0 -y 0 -w 60% -h 100%
```

## alpha

```
> ./vvin -t notepad alpha 75%
```

## always on top

```
> ./vvin -t notepad topmost

> ./vvin -t notepad topmost --restore
> ./vvin -t notepad topmost -r
```

## wait

### wait for a window to appear

```
> ./vvin wait notepad
```

### wait closed

```
> ./vvin wait notepad --closed
```

