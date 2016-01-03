/*
** env.go for env.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Mon 28 Dec 11:31:58 2015 gaspar_d
** Last update Sat  2 Jan 20:39:09 2016 gaspar_d
 */

package main

import (
	"gopkg.in/mgo.v2"
	"log"
	"os"
)

// global variable containing options & context informations
type Env struct {
	// represent command line option
	options         Options
	// homelog file & representatino
	homefile        *os.File
	homeval         HomeLogFile
	// logger
	trace           *log.Logger
	info            *log.Logger
	warning         *log.Logger
	error           *log.Logger
	// mongo information
	mongo           *mgo.Session
	dbpath          string
	backupdirectory string
}

// initialize the environment object
func (e *Env) SetupEnvironment(o Options) {
	traceHandle   := os.Stdout
	infoHandle    := os.Stdout
	warningHandle := os.Stdout
	errorHandle   := os.Stderr

	e.trace   = log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	e.info    = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	e.warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	e.error   = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	e.options = o
	e.checkBackupDirectory()
	e.checkHomeFile()
	e.connectMongo()
}


// ensure that the targeted instance is a secondary
// try to perform a rs.stepDown() if it is a primary node
func (e *Env) ensureSecondary() {
	if e.options.stepdown {
		isSec, err := e.mongoIsSecondary()
		if err != nil {
			os.Exit(1)
		}
		if !isSec {
			e.info.Printf("Currently connected to a primary node, performing a rs.stepDown()")
			if e.mongoStepDown() != nil {
				os.Exit(1)
			}
		}
	}
}

// cleanup the environment variable in case of failover
func (e *Env) CleanupEnv() {
	e.info.Printf("Operation failed, cleaning up the database")
	e.info.Printf("Performing fsyncUnlock")
	e.mongoFsyncUnLock()
	e.homefile.Close()
}

// find or create the backup directory
func (e *Env) checkBackupDirectory() {
	finfo, err := os.Stat(e.options.directory)
	if err != nil {
		os.Mkdir(e.options.directory, 0777)
		finfo, err = os.Stat(e.options.directory)
	}

	if err != nil {
		e.error.Printf("can not create create %s directory (%s)", e.options.directory, err)
		os.Exit(1)
	} else if !finfo.IsDir() {
		e.error.Printf("%s is not a directory", e.options.directory)
		os.Exit(1)
	}
}

// find of create the home file
func (e *Env) checkHomeFile() {
	homefile := e.options.directory + "/backup.json"
	_, err := os.Stat(homefile)

	if err != nil {
		e.homefile, err = os.OpenFile(homefile, os.O_CREATE|os.O_RDWR, 0700)
		err = e.homeval.Create(e.homefile)
		if err != nil {
			e.error.Printf("can not create  %s (%s)", homefile, err)
			os.Exit(1)
		}
	} else {
		e.homefile, err = os.OpenFile(homefile, os.O_RDWR, 0700)

		if err != nil {
			e.error.Printf("can not open  %s (%s)", homefile, err)
			os.Exit(1)
		}

		err = e.homeval.Read(e.homefile)

		if err != nil {
			e.error.Printf("can not parse %s (%s)", homefile, err)
			os.Exit(1)
		}
	}
}
