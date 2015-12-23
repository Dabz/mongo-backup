/*
** main.go for main.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 10:25:07 2015 gaspar_d
** Last update Wed 23 Dec 18:48:43 2015 gaspar_d
*/

package main

import (
)

func main() {
  option := parseOptions();
  env    := env{};
  env.setupEnvironment(option);
}
