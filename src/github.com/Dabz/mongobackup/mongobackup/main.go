/*
** main.go for main.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 10:25:07 2015 gaspar_d
** Last update Sun  3 Jan 15:20:11 2016 gaspar_d
*/

package main

import (
	"github.com/Dabz/mongobackup"
)

func main() {
  option := mongobackup.ParseOptions()
  env    := mongobackup.Env{}
  env.SetupEnvironment(option)

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
