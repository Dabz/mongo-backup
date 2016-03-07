/*
** oplog.go for oplog.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Sat 26 Dec 22:49:07 2015 gaspar_d
** Last update Mon  7 Mar 16:53:52 2016 gaspar_d
*/

package mongobackup

import (
  "os"
  "io"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "github.com/pierrec/lz4"
  "github.com/Dabz/utils"
)

const (
  OPLOG_DIR  = "oplog/"
  OPLOG_FILE = "oplog.bson"
)


// dump the cursor to a directory
// if compress option is specified, use lz4 while dumping
// return the error if any, the number of byte restored, the first & last
// oplog dumped
func (e *BackupEnv) BackupOplogToDir(cursor *mgo.Iter, dir string) (error, float32, bson.MongoTimestamp, bson.MongoTimestamp) {
  var (
    destfile io.Writer
    opcount  float32
    counter  float32
    dest     string
    pb       utils.ProgressBar
    lastop   bson.MongoTimestamp
    firstop  bson.MongoTimestamp
  )

  err     := os.MkdirAll(dir, 0777);
  if err != nil {
    return err, 0, firstop, lastop
  }

  dest     = dir + "/" + OPLOG_FILE
  opcount  = float32(e.getOplogCount())
  counter  = 0
  pb.Title = "oplog dump"
  lastop   = e.homeval.lastOplog
  firstop  = e.homeval.lastOplog

  pb.Show(0)

  if e.Options.Compress {
    dest         += ".lz4"
    dfile, err   := os.Create(dest)
    if err != nil {
      return err, 0, firstop, lastop
    }
    defer dfile.Close();
    destfile = lz4.NewWriter(dfile);
  } else {
    dfile, err := os.Create(dest);
    destfile   = dfile;
    if err != nil {
      return err, 0, firstop, lastop;
    }
    defer dfile.Close();
  }


  lastRow := bson.Raw{}
  isFirst := true
  for {
    raw   := &bson.Raw{}
    next  := cursor.Next(raw)

    if !next {
      // Record last entry saved
      if lastRow.Data != nil {
        lastRowUnmarshal := bson.M{}
        bson.Unmarshal(lastRow.Data, &lastRowUnmarshal)
        lastop = lastRowUnmarshal["ts"].(bson.MongoTimestamp)
      }
      break;
    }

    // Record first entry saved
    if isFirst {
      isFirst       = false
      rowUnmarshal := bson.M{}
      bson.Unmarshal(raw.Data, &rowUnmarshal)
      firstop = rowUnmarshal["ts"].(bson.MongoTimestamp)
      continue
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

  return nil, float32(e.GetDirSize(dir)), firstop, lastop
}


// dump the oplog between the entries to the requested output directory
func (e *BackupEnv) DumpOplogsToDir(from, to *BackupEntry) error {
  destdir   := e.Options.Output + "/" + OPLOG_DIR
  oplogfile := destdir + OPLOG_FILE
  err       := os.MkdirAll(destdir, 0700)
  pb        := utils.ProgressBar{}
  pb.Scale   = 3
  if err != nil {
    return err
  }

  destfile, err := os.OpenFile(oplogfile, os.O_CREATE|os.O_RDWR, 0700)
  if err != nil {
    return err
  }

  entries := e.homeval.GetIncEntriesBetween(from, to)
  total   := len(entries)
  for index, entry := range entries {
    var reader io.Reader
    if entry.Compress {
      sourcename      :=  OPLOG_FILE + ".lz4"
      sourcefile, err := os.Open(entry.Dest + "/" + sourcename)
      if err != nil {
        pb.End()
        return err
      }
      reader = lz4.NewReader(sourcefile)
    } else {
      sourcefile, err   := os.Open(entry.Dest + "/" + OPLOG_FILE)
      reader             = sourcefile
      if err != nil {
        pb.End()
        return err
      }
    }

    pb.Title   = "dumping " + entry.Id
    pb.Show(float32(index) / float32(total))

    io.Copy(destfile, reader)
  }
  pb.Title = "dumping"
  pb.Show(1)
  pb.End()
  return nil
}

