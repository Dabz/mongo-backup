/*
** delete.go for delete.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Sat  2 Jan 20:36:01 2016 gaspar_d
** Last update Mon  7 Mar 16:52:50 2016 gaspar_d
*/

package mongobackup

import (
  "os"
  "errors"
  "strings"
)


// Perform deletion according to the command line options
func (e *BackupEnv) PerformDeletion() error {
  if e.Options.Snapshot != "" {
    err := e.DeleteEntry(e.Options.Snapshot)
    e.homeval.Flush()
    if err != nil {
      e.error.Printf("Error while deleting backup %s (%s)", e.Options.Snapshot, err)
      e.CleanupBackupEnv()
      os.Exit(1)
    }
  } else if e.Options.Position != ""  || e.Options.Kind != "" {
    err := e.DeleteEntries(e.Options.Position, e.Options.Kind)
    if err != nil {
      e.error.Printf("Error while deleting backups (%s)", err)
      e.CleanupBackupEnv()
      os.Exit(1)
    }
  }

  return nil
}

// Remove & delete a specific backup
func (e *BackupEnv) DeleteEntry(id string) error {
  entry := e.homeval.GetBackupEntry(id)
  if entry == nil {
    return errors.New("Can not find the backup " + id)
  }

  err := e.homeval.RemoveEntry(*entry)
  if err != nil {
    e.error.Printf("Error while removing entry from the log file (%s), attempting to continue...", err)
  }

  err = os.RemoveAll(entry.Dest)
  if err != nil {
    e.error.Printf("Error while deleting backup files (%s), attempting to continue...", err)
  }

  return nil
}

// Remove & delete a range of backups
func (e *BackupEnv) DeleteEntries(criteria, kind string) error {
  var (
    entries     []BackupEntry
    ids         []string
    err         error
  )

  err, entries = e.homeval.FindEntries(criteria, kind)
  if err != nil {
    e.error.Printf("Error while retrieving entries (%s)", err)
    e.CleanupBackupEnv()
    os.Exit(1)
  }

  // fetch first & alst full backup
  firstFullBackup := BackupEntry{}
  lastFullBackup  := BackupEntry{}
  for _, entry := range entries {
    if firstFullBackup.Id == "" && entry.Type == "full" {
      firstFullBackup = entry
    }
    if entry.Type == "full" {
      lastFullBackup = entry
    }
  }

  // check that there is a range of backup available for deletion
  if firstFullBackup.Id == "" || firstFullBackup.Id == lastFullBackup.Id {
    e.warning.Printf("Cowardly not deleting backups as there is incremental backup depending on it")
    return nil
  }

  // let's delete them
  for _, entry := range entries {
    // we should not delete the last entry...
    if entry.Type == "full" && entry.Id != lastFullBackup.Id {
      err = e.homeval.RemoveEntry(entry)
      e.homeval.Flush()
      if err != nil {
        e.error.Printf("Error while removing entry from the log file (%s), attempting to continue...", err)
      }
      err = os.RemoveAll(entry.Dest)
      if err != nil {
        e.error.Printf("Error while deleting backup files (%s), attempting to continue...", err)
      }
      ids = append(ids, entry.Id)
    } else if entry.Type == "inc" { // if incremental backup, let's delete it right away
      err = e.homeval.RemoveEntry(entry)
      e.homeval.Flush()
      if err != nil {
        e.error.Printf("Error while removing entry from the log file (%s), attempting to continue...", err)
      }
      err = os.RemoveAll(entry.Dest)
      if err != nil {
        e.error.Printf("Error while deleting backup files (%s), attempting to continue...", err)
      }
      ids = append(ids, entry.Id)
    }
  }

  e.info.Printf("Success, backup %s has been deleted", strings.Join(ids, ","))

  return nil
}
