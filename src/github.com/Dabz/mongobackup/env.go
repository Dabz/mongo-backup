/*
** env.go for env.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Mon 28 Dec 11:31:58 2015 gaspar_d
** Last update Mon  7 Mar 16:52:59 2016 gaspar_d
*/

package mongobackup

import (
  "github.com/nightlyone/lockfile"
  "gopkg.in/mgo.v2"
  "io/ioutil"
  "time"
  "log"
  "os"
)

const (
  LockFileName = "backup.lock"
)

// global variable containing options & context informations
type BackupEnv struct {
  // represent command line option
  Options         Options
  // homelog file & representatino
  homefile        *os.File
  homeval         BackupHistoryFile
  // logger
  trace           *log.Logger
  info            *log.Logger
  warning         *log.Logger
  error           *log.Logger
  // mongo information
  mongo           *mgo.Session
  dbpath          string
  backupdirectory string
  // lock file
  lock            lockfile.Lockfile
}

// initialize the environment object
func (e *BackupEnv) SetupBackupEnvironment(o Options) error {
  if o.Debug {
    traceHandle   := os.Stdout
    infoHandle    := os.Stdout
    warningHandle := os.Stdout
    errorHandle   := os.Stderr

    e.trace   = log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
    e.info    = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    e.warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
    e.error   = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
  } else {
    traceHandle   := ioutil.Discard
    infoHandle    := os.Stdout
    warningHandle := os.Stdout
    errorHandle   := os.Stderr

    e.trace   = log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime)
    e.info    = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime)
    e.warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime)
    e.error   = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime)
  }

  e.lock, _ = lockfile.New(o.Directory + "/" + LockFileName)
  err      := e.lock.TryLock()
  if err == lockfile.ErrBusy {
    try := 1
    for {
      e.warning.Printf("Another process seems to be running, retrying (%d/5) ...", try)
      time.Sleep(10 * time.Duration(try) * time.Second)
      err  = e.lock.TryLock()
      if err == nil {
        break
      }
      try += 1
      if try > 5 {
        e.CleanupBackupEnv()
        return err
      }
    }
  }

  e.Options = o
  e.checkBackupDirectory()
  e.checkHomeFile()
  err = e.connectMongo()
  if err != nil {
    e.error.Printf("Error while connecting to mongo (%s)", err)
    return err
  }

  return nil
}


// ensure that the targeted instance is a secondary
// try to perform a rs.stepDown() if it is a primary node
func (e *BackupEnv) ensureSecondary() {
  if e.Options.Stepdown {
    isSec, err := e.mongoIsSecondary()
    if err != nil {
      e.error.Printf("Error while checking if the node is primary (%s)", err)
      e.CleanupBackupEnv()
      os.Exit(1)
    }
    if !isSec {
      e.info.Printf("Currently connected to a primary node, performing a rs.stepDown()")
      if e.mongoStepDown() != nil {
        e.CleanupBackupEnv()
        os.Exit(1)
      }
    }
  }
}

// cleanup the environment variable in case of failover
func (e *BackupEnv) CleanupBackupEnv() {
  if e.mongo != nil {
    e.info.Printf("Performing fsyncUnlock")
    e.mongoFsyncUnLock()
    e.homefile.Close()
  }
  e.lock.Unlock()
}

// find or create the backup directory
func (e *BackupEnv) checkBackupDirectory() {
  finfo, err := os.Stat(e.Options.Directory)
  if err != nil {
    os.Mkdir(e.Options.Directory, 0777)
    finfo, err = os.Stat(e.Options.Directory)
  }

  if err != nil {
    e.error.Printf("can not create create %s directory (%s)", e.Options.Directory, err)
    e.CleanupBackupEnv()
    os.Exit(1)
  } else if !finfo.IsDir() {
    e.error.Printf("%s is not a directory", e.Options.Directory)
    e.CleanupBackupEnv()
    os.Exit(1)
  }
}

// find of create the home file
func (e *BackupEnv) checkHomeFile() {
  homefile := e.Options.Directory + "/backup.json"
  _, err := os.Stat(homefile)

  if err != nil {
    e.homefile, err = os.OpenFile(homefile, os.O_CREATE|os.O_RDWR, 0700)
    err = e.homeval.Create(e.homefile)
    e.homeval.Flush()
    if err != nil {
      e.error.Printf("can not create  %s (%s)", homefile, err)
      e.CleanupBackupEnv()
      os.Exit(1)
    }
  } else {
    e.homefile, err = os.OpenFile(homefile, os.O_RDWR, 0700)

    if err != nil {
      e.error.Printf("can not open  %s (%s)", homefile, err)
      e.CleanupBackupEnv()
      os.Exit(1)
    }

    err = e.homeval.Read(e.homefile)

    if err != nil {
      e.error.Printf("can not parse %s (%s)", homefile, err)
      e.CleanupBackupEnv()
      os.Exit(1)
    }
  }
}
