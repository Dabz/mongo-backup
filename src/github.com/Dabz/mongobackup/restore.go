/*
** restore.go for restore.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Mon 28 Dec 23:33:35 2015 gaspar_d
** Last update Wed 30 Dec 15:05:51 2015 gaspar_d
*/

package main

import (
	"os"
	"io"
	"strings"
  "github.com/pierrec/lz4"
)

// check & perform a partial a global restore
func (e *Env) PerformRestore() {
	if e.options.pit != "" {
		e.perforIncrementalRestore()
	} else if e.options.snapshot != "" && e.options.output != "" {
		entry := e.homeval.GetBackupEntry(e.options.snapshot)
		if entry == nil {
			e.error.Printf("Backup %s can not be found", e.options.snapshot)
			os.Exit(1)
		}
		e. performFullRestore(entry)
	} else {
		e.error.Printf("Invalid configuration")
		os.Exit(1)
	}
}

// perform the restore & dump the oplog if required
func (e *Env) performFullRestore(entry *BackupEntry) {
	var (
		entryFull *BackupEntry
		err       error
		pb        Progessbar
		dirSize   int64
	)
	err = e.checkIfDirExist(e.options.output)
  e.info.Printf("Performing a restore of backup: %s", entry.Id);
	if err != nil {
		e.error.Printf("Can not access directory %s, cowardly failling (%s)", e.options.output, err)
		os.Exit(1)
	}
	if entry.Type == "inc" {
		entryFull = e.homeval.GetLastFullBackup(*entry)
	} else {
		entryFull = entry
	}

  pb.title        = "restoring"
  pb.scale        = 3
	dirSize         = e.GetDirSize(entryFull.Dest)

	pb.Show(0)
	err, restored  := e.restoreCopyDir(entryFull, entryFull.Dest, e.options.output, 0, dirSize, &pb)
	pb.End()
	if err != nil {
		e.error.Printf("Restore of %s failed (%s)", entryFull.Dest, err)
		os.Exit(1)
	}

	if entry.Type == "inc" {
		e.info.Printf("Dumping oplog of the requested snapshots")
		err := e.dumpOplogsToDir(entryFull, entry)
		if err != nil {
			e.error.Printf("Restore of %s failed while dumping oplog (%s)", entryFull.Dest,  err)
			os.Exit(1)
		}
	}

	e.info.Printf("Sucessful restoration, %fGB has been restored to %s", float32(restored) / (1024*1024*1024), e.options.output)
}

// dump the oplog between the entries to the requested output directory
func (e *Env) dumpOplogsToDir(from, to *BackupEntry) error {
	destdir   := e.options.output + "/oplog"
	oplogfile := destdir + "/oplog.bson"
	err       := os.MkdirAll(destdir, 0700)
	if err != nil {
		return err
	}

	destfile, err := os.OpenFile(oplogfile, os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		return err
	}
	destfile.Truncate(0)

	entries := e.homeval.GetEntriesBetween(from, to)
	for _, entry := range entries {
		e.info.Printf("Dumping oplog of backup %s", entry.Id)
		var reader io.Reader
		if entry.Compress {
		  sourcename      := "oplog.bson.lz4"
			sourcefile, err := os.Open(entry.Dest + "/" + sourcename)
			if err != nil {
			  return err
			}
			reader = lz4.NewReader(sourcefile)
		} else {
		  sourcename        := "oplog.bson"
			sourcefile, err   := os.Open(entry.Dest + "/" + sourcename)
			reader             = sourcefile
			if err != nil {
			  return err
			}
		}

		io.Copy(destfile, reader)
	}
	return nil
}

func (e *Env) restoreCopyFile(source string, dest string, entry *BackupEntry) (error, int64) {
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

func (e *Env) restoreCopyDir(entry *BackupEntry, source string, dest string, restoredByte int64, totalRestored int64, pb *Progessbar) (error, int64) {
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
			err,restoredByte = e.restoreCopyDir(entry, sourcefilepointer, destinationfilepointer, restoredByte, totalRestored, pb)
      if err != nil {
        e.error.Println(err)
				return err, 0
			}
		} else {
			err, byteSource := e.restoreCopyFile(sourcefilepointer, destinationfilepointer, entry)
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

func (e *Env) perforIncrementalRestore() {

}

func (e *Env) checkIfDirExist(dir string) (error) {
  _, err := os.Stat(dir);
	return err;
}
