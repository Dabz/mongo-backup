/*
** backup.go for backup.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 17:39:06 2015 gaspar_d
** Last update Fri 25 Dec 02:41:04 2015 gaspar_d
*/

package main

import (
  "os"
  "time"
  "log"
  "gopkg.in/mgo.v2"
)

type env struct {
  options  Options
  homefile *os.File
  trace   *log.Logger
  info    *log.Logger
  warning *log.Logger
  error   *log.Logger
  mongo   *mgo.Session
  dbpath  string
}

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

func (e *env) performBackup() {
  if (! e.options.incremental) {
    e.performFullBackup();
  } else {
    e.perforIncrementalBackup();
  }

}

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
  backupName      := time.Now().Format("20060102150405");
  backupDirectory := e.options.directory + "/" + backupName;

  err, size := e.CopyDir(e.dbpath, backupDirectory);
  sizeGb    := float64(size) / (1024*1024*1024);
  if (err != nil) {
    e.error.Print("An error occurred while backing up ...");
    e.cleanupEnv();
    os.Exit(1);
  }

  e.info.Printf("Success, %fGB of data has been saved in %s", sizeGb, backupDirectory);

  /* End of critical path */
  if (e.options.fsynclock) {
    e.info.Printf("Unlocking the database")
      if (e.mongoFsyncUnLock() != nil) {
        e.cleanupEnv();
        os.Exit(1);
      }
  }
}

func (e *env) perforIncrementalBackup() {

}

func (e *env) cleanupEnv() {
  e.info.Printf("Operation failed, cleaning up the database")
    if (e.options.fsynclock) {
      e.info.Printf("Performing fsyncUnlock");
    }
  e.mongoFsyncUnLock();
}

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

func (e *env) checkHomeFile() {
  file, err := os.OpenFile(e.options.directory + "/backup.json", os.O_CREATE | os.O_RDWR, 0777);
  if err != nil {
    e.error.Printf("can not open  %s (%s)", e.options.directory + "/backup.json", err);
    os.Exit(1);
  }

  e.homefile = file;
}
