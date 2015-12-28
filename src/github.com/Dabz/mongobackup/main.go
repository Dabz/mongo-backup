/*
** main.go for main.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 10:25:07 2015 gaspar_d
** Last update Mon 28 Dec 11:41:06 2015 gaspar_d
*/

package main


func main() {
  option := ParseOptions();
  env    := Env{};
  env.SetupEnvironment(option);
  env.PerformBackup();
}
