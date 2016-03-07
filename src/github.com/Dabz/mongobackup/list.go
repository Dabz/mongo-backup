/*
** list.go for list.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Mon 28 Dec 22:26:20 2015 gaspar_d
** Last update Mon  7 Mar 16:53:45 2016 gaspar_d
*/

package mongobackup

import (
  "fmt"
  "os"
)

// List all backups, if kinf is specified, list only backup with this kind
func (e *BackupEnv) List(kind string) {
  if e.homeval.content.Version == "" {
    e.error.Printf("Can not find a valid home file")
    e.CleanupBackupEnv()
    os.Exit(1)
  }

  err, entries := e.homeval.FindEntries(e.Options.Position, kind)
  if err != nil {
    e.error.Printf("Error while retrieving entries (%s)", err)
    e.CleanupBackupEnv()
    os.Exit(1)
  }

  for _, entry := range entries {
    fmt.Printf("id: %s\tts: %v\tkind: %s\ttype: %s\n", entry.Id, entry.Ts, entry.Kind, entry.Type)
  }
}
