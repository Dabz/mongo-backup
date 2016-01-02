/*
** restore.go for restore.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Mon 28 Dec 23:33:35 2015 gaspar_d
** Last update Fri  1 Jan 23:52:42 2016 gaspar_d
*/

package main

import (
	"os"
	"strings"
	"strconv"
	"time"
)

// perform the restore & dump the oplog if required
// oplog is not automatically replayed (futur impprovment?)
// to restore incremental backup or point in time, mongorestore
// has to be used
func (e *Env) PerformRestore() {
	var (
		entry *BackupEntry
	)

	if (e.options.snapshot != "" || e.options.pit != "") && e.options.output != "" {
		if e.options.pit == "" {
		  entry = e.homeval.GetBackupEntry(e.options.snapshot)
			if entry == nil {
				e.error.Printf("Backup %s can not be found", e.options.snapshot)
				os.Exit(1)
			}
	  } else {
			pit   := e.options.pit
			index := strings.Index(pit, ":")
			if index != -1 {
			  pit = pit[:index]
			}

			i, err := strconv.ParseInt(pit, 10, 64)
			if err != nil {
				e.error.Printf("Invalid point in time value: %s (%s)", e.options.pit, err)
				os.Exit(1)
			}
			ts := time.Unix(i, 0)

			entry = e.homeval.GetLastEntryAfter(ts)
			if entry == nil {
				e.error.Printf("A plan to restore to the date %s can not be found", ts)
				os.Exit(1)
			}

			err = e.homeval.CheckIncrementalConsistency(entry)
			if err != nil {
				e.error.Printf("Plan to restore the date %s is inconsistent (%s)", e.options.pit, err)
				os.Exit(1)
			}
		}

		e. performFullRestore(entry)
	} else {
		e.error.Printf("Invalid configuration")
		os.Exit(1)
	}
}

// perform the restore & dump of the oplog
func (e *Env) performFullRestore(entry *BackupEntry) {
	var (
		entryFull *BackupEntry
		err       error
		pb        Progessbar
		dirSize   int64
	)
	err = e.checkIfDirExist(e.options.output)
  e.info.Printf("Performing a restore of backup %s", entry.Id);
	if err != nil {
		e.error.Printf("Can not access directory %s, cowardly failling (%s)", e.options.output, err)
		os.Exit(1)
	}

	if entry.Type == "inc" {
		entryFull = e.homeval.GetLastFullBackup(*entry)
		e.info.Printf("Restoration of backup %s is needed first", entryFull.Id)
	} else {
		entryFull = entry
	}

  pb.title        = "restoring"
  pb.scale        = 3
	dirSize         = e.GetDirSize(entryFull.Dest)

	pb.Show(0)
	err, restored  := e.RestoreCopyDir(entryFull, entryFull.Dest, e.options.output, 0, dirSize, &pb)
	pb.End()
	if err != nil {
		e.error.Printf("Restore of %s failed (%s)", entryFull.Dest, err)
		os.Exit(1)
	}
	e.info.Printf("Sucessful restoration, %fGB has been restored to %s", float32(restored) / (1024*1024*1024), e.options.output)

	if entry.Type == "inc" {
		e.info.Printf("Dumping oplog of the required snapshots")
		err := e.DumpOplogsToDir(entryFull, entry)
		if err != nil {
			e.error.Printf("Restore of %s failed while dumping oplog (%s)", entryFull.Dest,  err)
			os.Exit(1)
		}
		message := "Success. To replay the oplog, start mongod and execute: "
		if e.options.pit == "" {
			message += "`mongorestore --oplogReplay " + e.options.output + "/oplog/`"
		} else {
			message += "`mongorestore --oplogReplay --oplogLimit " +  e.options.pit + " " + e.options.output + "/oplog/`"
		}

		e.info.Printf(message)
	}
}

