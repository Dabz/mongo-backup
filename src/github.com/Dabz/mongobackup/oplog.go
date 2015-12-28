/*
** oplog.go for oplog.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Sat 26 Dec 22:49:07 2015 gaspar_d
** Last update Mon 28 Dec 11:10:56 2015 gaspar_d
*/

package main

import (
  "os"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "io"
   "github.com/pierrec/lz4"
)


// dump the cursor to a directory
// if compress option is specified, use lz4 while dumping
func (e *Env) dumpOplogToDir(cursor *mgo.Iter, dir string) (error, float32) {
  var (
    destfile io.Writer
    opcount  float32
    counter  float32
    dest     string
    pb       Progessbar
  )

  err := os.MkdirAll(dir, 0777);
  if err != nil {
    return err, 0
  }

  dest     = dir + "/oplog.bson"
  opcount  = float32(e.getOplogCount())
  counter  = 0
  pb.title = "oplog dump"

  pb.Show(0)

  if e.options.compress {
    dest         += ".lz4"
    dfile, err   := os.Create(dest)
    if err != nil {
      return err, 0
    }
    defer dfile.Close();
    destfile = lz4.NewWriter(dfile);
  } else {
    dfile, err := os.Create(dest);
    destfile   = dfile;
    if err != nil {
      return err, 0;
    }
    defer dfile.Close();
  }


  for {
    raw  := &bson.Raw{}
    next := cursor.Next(raw)

    if !next {
      break;
    }

    buff := make([]byte, len(raw.Data))
    copy(buff, raw.Data)
    destfile.Write(buff)
    counter += 1
    pb.Show(counter / opcount)
  }

  pb.Show(1)
  pb.End()

  return nil, float32(e.GetDirSize(dir))
}
