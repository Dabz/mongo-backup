/*
** main.go for main.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 10:25:07 2015 gaspar_d
** Last update Mon  7 Mar 16:54:09 2016 gaspar_d
*/

package main

import (
  "github.com/Dabz/mongobackup"
  "fmt"
)

func main() {
  option := mongobackup.ParseOptions()
  env    := mongobackup.BackupEnv{}
  err    := env.SetupBackupEnvironment(option)

  if err != nil {
    fmt.Printf("Can not setup program environment (%s)", err)
  }

  if env.Options.Operation == mongobackup.OpBackup {
    env.PerformBackup()
  } else if env.Options.Operation == mongobackup.OpRestore {
    env.PerformRestore()
  } else if env.Options.Operation == mongobackup.OpList {
    env.List(env.Options.Kind)
  } else if env.Options.Operation == mongobackup.OpDelete {
    env.PerformDeletion()
  }
}
