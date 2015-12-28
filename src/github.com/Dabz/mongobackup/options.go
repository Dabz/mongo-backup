/*
** options.go for options.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 10:28:29 2015 gaspar_d
** Last update Mon 28 Dec 23:52:51 2015 gaspar_d
 */

package main

import (
	"code.google.com/p/getopt"
	"os"
)

const (
	OP_BACKUP  = 0
	OP_RESTORE = 1
	OP_LIST    = 4

	DEFAULT_KIND = "backup"
	DEFAULT_DIR  = "mongo-backup"
)

// represent the command lines option
type Options struct {
	// heneral options
	operation   int
	directory   string
	kind        string
	stepdown    bool
	// backup options
	fsynclock   bool
	incremental bool
	compress    bool
	// mongo options
	mongohost   string
	mongouser   string
	mongopwd    string
	// restore options
	output      string
	pit         string
	snapshot    string
}

// parse the command line and create the Options struct
func ParseOptions() Options {
	var lineOption Options

	optDirectory   := getopt.StringLong("basedir", 'b', DEFAULT_DIR, "base directory to save & restore backup")
	optKind        := getopt.StringLong("kind", 'k', DEFAULT_KIND, "metadata associated to the backup")
	optNoStepdown  := getopt.BoolLong("nostepdown", 0, "no rs.stepDown() if this is the primary node")
	optNoFsyncLock := getopt.BoolLong("nofsynclock", 0, "Avoid using fsyncLock() and fsyncUnlock()")
	optNoCompress  := getopt.BoolLong("nocompress", 0, "disable compression for backup & restore")
	optFull        := getopt.BoolLong("full", 0, "perform a non incremental backup")
	optHelp        := getopt.BoolLong("help", 0, "")

	optMongo     := getopt.StringLong("host", 'h', "localhost:27017", "mongo hostname")
	optMongoUser := getopt.StringLong("user", 'u', "", "mongo username")
	optMongoPwd  := getopt.StringLong("pwd", 'p', "", "mongo password")

	optPitTime   := getopt.StringLong("pit", 't', "", "point in time recovery (using oplog format: unixtimetamp:opcount)")
	optSnapshot  := getopt.StringLong("snapshot", 's', "", "backup to restore")
	optOutput    := getopt.StringLong("out", 'o', "", "output directory")

	getopt.SetParameters("backup|restore|list")

	getopt.Parse()

	if getopt.Arg(0) == "backup" {
		lineOption.operation = OP_BACKUP
	} else if getopt.Arg(0) == "restore" {
		lineOption.operation = OP_RESTORE
  } else if getopt.Arg(0) == "list" {
		lineOption.operation = OP_LIST
	} else {
		getopt.Usage()
		os.Exit(1)
	}

	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	}

	lineOption.stepdown    = !*optNoStepdown
	lineOption.fsynclock   = !*optNoFsyncLock
	lineOption.incremental = !*optFull
	lineOption.directory   = *optDirectory
	lineOption.compress    = !*optNoCompress

	lineOption.mongohost = *optMongo
	lineOption.mongouser = *optMongoUser
	lineOption.mongopwd  = *optMongoPwd
	lineOption.kind      = *optKind
	lineOption.pit       = *optPitTime
	lineOption.snapshot  = *optSnapshot
	lineOption.output    = *optOutput

	if !validateOptions(lineOption) {
		getopt.Usage()
		os.Exit(1)
	}

	return lineOption
}

// validate the option to see if there is
// any incoherence (TODO)
func validateOptions(o Options) bool {
	return true
}
