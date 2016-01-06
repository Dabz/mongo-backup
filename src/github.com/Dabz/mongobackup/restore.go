/*
** restore.go for restore.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Mon 28 Dec 23:33:35 2015 gaspar_d
** Last update Wed  6 Jan 20:05:22 2016 gaspar_d
*/

package mongobackup

import (
	"os"
	"strings"
	"strconv"
	"time"
	"github.com/Dabz/utils"
)

// perform the restore & dump the oplog if required
// oplog is not automatically replayed (futur impprovment?)
// to restore incremental backup or point in time, mongorestore
// has to be used
func (e *BackupEnv) PerformRestore() {
	var (
		entry *BackupEntry
	)

	if (e.Options.Snapshot != "" || e.Options.Pit != "") && e.Options.Output != "" {
		if e.Options.Pit == "" {
		  entry = e.homeval.GetBackupEntry(e.Options.Snapshot)
			if entry == nil {
				e.error.Printf("Backup %s can not be found", e.Options.Snapshot)
				e.CleanupBackupEnv()
				os.Exit(1)
			}
	  } else {
			pit   := e.Options.Pit
			index := strings.Index(pit, ":")
			if index != -1 {
			  pit = pit[:index]
			}

			i, err := strconv.ParseInt(pit, 10, 64)
			if err != nil {
				e.error.Printf("Invalid point in time value: %s (%s)", e.Options.Pit, err)
				e.CleanupBackupEnv()
				os.Exit(1)
			}
			ts := time.Unix(i, 0)

			entry = e.homeval.GetLastEntryAfter(ts)
			if entry == nil {
				e.error.Printf("A plan to restore to the date %s can not be found", ts)
				e.CleanupBackupEnv()
				os.Exit(1)
			}

			err = e.homeval.CheckIncrementalConsistency(entry)
			if err != nil {
				e.error.Printf("Plan to restore the date %s is inconsistent (%s)", e.Options.Pit, err)
				e.CleanupBackupEnv()
				os.Exit(1)
			}
		}

		e. performFullRestore(entry)
	} else {
		e.error.Printf("Invalid configuration")
		e.CleanupBackupEnv()
		os.Exit(1)
	}
}

// perform the restore & dump of the oplog
func (e *BackupEnv) performFullRestore(entry *BackupEntry) {
	var (
		entryFull *BackupEntry
		err       error
		pb        utils.ProgressBar
		dirSize   int64
	)
	err = e.checkIfDirExist(e.Options.Output)
  e.info.Printf("Performing a restore of backup %s", entry.Id);
	if err != nil {
		e.error.Printf("Can not access directory %s, cowardly failling (%s)", e.Options.Output, err)
		e.CleanupBackupEnv()
		os.Exit(1)
	}

	if entry.Type == "inc" {
		entryFull = e.homeval.GetLastFullBackup(*entry)
		if entryFull == nil {
			e.error.Printf("Error, can not retrieve a valid full backup before incremental backup %s", entry.Id)
			e.CleanupBackupEnv()
			os.Exit(1)
		}
		e.info.Printf("Restoration of backup %s is needed first", entryFull.Id)
	} else {
		entryFull = entry
	}

  pb.Title = "restoring"
  pb.Scale = 3
	dirSize  = e.GetDirSize(entryFull.Dest)

	pb.Show(0)
	err, restored  := e.RestoreCopyDir(entryFull, entryFull.Dest, e.Options.Output, 0, dirSize, &pb)
	pb.End()
	if err != nil {
		e.error.Printf("Restore of %s failed (%s)", entryFull.Dest, err)
		e.CleanupBackupEnv()
		os.Exit(1)
	}
	e.info.Printf("Sucessful restoration, %fGB has been restored to %s", float32(restored) / (1024*1024*1024), e.Options.Output)

	if entry.Type == "inc" {
		e.info.Printf("Dumping oplog of the required snapshots")
		err := e.DumpOplogsToDir(entryFull, entry)
		if err != nil {
			e.error.Printf("Restore of %s failed while dumping oplog (%s)", entryFull.Dest,  err)
			e.CleanupBackupEnv()
			os.Exit(1)
		}
		message := "Success. To replay the oplog, start mongod and execute: "
		if e.Options.Pit == "" {
			message += "`mongorestore --oplogReplay " + e.Options.Output + "/oplog/`"
		} else {
			message += "`mongorestore --oplogReplay --oplogLimit " +  e.Options.Pit + " " + e.Options.Output + "/oplog/`"
		}

		e.info.Printf(message)
	}
}

