/*
** mongo.go for mongo.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Wed 23 Dec 23:59:53 2015 gaspar_d
** Last update Sun  3 Jan 15:16:24 2016 gaspar_d
 */

package mongobackup

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
)

// create mongoclient object
func (e *Env) connectMongo() {
	var err error
	e.mongo, err = mgo.Dial(e.Options.Mongohost + "?connect=direct")
	if err != nil {
		e.error.Printf("Can not connect to %s (%s)", e.Options.Mongohost, err)
		e.CleanupEnv()
		os.Exit(1)
	}

	if e.Options.Mongouser != "" && e.Options.Mongopwd != "" {
		err := e.mongo.DB("admin").Login(e.Options.Mongouser, e.Options.Mongopwd)
		if err != nil {
			e.error.Printf("Can not login with %s user (%s)", e.Options.Mongouser, err)
			e.CleanupEnv()
			os.Exit(1)
		}
	}

	e.mongo.SetMode(mgo.SecondaryPreferred, true)
}

// fetch the dbPath of the mongo instance using the
// db.adminCommand({getCmdLineOpts: 1}) command
func (e *Env) fetchDBPath() {
	result := bson.M{}
	err := e.mongo.DB("admin").Run(bson.D{{"getCmdLineOpts", 1}}, &result)
	if err != nil {
		e.error.Printf("Can not perform command getCmdLineOpts (%s)", err)
		e.CleanupEnv()
		os.Exit(1)
	}

	e.dbpath = result["parsed"].(bson.M)["storage"].(bson.M)["dbPath"].(string)
}

// lock mongodb instance db.fsyncLock()
func (e *Env) mongoFsyncLock() error {
	result := bson.M{}
	err    := e.mongo.DB("admin").Run(bson.D{{"fsync", 1}, {"lock", true}}, &result)
	if err != nil {
		e.error.Printf("Can not perform command fsyncLock (%s)", err)
	}
	return err
}

// unlock mongodb instance db.fsyncUnlock
func (e *Env) mongoFsyncUnLock() error {
	result := bson.M{}
	err := e.mongo.DB("admin").C("$cmd.sys.unlock").Find(bson.M{}).One(&result)

	if err != nil {
		e.error.Printf("Can not perform command fsyncUnlock (%s)", err)
	}
	return err
}

// check if a mongodb instance is a secondary
func (e *Env) mongoIsSecondary() (bool, error) {
	result := bson.M{}
	err    := e.mongo.DB("admin").Run(bson.D{{"isMaster", 1}}, &result)
	if err != nil {
		e.error.Printf("Can not perform command isMaster (%s)", err)
	}

	return result["secondary"].(bool), err
}

// perform an rs.stepDown() on the connected instance
func (e *Env) mongoStepDown() error {
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
func (e *Env) getOplogLastEntries() bson.M {
	result := bson.M{}
	_       = e.mongo.DB("local").C("oplog.rs").Find(bson.M{}).Sort("-$natural").One(&result)

	return result
}

// get the first oplog entry
func (e *Env) getOplogFirstEntries() bson.M {
	result := bson.M{}
	_       = e.mongo.DB("local").C("oplog.rs").Find(bson.M{}).Sort("$natural").One(&result)

	return result
}

// get oplog entries that are greater than ts
func (e *Env) getOplogEntries(ts bson.MongoTimestamp) (iter *mgo.Iter) {
	query := bson.M{"ts": bson.M{"$gte": ts}}
	iter   = e.mongo.DB("local").C("oplog.rs").Find(query).Iter()
	return iter
}

// get the oplog number of entry
func (e *Env) getOplogCount() int {
	count, _ := e.mongo.DB("local").C("oplog.rs").Count()
	return count
}
