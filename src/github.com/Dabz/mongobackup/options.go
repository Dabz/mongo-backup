/*
** options.go for options.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 10:28:29 2015 gaspar_d
** Last update Fri  1 Jan 03:09:12 2016 gaspar_d
 */

package main

import (
	"code.google.com/p/getopt"
	"os"
	"fmt"
)

const (
	OP_BACKUP  = 0
	OP_RESTORE = 1
	OP_LIST    = 4

	DEFAULT_KIND = "backup"
	DEFAULT_DIR  = "mongo-backup"
)


// abstract structure standing for command line options
type Options struct {
	// general options
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
	var (
		lineOption Options
		set        *getopt.Set
	)

	set = getopt.New()

	optDirectory   := set.StringLong("basedir", 'b', DEFAULT_DIR, "base directory to save & restore backup")
	optKind        := set.StringLong("kind", 'k', DEFAULT_KIND, "metadata associated to the backup")
	optNoStepdown  := set.BoolLong("nostepdown", 0, "no rs.stepDown() if this is the primary node")
	optNoFsyncLock := set.BoolLong("nofsynclock", 0, "Avoid using fsyncLock() and fsyncUnlock()")
	optNoCompress  := set.BoolLong("nocompress", 0, "disable compression for backup & restore")
	optFull        := set.BoolLong("full", 0, "perform a non incremental backup")
	optHelp        := set.BoolLong("help", 'h', "")

	optMongo     := set.StringLong("host", 0, "localhost:27017", "mongo hostname")
	optMongoUser := set.StringLong("username", 'u', "", "mongo username")
	optMongoPwd  := set.StringLong("password", 'p', "", "mongo password")

	optPitTime   := set.StringLong("pit", 0, "", "point in time recovery (using oplog format: unixtimetamp:opcount)")
	optSnapshot  := set.StringLong("snapshot", 0, "", "backup to restore")
	optOutput    := set.StringLong("out", 'o', "", "output directory")

	set.SetParameters("backup|restore|list")

	err := set.Getopt(os.Args[1:], nil);
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		set.PrintUsage(os.Stdout)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		PrintHelp()
		os.Exit(1)
	} else if os.Args[1] == "backup" {
		lineOption.operation = OP_BACKUP
	} else if os.Args[1] == "restore" {
		lineOption.operation = OP_RESTORE
  } else if os.Args[1] == "list" {
		lineOption.operation = OP_LIST
  } else if os.Args[1] == "help"  || (*optHelp) {
		PrintHelp()
		os.Exit(0)
	} else {
		PrintHelp()
		os.Exit(1)
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

func PrintHelp() {
	var helpMessage []string
	helpMessage = append(helpMessage,  "--basedir=string  base directory to save & restore backup")
	helpMessage = append(helpMessage,  "--kind=string  metadata associated to the backup")
	helpMessage = append(helpMessage,  "--nostepdown  no rs.stepDown() if this is the primary node")
	helpMessage = append(helpMessage,  "--nofsynclock  Avoid using fsyncLock() and fsyncUnlock()")
	helpMessage = append(helpMessage,  "--nocompress  disable compression for backup & restore")
	helpMessage = append(helpMessage,  "--full  perform a non incremental backup")
	helpMessage = append(helpMessage,  "--host=string  mongo hostname")
	helpMessage = append(helpMessage,  "--username=string  mongo username")
	helpMessage = append(helpMessage,  "--password=string  mongo password")
	helpMessage = append(helpMessage,  "--pit=string  point in time recovery (using oplog format: unixtimetamp:opcount)")
	helpMessage = append(helpMessage,  "--snapshot=string  to restore a specific backup")
	helpMessage = append(helpMessage,  "--out=string  output directory")


	fmt.Printf("\nUsage:\n\n    %s command options\n", os.Args[0])

	fmt.Printf("\n")
	fmt.Printf("Commands:\n")
	fmt.Printf("\n")
	fmt.Printf("    %s backup [--full] [--kind string] [--nocompress] [--nofsynclock] [--nostepdown]\n", os.Args[0])
	fmt.Printf("    %s restore --out string [--snapshot string] [--pit string]\n", os.Args[0])
	fmt.Printf("    %s list [--kind string]\n", os.Args[0])
	fmt.Printf("\n")
	fmt.Printf("Options:\n")
	fmt.Printf("\n")

	for _, help := range helpMessage {
		fmt.Print("    ")
		fmt.Print(help)
		fmt.Print("\n")
	}
}
