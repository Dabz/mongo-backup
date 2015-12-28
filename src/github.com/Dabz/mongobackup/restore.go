/*
** restore.go for restore.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Mon 28 Dec 23:33:35 2015 gaspar_d
** Last update Tue 29 Dec 00:02:12 2015 gaspar_d
*/

package main

import (
	"os"
)

func (e *Env) PerformRestore() {
	if e.options.pit != "" {
		e.perforIncrementalRestore()
	} else if e.options.snapshot != "" && e.options.output != "" {
		e.performFullRestore()
	} else {
		e.error.Printf("Invalid configuration")
		os.Exit(1)
	}
}

func (e *Env) performFullRestore() {
	err := e.checkIfDirExist(e.options.output)
	if err != nil {
		e.error.Printf("Can not access directory %s, cowardly failling (%s)", e.options.output, err)
		os.Exit(1)
	}
}

func (e *Env) perforIncrementalRestore() {

}

func (e *Env) checkIfDirExist(dir string) (error) {
  _, err := os.Stat(dir);
	return err;
}
