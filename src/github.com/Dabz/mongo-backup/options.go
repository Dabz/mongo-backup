/*
** options.go for options.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 10:28:29 2015 gaspar_d
** Last update Wed 23 Dec 17:49:45 2015 gaspar_d
*/

package main

import (
   "os"
   "code.google.com/p/getopt"
)

const (
   OP_BACKUP  = 0
   OP_RESTORE = 1
)

type Options struct {
  operation   int
  stepdown    bool
  fsynclock   bool
  incremental bool
  directory   string
}


func parseOptions() (Options) {
  var lineOption Options;

  optDirectory   := getopt.StringLong("dir", 'o', "mongo-backup", "base directory for backup")
  optNoStepdown  := getopt.BoolLong("nostepdown",   0, "perform a rs.stepDown() before the operation")
  optNoFsyncLock := getopt.BoolLong("nofsynclock",  0, "use fsyncLock() and fsyncUnlock()")
  optFull        := getopt.BoolLong("full",       0, "perform a non incremental backup")
  optHelp        := getopt.BoolLong("help",       0,  "Help")
  getopt.SetParameters("backup/restore")

  getopt.Parse()

  if (len(os.Args) - 1  < 1) {
    getopt.Usage();
    os.Exit(1);
  } else if (os.Args[len(os.Args) - 1] == "backup") {
    lineOption.operation = OP_BACKUP;
  } else if (os.Args[len(os.Args) - 1] == "restore") {
    lineOption.operation = OP_RESTORE;
  }

  if (*optHelp) {
    getopt.Usage();
    os.Exit(0);
  }

  lineOption.stepdown    = ! *optNoStepdown;
  lineOption.fsynclock   = ! *optNoFsyncLock;
  lineOption.incremental = ! *optFull;
  lineOption.directory   = *optDirectory;

  if (!validateOptions(lineOption)) {
    getopt.Usage();
    os.Exit(1)
  }

  return lineOption;
}

func validateOptions(o Options) bool {
  return true;
}
