/*
** oplog.go for oplog.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Sat 26 Dec 22:49:07 2015 gaspar_d
** Last update Tue 29 Dec 20:59:25 2015 gaspar_d
*/

package main

import (
  "os"
  "io"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
   "github.com/pierrec/lz4"
)


// dump the cursor to a directory
// if compress option is specified, use lz4 while dumping
func (e *Env) dumpOplogToDir(cursor *mgo.Iter, dir string) (error, float32, bson.MongoTimestamp) {
  var (
    destfile io.Writer
    opcount  float32
    counter  float32
    dest     string
    pb       Progessbar
		lastop   bson.MongoTimestamp
  )

  err := os.MkdirAll(dir, 0777);
  if err != nil {
    return err, 0, lastop
  }

  dest     = dir + "/oplog.bson"
  opcount  = float32(e.getOplogCount())
  counter  = 0
  pb.title = "oplog dump"
	lastop   = e.homeval.lastOplog

  pb.Show(0)

  if e.options.compress {
    dest         += ".lz4"
    dfile, err   := os.Create(dest)
    if err != nil {
      return err, 0, lastop
    }
    defer dfile.Close();
    destfile = lz4.NewWriter(dfile);
  } else {
    dfile, err := os.Create(dest);
    destfile   = dfile;
    if err != nil {
      return err, 0, lastop;
    }
    defer dfile.Close();
  }


	lastRow := bson.Raw{}
  for {
		raw   := &bson.Raw{}
    next  := cursor.Next(raw)

    if !next {
			if lastRow.Data != nil {
			  lastRowUnmarshal := bson.M{}
			  bson.Unmarshal(lastRow.Data, &lastRowUnmarshal)
			  lastop = lastRowUnmarshal["ts"].(bson.MongoTimestamp)
		  }
      break;
    }

    buff := make([]byte, len(raw.Data))
    copy(buff, raw.Data)
    destfile.Write(buff)
    counter += 1
		lastRow  = *raw
    pb.Show(counter / opcount)
  }

  pb.Show(1)
  pb.End()

  return nil, float32(e.GetDirSize(dir)), lastop
}
