/*
** list.go for list.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Mon 28 Dec 22:26:20 2015 gaspar_d
** Last update Mon 28 Dec 23:32:19 2015 gaspar_d
 */

package main

import (
	"fmt"
	"os"
)

// List all backups, if kinf is specified, list only backup with this kind
func (e *Env) List(kind string) {
	if e.homeval.content.Version == "" {
		e.error.Printf("Can not find a valid home file")
		os.Exit(1)
	}

	entries := e.homeval.content.Entries

	for _, entry := range entries {
		if kind == DEFAULT_KIND || entry.Kind == kind {
			fmt.Printf("ts: %s\tlocation: %s\tkind: %s\ttype: %s\n", entry.Ts, entry.Dest, entry.Kind, entry.Type)
		}
	}
}
