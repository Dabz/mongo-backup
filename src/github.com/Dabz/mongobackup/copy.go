/*
** copy.go for copy.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Thu 24 Dec 23:43:24 2015 gaspar_d
** Last update Mon 28 Dec 11:16:15 2015 gaspar_d
*/

package mongobackup

import (
  "os"
  "io"
  "github.com/pierrec/lz4"
)



// Copy a file to another destination
// if the compress flag is present, compress the file while copying using lz4
func (e *env) CopyFile(source string, dest string) (err error, backedByte int64) {
  sourcefile, err := os.Open(source);
  if err != nil {
    return err, 0;
  }

  defer sourcefile.Close();

  var destfile io.Writer;
  if (e.options.compress) {
    dest         += ".lz4";
    dfile, err   := os.Create(dest);
    if err != nil {
      return err, 0;
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

  _, err = io.Copy(destfile, sourcefile)
  if err == nil {
    sourceinfo, err := os.Stat(source);
      if err != nil {
        err = os.Chmod(dest, sourceinfo.Mode());
      }
  }

  sourceinfo, _ := os.Stat(source);

  return nil, sourceinfo.Size();
}

// Return the total size of the directory in byte
func (e *env) GetDirSize(source string) (int64) {
  directory, _   := os.Open(source);
  var sum int64   = 0;
  defer directory.Close();

  objects, _ := directory.Readdir(-1)
  for _, obj := range objects {
    if obj.IsDir() {
      sum += e.GetDirSize(source + "/" + obj.Name());
    } else {
      stat, _ := os.Stat(source + "/" + obj.Name());
      sum += stat.Size();
    }
  }

  return sum;
}



// Copy a directory into another and compress all files if required
func (e *env) CopyDir(source string, dest string) (err error, backedByte int64) {
  totalSize      := e.GetDirSize(source)
  pb             := Progessbar{}
  pb.title        = "backup"
  pb.scale        = 3
  err, _          =  e.recCopyDir(source, dest, 0, totalSize, &pb)

  pb.End();

  if err != nil {
    return err, 0
  }

  return nil, e.GetDirSize(dest)
}


// Recursive copy directory function
func (e *env) recCopyDir(source string, dest string, backedByte int64, totalSize int64, pb *Progessbar) (err error, oBackedByte int64) {
  sourceinfo, err := os.Stat(source);

  if err != nil {
    return err, 0;
  }

  err = os.MkdirAll(dest, sourceinfo.Mode());
  if err != nil {
    return err, 0;
  }

  directory, _ := os.Open(source)
  objects, err := directory.Readdir(-1)

  for _, obj := range objects {
    if (obj.Name() == "mongod.lock") {
      continue;
    }

    sourcefilepointer := source + "/" + obj.Name()
    destinationfilepointer := dest + "/" + obj.Name()

    if obj.IsDir() {
      err, backedByte  = e.recCopyDir(sourcefilepointer, destinationfilepointer, backedByte, totalSize, pb)
      if err != nil {
        e.error.Println(err)
      }
    } else {
      err, size := e.CopyFile(sourcefilepointer, destinationfilepointer);
      if err != nil {
        e.error.Println(err);
      }
      backedByte = backedByte + size;
      pb.Show(float32(backedByte) / float32(totalSize))
    }
  }

  return nil, backedByte
}
