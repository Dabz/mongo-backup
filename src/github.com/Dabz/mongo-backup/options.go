/*
** options.go for options.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 10:28:29 2015 gaspar_d
** Last update Thu 24 Dec 00:04:15 2015 gaspar_d
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
  mongohost  string
  mongouser  string
  mongopwd   string
}


func parseOptions() (Options) {
  var lineOption Options;

  optDirectory   := getopt.StringLong("dir"       , 'o' , "mongo-backup", "base directory for backup")
  optNoStepdown  := getopt.BoolLong("nostepdown"  , 0   , "perform a rs.stepDown() before the operation")
  optNoFsyncLock := getopt.BoolLong("nofsynclock" , 0   , "use fsyncLock() and fsyncUnlock()")
  optFull        := getopt.BoolLong("full"        , 0   , "perform a non incremental backup")
  optHelp        := getopt.BoolLong("help"        , 0   , "Help")

  optMongo       := getopt.StringLong("host" , 'h' , "localhost:27017" , "mongo hostname used for connection");
  optMongoUser   := getopt.StringLong("user" , 'u' , ""                , "mongo username");
  optMongoPwd    := getopt.StringLong("pwd"  , 'p' , ""                , "mongo username password");

  getopt.SetParameters("backup/restore")

  getopt.Parse()

  if (getopt.Arg(0) == "backup") {
    lineOption.operation = OP_BACKUP;
  } else if (getopt.Arg(0) == "restore") {
    lineOption.operation = OP_RESTORE;
  } else {
    getopt.Usage();
    os.Exit(1);
  }

  if (*optHelp) {
    getopt.Usage();
    os.Exit(0);
  }

  lineOption.stepdown    = ! *optNoStepdown;
  lineOption.fsynclock   = ! *optNoFsyncLock;
  lineOption.incremental = ! *optFull;
  lineOption.directory   = *optDirectory;

  lineOption.mongohost = *optMongo;
  lineOption.mongouser = *optMongoUser;
  lineOption.mongopwd  = *optMongoPwd;

  if (!validateOptions(lineOption)) {
    getopt.Usage();
    os.Exit(1)
  }

  return lineOption;
}

func validateOptions(o Options) bool {
  return true;
}
