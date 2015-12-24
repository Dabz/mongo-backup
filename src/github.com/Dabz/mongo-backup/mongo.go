/*
** mongo.go for mongo.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 23:59:53 2015 gaspar_d
** Last update Thu 24 Dec 01:19:46 2015 gaspar_d
*/


package main

import (
   "os"
   "gopkg.in/mgo.v2"
   "gopkg.in/mgo.v2/bson"
)

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

func (e *env) fetchRequireInformation() {
  result := bson.M{};
  err     := e.mongo.DB("admin").Run(bson.D{{"getCmdLineOpts", 1}}, &result);
  if (err != nil) {
    e.error.Printf("Can not perform command getCmdLineOpts (%s)", err);
    e.cleanupEnv();
    os.Exit(1);
  }

  e.dbpath = result["parsed"].(bson.M)["storage"].(bson.M)["dbPath"].(string);
}

func (e *env) mongoFsyncLock() {
  result := bson.M{};
  err    := e.mongo.DB("admin").Run(bson.D{{"fsync", 1}, {"lock", true}}, &result);
  if (err != nil) {
    e.error.Printf("Can not perform command fsyncLock (%s)", err);
    e.cleanupEnv();
    os.Exit(1);
  }
}


func (e *env) mongoFsyncUnLock() {
  result := bson.M{};
  err    := e.mongo.DB("admin").Run(bson.D{{"fsyncUnlock", 1}}, &result);
  if (err != nil) {
    e.error.Printf("Can not perform command fsyncUnlock (%s)", err);
    os.Exit(1);
  }
}

func (e *env) mongoIsSecondary() (bool) {
  result := bson.M{};
  err    := e.mongo.DB("admin").Run(bson.D{{"isMaster", 1}}, &result);
  if (err != nil) {
    e.error.Printf("Can not perform command isMaster (%s)", err);
    e.cleanupEnv();
    os.Exit(1);
  }

  return result["secondary"].(bool);
}

func (e *env) mongoStepDown() {
  result := bson.M{};
  err    := e.mongo.DB("admin").Run(bson.D{{"replSetStepDown", 60}}, &result);
  e.mongo.Refresh();
  if (err != nil && ! e.mongoIsSecondary()) {
    e.error.Printf("Can not perform command replSetStepDown (%s)", err);
    e.cleanupEnv();
    os.Exit(1);
  }
}
