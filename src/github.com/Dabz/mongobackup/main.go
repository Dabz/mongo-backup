/*
** main.go for main.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 10:25:07 2015 gaspar_d
** Last update Fri 25 Dec 02:06:24 2015 gaspar_d
*/

package mongobackup

import (
)

func main() {
  option := parseOptions();
  env    := env{};
  env.setupEnvironment(option);
  env.performBackup();
}
