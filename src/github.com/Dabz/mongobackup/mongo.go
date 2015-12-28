/*
** mongo.go for mongo.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 23:59:53 2015 gaspar_d
** Last update Mon 28 Dec 11:15:39 2015 gaspar_d
*/


package mongobackup

import (
   "os"
   "gopkg.in/mgo.v2"
   "gopkg.in/mgo.v2/bson"
)

// create mongoclient object
func (e *env) connectMongo() {
  var err error;
  e.mongo, err = mgo.Dial(e.options.mongohost + "?connect=direct");
  if (err != nil) {
    e.error.Printf("Can not connect to %s (%s)", e.options.mongohost, err);
    e.cleanupEnv();
    os.Exit(1);
  }


  if (e.options.mongouser != "" && e.options.mongopwd != "") {
    err := e.mongo.DB("admin").Login(e.options.mongouser, e.options.mongopwd);
    if (err != nil) {
      e.error.Printf("Can not login with %s user (%s)", e.options.mongouser, err);
      e.cleanupEnv();
      os.Exit(1);
    }
  }

  e.mongo.SetMode(mgo.SecondaryPreferred, true);
}

// fetch the dbPath of the mongo instance using the
// db.adminCommand({getCmdLineOpts: 1}) command
func (e *env) fetchDBPath() {
  result := bson.M{};
  err     := e.mongo.DB("admin").Run(bson.D{{"getCmdLineOpts", 1}}, &result);
  if (err != nil) {
    e.error.Printf("Can not perform command getCmdLineOpts (%s)", err);
    e.cleanupEnv();
    os.Exit(1);
  }

  e.dbpath = result["parsed"].(bson.M)["storage"].(bson.M)["dbPath"].(string);
}

// lock mongodb instance db.fsyncLock()
func (e *env) mongoFsyncLock() (error) {
  result := bson.M{};
  err    := e.mongo.DB("admin").Run(bson.D{{"fsync", 1}, {"lock", true}}, &result);
  if (err != nil) {
    e.error.Printf("Can not perform command fsyncLock (%s)", err);
  }
  return err;
}

// unlock mongodb instance db.fsyncUnlock
func (e *env) mongoFsyncUnLock() (error) {
  result := bson.M{};
  err    := e.mongo.DB("admin").C("$cmd.sys.unlock").Find(bson.M{}).One(&result);

  if (err != nil) {
    e.error.Printf("Can not perform command fsyncUnlock (%s)", err);
  }
  return err;
}

// check if a mongodb instance is a secondary
func (e *env) mongoIsSecondary() (bool, error) {
  result := bson.M{};
  err    := e.mongo.DB("admin").Run(bson.D{{"isMaster", 1}}, &result);
  if (err != nil) {
    e.error.Printf("Can not perform command isMaster (%s)", err);
  }

  return result["secondary"].(bool), err;
}

// perform an rs.stepDown() on the connected instance
func (e *env) mongoStepDown() (error) {
  result := bson.M{};
  err    := e.mongo.DB("admin").Run(bson.D{{"replSetStepDown", 60}}, &result);
  e.mongo.Refresh();

  isSec, err := e.mongoIsSecondary()
  if (err != nil && isSec) {
    e.error.Printf("Can not perform command replSetStepDown (%s)", err);
  }

  return err;
}

// get the last oplog entry
func (e *env) getOplogLastEntries() (bson.M) {
  result := bson.M{};
  _       = e.mongo.DB("local").C("oplog.rs").Find(bson.M{}).Sort("-$natural").One(&result);

  return result;
}

// get the first oplog entry
func (e *env) getOplogFirstEntries() (bson.M) {
  result := bson.M{};
  _       = e.mongo.DB("local").C("oplog.rs").Find(bson.M{}).Sort("$natural").One(&result);

  return result;
}

// get oplog entries that are greater than ts
func (e *env) getOplogEntries(ts bson.MongoTimestamp) (iter *mgo.Iter) {
  query := bson.M{"ts": bson.M{"$gt": ts}};
  iter   = e.mongo.DB("local").C("oplog.rs").Find(query).Iter()
  return iter;
}

// get the oplog number of entry
func (e *env) getOplogCount() (int) {
  count, _ := e.mongo.DB("local").C("oplog.rs").Count();
  return count;
}
