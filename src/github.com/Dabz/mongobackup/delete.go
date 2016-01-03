/*
** delete.go for delete.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Sat  2 Jan 20:36:01 2016 gaspar_d
** Last update Sun  3 Jan 00:39:53 2016 gaspar_d
*/

package main

import (
	"os"
	"errors"
	"strings"
)


// Perform deletion according to the command line options
func (e *Env) PerformDeletion() error {
	if e.options.snapshot != "" {
		err := e.DeleteEntry(e.options.snapshot)
		if err != nil {
			e.error.Printf("Error while deleting backup %s (%s)", e.options.snapshot, err)
			os.Exit(1)
		}
	} else if e.options.position != ""  || e.options.kind != "" {
		err := e.DeleteEntries(e.options.position, e.options.kind)
		if err != nil {
			e.error.Printf("Error while deleting backups (%s)", err)
			os.Exit(1)
		}
	}

	return nil
}

// Remove & delete a specific backup
func (e *Env) DeleteEntry(id string) error {
	entry := e.homeval.GetBackupEntry(id)
	if entry != nil {
		return errors.New("Can not find the backup")
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
func (e *Env) DeleteEntries(criteria, kind string) error {
	var (
		entries []BackupEntry
		ids     []string
		err     error
	)

	err, entries = e.homeval.FindEntries(criteria, kind)
	if err != nil {
		e.error.Printf("Error while retrieving entries (%s)", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		e.info.Printf("Deleting backup %s (%s)", entry.Id, entry.Dest)
		err = e.homeval.RemoveEntry(entry)
		if err != nil {
			e.error.Printf("Error while removing entry from the log file (%s), attempting to continue...", err)
		}
		err = os.RemoveAll(entry.Dest)
		if err != nil {
			e.error.Printf("Error while deleting backup files (%s), attempting to continue...", err)
		}

		ids = append(ids, entry.Id)
	}

	e.info.Printf("Success, backup %s has been deleted", strings.Join(ids, ","))

	return nil
}
