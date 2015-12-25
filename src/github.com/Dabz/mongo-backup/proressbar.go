/*
** proressbar.go for proressbar.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Thu 24 Dec 23:55:40 2015 gaspar_d
** Last update Fri 25 Dec 02:32:15 2015 gaspar_d
*/

package main

import (
  "fmt"
  "sync"
  "syscall"
  "unsafe"
)

type WinSize struct {
  Ws_row    uint16 // rows, in characters
  Ws_col    uint16 // columns, in characters
  Ws_xpixel uint16 // horizontal size, pixels
  Ws_ypixel uint16 // vertical size, pixels
}

// see: http://www.delorie.com/djgpp/doc/libc/libc_495.html
func PBGetWinSize() (*WinSize, error) {
  ws := &WinSize{}

  _, _, err := syscall.Syscall(
    uintptr(syscall.SYS_IOCTL),
    uintptr(syscall.Stdout),
    uintptr(syscall.TIOCGWINSZ),
    uintptr(unsafe.Pointer(ws)),
  )

  ws.Ws_col = ws.Ws_col / 3;

  if err != 0 {
    return nil, err
  }

  return ws, nil
}

// clear whole line and move cursor to leftmost of line
func (e *env) PBClear() {
  fmt.Print("\033[2K\033[0G")
}

func (e *env) PBRepeat(str string, count int) string {
  var out string

  for i := 0; i < count; i++ {
    out += str
  }

  return out
}

const (
  remain = 5
)

var (
  mutex = &sync.Mutex{}
)

func (e *env) PBShow(percent float32) error {
  var (
    ws   *WinSize
    err  error
    ps   string
    half bool

    num   string
    pg    string
    space string

    pgl int
    l   int
  )

  mutex.Lock()
  defer mutex.Unlock()

  if ws, err = PBGetWinSize(); err != nil {
    return err
  }

  num = fmt.Sprintf("%.2f%%", percent*100)
  pgl = int(ws.Ws_col) - remain - 2 - 7
  half = int(percent*1000)%10 != 0
  percent = percent * 100 / 100
  count := percent * float32(pgl)
  pg = e.PBRepeat("=", int(count))

  if half {
    pg += "-"
  }

  l = pgl - len(pg)
  if l > 0 {
    space = e.PBRepeat(" ", l)
  }

  ps = pg + space

  e.PBClear()

  if int(percent) == 1 {
    fmt.Print(fmt.Sprintf("|%s| %s\n", ps, num))
  } else {
    fmt.Print(fmt.Sprintf("|%s| %s", ps, num))
  }

  return nil
}
