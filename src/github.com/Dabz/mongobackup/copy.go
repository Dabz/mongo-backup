/*
** copy.go for copy.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Thu 24 Dec 23:43:24 2015 gaspar_d
** Last update Wed  6 Jan 20:05:17 2016 gaspar_d
*/

package mongobackup

import (
  "os"
  "io"
	"strings"
  "github.com/pierrec/lz4"
	"github.com/Dabz/utils"
)



// Copy a file to another destination
// if the compress flag is present, compress the file while copying using lz4
func (e *BackupEnv) CopyFile(source string, dest string) (err error, backedByte int64) {
  sourcefile, err := os.Open(source)
  if err != nil {
    return err, 0
  }

  defer sourcefile.Close()

  var destfile io.Writer
  if (e.Options.Compress) {
    dest         += ".lz4"
    dfile, err   := os.Create(dest)
    if err != nil {
      return err, 0
    }
    defer dfile.Close()
    destfile = lz4.NewWriter(dfile)
  } else {
    dfile, err := os.Create(dest)
    destfile   = dfile
    if err != nil {
      return err, 0
    }
    defer dfile.Close()
  }

  _, err = io.Copy(destfile, sourcefile)
	if err != nil {
		return err, 0
	}

	sourceinfo, err := os.Stat(source);
	if err != nil {
		return err, 0
	}

  return nil, sourceinfo.Size();
}

// Return the total size of the directory in byte
func (e *BackupEnv) GetDirSize(source string) (int64) {
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
func (e *BackupEnv) CopyDir(source string, dest string) (err error, backedByte int64) {
  totalSize      := e.GetDirSize(source)
  pb             := utils.ProgressBar{}
  pb.Title        = "backup"
  pb.Scale        = 3
  err, _          = e.recCopyDir(source, dest, 0, totalSize, &pb)

  pb.End();

  if err != nil {
    return err, 0
  }

  return nil, e.GetDirSize(dest)
}


// Recursive copy directory function
func (e *BackupEnv) recCopyDir(source string, dest string, backedByte int64, totalSize int64, pb *utils.ProgressBar) (err error, oBackedByte int64) {
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

    sourcefilepointer      := source + "/" + obj.Name()
    destinationfilepointer := dest + "/" + obj.Name()

    if obj.IsDir() {
      err, backedByte  = e.recCopyDir(sourcefilepointer, destinationfilepointer, backedByte, totalSize, pb)
      if err != nil {
        e.error.Println(err)
				return err, 0
      }
    } else {
      err, size := e.CopyFile(sourcefilepointer, destinationfilepointer);
      if err != nil {
        e.error.Println(err);
				return err, 0
      }
      backedByte = backedByte + size;
      pb.Show(float32(backedByte) / float32(totalSize))
    }
  }

  return nil, backedByte
}


// restore & uncompress a backup to a specific location
func (e *BackupEnv) RestoreCopyDir(entry *BackupEntry, source string, dest string, restoredByte int64, totalRestored int64, pb *utils.ProgressBar) (error, int64) {
  directory, _  := os.Open(source)
  objects, err  := directory.Readdir(-1)

	if err != nil {
		return err, 0
	}

  for _, obj := range objects {
    sourcefilepointer      := source + "/" + obj.Name()
    destinationfilepointer := dest + "/" + obj.Name()
		if entry.Compress {
			destinationfilepointer = strings.TrimSuffix(destinationfilepointer, ".lz4")
		}

    if obj.IsDir() {
			err,restoredByte = e.RestoreCopyDir(entry, sourcefilepointer, destinationfilepointer, restoredByte, totalRestored, pb)
      if err != nil {
        e.error.Println(err)
				return err, 0
			}
		} else {
			err, byteSource := e.RestoreCopyFile(sourcefilepointer, destinationfilepointer, entry)
			restoredByte    += byteSource
			pb.Show(float32(restoredByte) / float32(totalRestored))
      if err != nil {
        e.error.Println(err)
				return err, 0
			}
		}
	}

	return nil, restoredByte
}


// Copy & Uncompress a specific file if required
func (e *BackupEnv) RestoreCopyFile(source string, dest string, entry *BackupEntry) (error, int64) {
	var (
    sourcefile *os.File
    destfile   *os.File
		err        error
		reader     io.Reader
		writer     io.Writer
	)

  sourcefile, err = os.Open(source);
  if err != nil {
    return err, 0;
  }
  defer sourcefile.Close();

	if entry.Compress {
		reader = lz4.NewReader(sourcefile)
	} else {
		reader = sourcefile
	}

	destfile, err = os.Create(dest);
	writer        = destfile
	if err != nil {
		return err, 0;
	}
	defer destfile.Close();

  _, err = io.Copy(writer, reader)
	if err != nil {
		return err, 0
	}

	sourceinfo, err := os.Stat(source);
	if err != nil {
		return err, 0
	}

  return nil, sourceinfo.Size();
}


func (e *BackupEnv) checkIfDirExist(dir string) (error) {
  _, err := os.Stat(dir);
	return err;
}
