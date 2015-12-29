/*
** proressbar.go for proressbar.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Thu 24 Dec 23:55:40 2015 gaspar_d
** Last update Tue 29 Dec 22:21:52 2015 gaspar_d
*/

package main

import (
  "fmt"
  "sync"
  "syscall"
  "unsafe"
)

type Progessbar struct {
  title string
  scale uint16
  ended bool
}

type WinSize struct {
  Ws_row    uint16 // rows, in characters
  Ws_col    uint16 // columns, in characters
  Ws_xpixel uint16 // horizontal size, pixels
  Ws_ypixel uint16 // vertical size, pixels
}

// see: http://www.delorie.com/djgpp/doc/libc/libc_495.html
func (p *Progessbar) GetWinSize() (*WinSize, error) {
  ws := &WinSize{}

  _, _, err := syscall.Syscall(
    uintptr(syscall.SYS_IOCTL),
    uintptr(syscall.Stdout),
    uintptr(syscall.TIOCGWINSZ),
    uintptr(unsafe.Pointer(ws)),
  )

 if p.scale == 0 {
   p.scale = 3
 }

  ws.Ws_col = ws.Ws_col / p.scale

  if err != 0 {
    return nil, err
  }

  return ws, nil
}

// clear whole line and move cursor to leftmost of line
func (p *Progessbar) Clear() {
  fmt.Print("\033[2K\033[0G")
}

func (b *Progessbar) Repeat(str string, count int) string {
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

func (p *Progessbar) Show(percent float32) error {
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

  if p.ended {
    return nil;
  }

  mutex.Lock()
  defer mutex.Unlock()

  if ws, err = p.GetWinSize(); err != nil {
    return err
  }

  num = fmt.Sprintf("(%.2f%%)", percent*100)
  pgl = int(ws.Ws_col) - remain - 2 - 7
  half = int(percent*1000)%10 != 0
  percent = percent * 100 / 100
  count := percent * float32(pgl)
  pg = p.Repeat("=", int(count))

  if half {
    pg += "-"
  }

  l = pgl - len(pg)
  if l > 0 {
    space = p.Repeat(" ", l)
  }

  ps = pg + space

  p.Clear()

  if int(percent) == 1 && !p.ended {
    fmt.Print(fmt.Sprintf("%s |%s| %s\n", p.title, ps, num))
    p.ended = true
  } else {
    fmt.Print(fmt.Sprintf("%s |%s| %s", p.title, ps, num))
  }

  return nil
}

func (p *Progessbar) End() {
  if (!p.ended) {
    fmt.Print("\n")
    p.ended = true
  }
}
