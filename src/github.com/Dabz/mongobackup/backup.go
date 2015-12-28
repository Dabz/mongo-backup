/*
** backup.go for backup.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 17:39:06 2015 gaspar_d
** Last update Mon 28 Dec 11:10:13 2015 gaspar_d
*/

package mongobackup

import (
  "os"
  "time"
  "log"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
)

type env struct {
  options            Options
  homefile           *os.File
  homeval            HomeLogFile
  trace              *log.Logger
  info               *log.Logger
  warning            *log.Logger
  error              *log.Logger
  mongo              *mgo.Session
  dbpath             string
  backupdirectory    string
}

// initialize the environment object
func (e *env) setupEnvironment(o Options) {
  traceHandle   := os.Stdout;
  infoHandle    := os.Stdout;
  warningHandle := os.Stdout;
  errorHandle   := os.Stderr;

  e.trace   = log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile);
  e.info    = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile);
  e.warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile);
  e.error   = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile);

  e.options = o;
  e.checkBackupDirectory();
  e.checkHomeFile();
  e.connectMongo();
  e.setupMongo();
}

// connect to mongo
func (e *env) setupMongo() {
  if (e.options.stepdown) {
    isSec, err := e.mongoIsSecondary();
    if (err != nil) {
      os.Exit(1);
    }
    if (! isSec) {
      e.info.Printf("Currently connected to a primary node, performing a rs.stepDown()");
      if (e.mongoStepDown() != nil) {
        os.Exit(1);
      }
    }
  }
}

// perform a backup according to the specified options
func (e *env) performBackup() {
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
func (e *env) performFullBackup() {
  e.fetchDBPath();
  e.info.Printf("Performing full backup of: %s", e.dbpath);

  if (e.options.fsynclock) {
    e.info.Printf("Locking the database")
    if(e.mongoFsyncLock() != nil) {
      e.cleanupEnv();
      os.Exit(1);
    }
  }
  /* Begining critical path */
  err, size := e.CopyDir(e.dbpath, e.backupdirectory);
  sizeGb    := float64(size) / (1024*1024*1024);
  if (err != nil) {
    e.error.Print("An error occurred while backing up ...");
    e.cleanupEnv();
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
        e.cleanupEnv();
        os.Exit(1);
      }
  }
}

// perform an incremental backup
// oplog greater than the last known oplog will be dump
// to the directory.
// if a common point in the oplog can not be found, a
// full backup has to be performed
func (e *env) perforIncrementalBackup() {
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
    e.cleanupEnv()
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

// cleanup the environment variable in case of failover
func (e *env) cleanupEnv() {
  e.info.Printf("Operation failed, cleaning up the database")
    if (e.options.fsynclock) {
      e.info.Printf("Performing fsyncUnlock");
    }
  e.mongoFsyncUnLock();
  e.homefile.Close();
}

// find or create the backup directory
func (e *env) checkBackupDirectory() {
  finfo, err := os.Stat(e.options.directory);
  if err != nil {
    os.Mkdir(e.options.directory, 0777);
    finfo, err = os.Stat(e.options.directory);
  }

  if err != nil {
    e.error.Printf("can not create create %s directory (%s)", e.options.directory, err);
    os.Exit(1);
  } else if !finfo.IsDir() {
    e.error.Printf("%s is not a directory", e.options.directory);
    os.Exit(1);
  }
}

// find of create the home file
func (e *env) checkHomeFile() {
  homefile := e.options.directory + "/backup.json";
  _, err   := os.Stat(homefile);

  if (err != nil) {
    e.homefile, err = os.OpenFile(homefile, os.O_CREATE | os.O_RDWR, 0700);
    err             = e.homeval.Create(e.homefile);
    if err != nil {
      e.error.Printf("can not create  %s (%s)", homefile, err);
      os.Exit(1);
    }
  } else {
    e.homefile, err = os.OpenFile(homefile, os.O_RDWR, 0700);

    if err != nil {
      e.error.Printf("can not open  %s (%s)", homefile, err);
      os.Exit(1);
    }

    err = e.homeval.Read(e.homefile);

    if (err != nil) {
      e.error.Printf("can not parse %s (%s)", homefile, err);
      os.Exit(1);
    }
  }
}
