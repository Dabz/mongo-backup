/*
** options.go for options.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 10:28:29 2015 gaspar_d
** Last update Mon  7 Mar 16:53:55 2016 gaspar_d
*/

package mongobackup

import (
  "github.com/pborman/getopt"
  "os"
  "fmt"
)

const (
  OpBackup  = 0
  OpRestore = 1
  OpList    = 4
  OpDelete  = 8

  DefaultKind = "backup"
  DefaultDir  = "mongo-backup"
)


// abstract structure standing for command line options
type Options struct {
  // general options
  Operation   int
  Directory   string
  Kind        string
  Stepdown    bool
  Position    string
  Debug       bool
  // backup options
  Fsynclock   bool
  Incremental bool
  Compress    bool
  // mongo options
  Mongohost   string
  Mongouser   string
  Mongopwd    string
  // restore options
  Output      string
  Pit         string
  Snapshot    string
}

// parse the command line and create the Options struct
func ParseOptions() Options {
  var (
    lineOption Options
    set        *getopt.Set
  )

  set     = getopt.New()
  pwd, _ := os.Getwd()

  optDirectory   := set.StringLong("basedir", 'b', pwd + "/" + DefaultDir, "")
  optKind        := set.StringLong("kind", 'k', DefaultKind, "")
  optNoStepdown  := set.BoolLong("nostepdown", 0, "")
  optNoFsyncLock := set.BoolLong("nofsynclock", 0, "")
  optNoCompress  := set.BoolLong("nocompress", 0, "")
  optFull        := set.BoolLong("full", 0, "")
  optHelp        := set.BoolLong("help", 'h', "")
  optDebug       := set.BoolLong("debug", 'd', "")

  optMongo     := set.StringLong("host", 0, "localhost:27017", "")
  optMongoUser := set.StringLong("username", 'u', "", "")
  optMongoPwd  := set.StringLong("password", 'p', "", "")

  optPitTime   := set.StringLong("pit", 0, "", "")
  optSnapshot  := set.StringLong("snapshot", 0, "", "")
  optOutput    := set.StringLong("out", 'o', "", "")

  optPosition  := set.StringLong("entries", 0, "", "")

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
    lineOption.Operation = OpBackup
  } else if os.Args[1] == "restore" {
    lineOption.Operation = OpRestore
  } else if os.Args[1] == "list" {
    lineOption.Operation = OpList
  } else if os.Args[1] == "delete" {
    lineOption.Operation = OpDelete
  } else if os.Args[1] == "help"  || (*optHelp) {
    PrintHelp()
    os.Exit(0)
  } else {
    PrintHelp()
    os.Exit(1)
  }

  lineOption.Stepdown    = !*optNoStepdown
  lineOption.Fsynclock   = !*optNoFsyncLock
  lineOption.Incremental = !*optFull
  lineOption.Directory   = *optDirectory
  lineOption.Compress    = !*optNoCompress
  lineOption.Debug       = *optDebug

  lineOption.Mongohost = *optMongo
  lineOption.Mongouser = *optMongoUser
  lineOption.Mongopwd  = *optMongoPwd
  lineOption.Kind      = *optKind
  lineOption.Pit       = *optPitTime
  lineOption.Snapshot  = *optSnapshot
  lineOption.Output    = *optOutput
  lineOption.Position  = *optPosition

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

  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", "-b", "--basedir=string", "base directory to save & restore backup"))
  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", "-k", "--kind=string", "metadata associated to the backup"))
  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", ""  , "--nostepdown", "no rs.stepDown() if this is the primary node"))
  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", ""  , "--nofsynclock", "Avoid using fsyncLock() and fsyncUnlock()"))
  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", ""  , "--nocompress", "disable compression for backup & restore"))
  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", ""  , "--full", "perform a non incremental backup"))
  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", ""  , "--host=string", "mongo hostname"))
  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", "-u", "--username=string", "mongo username"))
  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", "-p", "--password=string", "mongo password"))
  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", ""  , "--pit=string", "point in time recovery (using oplog format: unixtimetamp:opcount)"))
  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", ""  , "--snapshot=string", "to restore a specific backup"))
  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", "-o", "--out=string", "output directory"))
  helpMessage = append(helpMessage,  fmt.Sprintf("%-5s %-20s %s", ""  ,"--entries=string", "criteria string (format number[+-])"))


  fmt.Printf("\nUsage:\n\n    %s command options\n", os.Args[0])

  fmt.Printf("\n")
  fmt.Printf("Commands:\n")
  fmt.Printf("\n")
  fmt.Printf("    %-35s %s %s\n", "perform an incremental backup", os.Args[0], "backup [--kind string] [--nocompress] [--nofsynclock] [--nostepdown]")
  fmt.Printf("    %-35s %s %s\n", "perform a full backup", os.Args[0], "backup --full [--kind string] [--nocompress] [--nofsynclock] [--nostepdown]")
  fmt.Printf("    %-35s %s %s\n", "restore a specific backup", os.Args[0], "restore --out string --snapshot string")
  fmt.Printf("    %-35s %s %s\n", "perform a point in time restore", os.Args[0], "restore --out string --pit string")
  fmt.Printf("    %-35s %s %s\n", "delete a range of backup", os.Args[0], "delete --kind string --entries string")
  fmt.Printf("    %-35s %s %s\n", "delete a specific backup", os.Args[0], "delete --snapshot string")
  fmt.Printf("    %-35s %s %s\n", "list available backups", os.Args[0], "list [--kind string] [--entries string]")
  fmt.Printf("\n")
  fmt.Printf("Options:\n")
  fmt.Printf("\n")

  for _, help := range helpMessage {
    fmt.Print("    ")
    fmt.Print(help)
    fmt.Print("\n")
  }
}
