/*
** mongo.go for mongo.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 23:59:53 2015 gaspar_d
** Last update Mon  7 Mar 16:53:48 2016 gaspar_d
*/

package mongobackup

import (
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "errors"
  "fmt"
  "os"
)

// create mongoclient object
func (e *BackupEnv) connectMongo() error {
  var err error
  e.mongo, err = mgo.Dial(e.Options.Mongohost + "?connect=direct")
  if err != nil {
    return errors.New(fmt.Sprintf("Can not connect to %s (%s)", e.Options.Mongohost, err))
  }

  if e.Options.Mongouser != "" && e.Options.Mongopwd != "" {
    err := e.mongo.DB("admin").Login(e.Options.Mongouser, e.Options.Mongopwd)
    if err != nil {
      return errors.New(fmt.Sprintf("Can not login with %s user (%s)", e.Options.Mongouser, err))
    }
  }

  e.mongo.SetMode(mgo.SecondaryPreferred, true)

  return nil
}

// fetch the dbPath of the mongo instance using the
// db.adminCommand({getCmdLineOpts: 1}) command
func (e *BackupEnv) fetchDBPath() {
  result := bson.M{}
  err := e.mongo.DB("admin").Run(bson.D{{"getCmdLineOpts", 1}}, &result)
  if err != nil {
    e.error.Printf("Can not perform command getCmdLineOpts (%s)", err)
    e.CleanupBackupEnv()
    os.Exit(1)
  }

  e.dbpath = result["parsed"].(bson.M)["storage"].(bson.M)["dbPath"].(string)
}

// lock mongodb instance db.fsyncLock()
func (e *BackupEnv) mongoFsyncLock() error {
  result := bson.M{}
  err    := e.mongo.DB("admin").Run(bson.D{{"fsync", 1}, {"lock", true}}, &result)
  if err != nil {
    e.error.Printf("Can not perform command fsyncLock (%s)", err)
  }
  return err
}

// unlock mongodb instance db.fsyncUnlock
func (e *BackupEnv) mongoFsyncUnLock() error {
  result := bson.M{}
  err := e.mongo.DB("admin").C("$cmd.sys.unlock").Find(bson.M{}).One(&result)

  if err != nil {
    e.error.Printf("Can not perform command fsyncUnlock (%s)", err)
  }
  return err
}

// check if a mongodb instance is a secondary
func (e *BackupEnv) mongoIsSecondary() (bool, error) {
  result := bson.M{}
  err    := e.mongo.DB("admin").Run(bson.D{{"isMaster", 1}}, &result)
  if err != nil {
    err = errors.New(fmt.Sprintf("Can not perform command isMaster (%s)", err))
    return false, err
  }

  if result["secondary"] == nil {
    return false, errors.New("Cowardly refusing to perform backup on standalone node. Please check MongoDB architecture guide")
  }

  return result["secondary"].(bool), err
}

// perform an rs.stepDown() on the connected instance
func (e *BackupEnv) mongoStepDown() error {
  result := bson.M{}
  err    := e.mongo.DB("admin").Run(bson.D{{"replSetStepDown", 60}, {"secondaryCatchUpPeriodSecs", 60}}, &result)
  e.mongo.Refresh()
  isSec, err := e.mongoIsSecondary()
  if err != nil && !isSec {
    e.error.Printf("Can not perform command replStepDown (%s)", err)
  }

  return err
}

// get the last oplog entry
func (e *BackupEnv) getOplogLastEntries() bson.M {
  result := bson.M{}
  _       = e.mongo.DB("local").C("oplog.rs").Find(bson.M{}).Sort("-$natural").One(&result)

  return result
}

// get the first oplog entry
func (e *BackupEnv) getOplogFirstEntries() bson.M {
  result := bson.M{}
  _       = e.mongo.DB("local").C("oplog.rs").Find(bson.M{}).Sort("$natural").One(&result)

  return result
}

// get oplog entries that are greater than ts
func (e *BackupEnv) getOplogEntries(ts bson.MongoTimestamp) (iter *mgo.Iter) {
  query := bson.M{"ts": bson.M{"$gte": ts}}
  iter   = e.mongo.DB("local").C("oplog.rs").Find(query).Iter()
  return iter
}

// get the oplog number of entry
func (e *BackupEnv) getOplogCount() int {
  count, _ := e.mongo.DB("local").C("oplog.rs").Count()
  return count
}
