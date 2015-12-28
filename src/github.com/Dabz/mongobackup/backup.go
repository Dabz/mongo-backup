/*
** backup.go for backup.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 17:39:06 2015 gaspar_d
** Last update Mon 28 Dec 11:40:10 2015 gaspar_d
*/

package main

import (
  "os"
  "time"
  "gopkg.in/mgo.v2/bson"
)

// perform a backup according to the specified options
func (e *Env) PerformBackup() {
  backupName       := time.Now().Format("20060102150405");
  e.backupdirectory = e.options.directory + "/" + backupName;

  if (! e.options.incremental) {
    e.performFullBackup();
  } else {
    e.perforIncrementalBackup();
  }
}

// perform a full backup
// this is done by doing a filesystem copy of the targeted dbpath
func (e *Env) performFullBackup() {
  e.fetchDBPath();
  e.info.Printf("Performing full backup of: %s", e.dbpath);

  if (e.options.fsynclock) {
    e.info.Printf("Locking the database")
    if(e.mongoFsyncLock() != nil) {
      e.CleanupEnv();
      os.Exit(1);
    }
  }
  /* Begining critical path */
  err, size := e.CopyDir(e.dbpath, e.backupdirectory);
  sizeGb    := float64(size) / (1024*1024*1024);
  if (err != nil) {
    e.error.Print("An error occurred while backing up ...");
    e.CleanupEnv();
    os.Exit(1);
  }

   newEntry       := BackupEntry{};
   newEntry.Ts     = time.Now();
   newEntry.Source = e.dbpath;
   newEntry.Dest   = e.backupdirectory;
   newEntry.Kind   = e.options.kind;
   newEntry.Type   = "full";
   newEntry.LastOplog = e.getOplogLastEntries()["ts"].(bson.MongoTimestamp);
   e.homeval.AddNewEntry(newEntry);


  e.info.Printf("Success, %fGB of data has been saved in %s", sizeGb, e.backupdirectory);

  /* End of critical path */
  if (e.options.fsynclock) {
    e.info.Printf("Unlocking the database")
      if (e.mongoFsyncUnLock() != nil) {
        e.CleanupEnv();
        os.Exit(1);
      }
  }
}

// perform an incremental backup
// oplog greater than the last known oplog will be dump
// to the directory.
// if a common point in the oplog can not be found, a
// full backup has to be performed
func (e *Env) perforIncrementalBackup() {
  var (
    lastSavedOplog    bson.MongoTimestamp
    firstOplogEntries bson.MongoTimestamp
    lastOplogEntry    bson.MongoTimestamp
  )

  e.info.Printf("Performing an incremental backup of: %s", e.options.mongohost);

  lastSavedOplog    = e.homeval.lastOplog;
  firstOplogEntries = e.getOplogFirstEntries()["ts"].(bson.MongoTimestamp);
  lastOplogEntry    = e.getOplogLastEntries()["ts"].(bson.MongoTimestamp);

  if (firstOplogEntries > lastSavedOplog) {
    e.error.Printf("Can not find a common point in the oplog");
    e.error.Printf("You must perform a full backup");

    os.Exit(1);
  }

  cursor    := e.getOplogEntries(lastSavedOplog)
  err, size := e.dumpOplogToDir(cursor, e.backupdirectory)

  if (err != nil) {
    e.error.Printf("Error while dumping oplog to %s (%s)", e.backupdirectory, err)
    e.CleanupEnv()
    os.Exit(1)
  }

   newEntry       := BackupEntry{}
   newEntry.Ts     = time.Now()
   newEntry.Source = e.options.mongohost
   newEntry.Dest   = e.backupdirectory
   newEntry.Kind   = e.options.kind
   newEntry.Type   = "inc"
   newEntry.LastOplog = lastOplogEntry
   newEntry.Compress  = e.options.compress
   e.homeval.AddNewEntry(newEntry)

  e.info.Printf("Success, %fMB of data has been saved in %s", size / (1024*1024), e.backupdirectory);
}
