/*
** main.go for main.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 10:25:07 2015 gaspar_d
** Last update Mon 28 Dec 23:51:35 2015 gaspar_d
*/

package main


func main() {
  option := ParseOptions()
  env    := Env{}
  env.SetupEnvironment(option)
	if env.options.operation == OP_BACKUP {
    env.PerformBackup()
	} else if env.options.operation == OP_RESTORE {
		env.PerformRestore()
  } else if env.options.operation == OP_LIST {
		env.List(env.options.kind)
	}
}
