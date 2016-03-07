/*
** backup.go for backup.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 17:39:06 2015 gaspar_d
** Last update Mon  7 Mar 16:52:38 2016 gaspar_d
*/

package mongobackup

import (
  "os"
  "time"
  "strconv"
  "gopkg.in/mgo.v2/bson"
)

// perform an incremental or full backup
// check the documentation of the performFullBackup, perforIncrementalBackup
// for more information
func (e *BackupEnv) PerformBackup() {
  backupId         := strconv.Itoa(e.homeval.content.Sequence)
  e.backupdirectory = e.Options.Directory + "/" + backupId;
  e.ensureSecondary();

  if (! e.Options.Incremental) {
    e.performFullBackup(backupId);
  } else {
    e.perforIncrementalBackup(backupId);
  }
}

// perform a full backup
// if compress option is passed, will compress using lz4
// by default will lock the db with fsyncLock
// will perform a rs.stepDown() if the node is primary
func (e *BackupEnv) performFullBackup(backupId string) {
  newEntry := BackupEntry{}
  e.fetchDBPath();
  e.info.Printf("Performing full backup of: %s", e.dbpath);

  if (e.Options.Fsynclock) {
    e.info.Printf("Locking the database")
    if(e.mongoFsyncLock() != nil) {
      e.CleanupBackupEnv();
      os.Exit(1);
    }
  }

  /* Begining critical path */
  err, size := e.CopyDir(e.dbpath, e.backupdirectory);
  sizeGb    := float64(size) / (1024*1024*1024);
  if (err != nil) {
    e.error.Print("An error occurred while backing up ...");
    e.CleanupBackupEnv();
    os.Exit(1);
  }

  /* Dumping oplog for PIT recovery */
  firstOplogEntries := e.getOplogFirstEntries()["ts"].(bson.MongoTimestamp)

  if (firstOplogEntries > e.homeval.lastOplog) {
    e.warning.Printf("Can not find a common point in the oplog")
    e.warning.Printf("point in time restore is not available before this backup")
    newEntry.LastOplog   = e.getOplogLastEntries()["ts"].(bson.MongoTimestamp)
    newEntry.FirstOplog  = firstOplogEntries
  } else {
    cursor            := e.getOplogEntries(e.homeval.lastOplog)
    err, _, fop, lop  := e.BackupOplogToDir(cursor, e.backupdirectory)

    if (err != nil) {
      e.error.Printf("Error while dumping oplog to %s (%s)", e.backupdirectory, err)
      e.CleanupBackupEnv()
      os.Exit(1)
    }

    newEntry.LastOplog       = lop
    newEntry.FirstOplog      = fop
  }

  newEntry.Id     = backupId
  newEntry.Ts     = time.Now()
  newEntry.Source = e.dbpath
  newEntry.Dest   = e.backupdirectory
  newEntry.Kind   = e.Options.Kind
  newEntry.Type   = "full"
  newEntry.Compress        = e.Options.Compress
  e.homeval.AddNewEntry(newEntry)
  e.homeval.Flush()


  e.info.Printf("Success, %fGB of data has been saved in %s", sizeGb, e.backupdirectory);

  /* End of critical path */
  if (e.Options.Fsynclock) {
    e.info.Printf("Unlocking the database")
    if (e.mongoFsyncUnLock() != nil) {
      e.CleanupBackupEnv();
      os.Exit(1);
    }
  }
}

// perform an incremental backup
// oplog greater than the last known oplog will be dump
// if a common point in the oplog can not be found, a
// full backup has to be performed
func (e *BackupEnv) perforIncrementalBackup(backupId string) {
  var (
    lastSavedOplog    bson.MongoTimestamp
    firstOplogEntries bson.MongoTimestamp
  )

  e.info.Printf("Performing incremental backup of: %s", e.Options.Mongohost);

  lastSavedOplog    = e.homeval.lastOplog;
  firstOplogEntries = e.getOplogFirstEntries()["ts"].(bson.MongoTimestamp);

  if (firstOplogEntries > lastSavedOplog) {
    e.error.Printf("Can not find a common point in the oplog");
    e.error.Printf("You must perform a full backup");

    e.CleanupBackupEnv()
    os.Exit(1);
  }

  cursor               := e.getOplogEntries(lastSavedOplog)
  err, size, fop, lop  := e.BackupOplogToDir(cursor, e.backupdirectory)

  if (err != nil) {
    e.error.Printf("Error while dumping oplog to %s (%s)", e.backupdirectory, err)
    e.CleanupBackupEnv()
    os.Exit(1)
  }

  firstOplogEntries = e.getOplogFirstEntries()["ts"].(bson.MongoTimestamp);
  if firstOplogEntries > lastSavedOplog {
    e.warning.Printf("Possible gap in the oplog, last known entry has been reached during the operation")
    e.warning.Printf("if this message appears often, please consider increasing the oplog size")
    e.warning.Printf("https://docs.mongodb.org/manual/tutorial/change-oplog-size/")
  }

  newEntry       := BackupEntry{}
  newEntry.Id     = backupId
  newEntry.Ts     = time.Now()
  newEntry.Source = e.Options.Mongohost
  newEntry.Dest   = e.backupdirectory
  newEntry.Kind   = e.Options.Kind
  newEntry.Type   = "inc"
  newEntry.LastOplog  = lop
  newEntry.FirstOplog = fop
  newEntry.Compress   = e.Options.Compress
  e.homeval.AddNewEntry(newEntry)
  e.homeval.Flush()

  e.info.Printf("Success, %fMB of data has been saved in %s", size / (1024*1024), e.backupdirectory);
}
