/*
** backup.go for backup.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 17:39:06 2015 gaspar_d
** Last update Wed 23 Dec 18:48:37 2015 gaspar_d
*/

package main

import (
  "os"
  "log"
)

type env struct {
  options  Options
  homefile *os.File
  trace   *log.Logger
  info    *log.Logger
  warning *log.Logger
  error   *log.Logger
}

func (e *env) setupEnvironment(o Options) {
  traceHandle   := os.Stdout;
  infoHandle    := os.Stdout;
  warningHandle := os.Stdout;
  errorHandle   := os.Stderr;

  e.trace   = log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile);
  e.info    = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile);
  e.warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile);
  e.error   = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile);

  e.options = o;
  e.checkBackupDirectory();
  e.checkHomeFile();
}

func (e *env) checkBackupDirectory() {
  finfo, err := os.Stat(e.options.directory);
  if err != nil {
    os.Mkdir(e.options.directory, 0777);
    finfo, err = os.Stat(e.options.directory);
  }

  if err != nil {
    e.error.Printf("can not create create %s directory (%s)", e.options.directory, err);
    os.Exit(1);
  } else if !finfo.IsDir() {
    e.error.Printf("%s is not a directory", e.options.directory);
    os.Exit(1);
  }
}

func (e *env) checkHomeFile() {
  file, err := os.OpenFile(e.options.directory + "/backup.json", os.O_CREATE | os.O_RDWR, 0777);
  if err != nil {
    e.error.Printf("can not open  %s (%s)", e.options.directory + "/backup.json", err);
    os.Exit(1);
  }

  e.homefile = file;
}
