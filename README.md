# Mongo-backup
***Incremental backup tool for MongoDB***

## Overview
Mongobackup is an external tool performing full & incremental backup. Backup are stored on the filesystem and compress using the [lz4 algorithm](https://code.google.com/p/lz4/). Full backup are done by performing a complete file system copy of the dbPath and partial oplog dump is used for incremental backup.

## Features
  * Full & incremental backup
  * Point in Time restore for incremental & full backup 
  * Backup separation by _kind_ (e.g. daily, monthly, weekly, ...)
  * Partial dump of the oplog. Only the operations after the last known one are dumped
  * Secure backup by using [fsyncLock & fsyncUnlock] (https://docs.mongodb.org/manual/reference/method/db.fsyncLock/)
  * Avoid interacting with the primary node

## Usages
Perform an incremental backup
```
./bin/mongobackup backup [--kind string] [--nocompress] [--nofsynclock] [--nostepdown]
```
Perform a full backup         
```
./bin/mongobackup backup --full [--kind string] [--nocompress] [--nofsynclock] [--nostepdown]
```
Restore a specific backup
```
./bin/mongobackup restore --out string --snapshot string
```
Perform a point in time restore
```
./bin/mongobackup restore --out string --pit string
```
Delete a range of backup
```
./bin/mongobackup delete --kind string --entries string
```
Delete a specific backup
```
./bin/mongobackup delete --snapshot string
```
List available backups
```
./bin/mongobackup list [--kind string] [--entries string]
```

## Sample configuration

Scheduling has to be performed using an external tool, e.g. cron
Bellow a sample configuration for a daily backup where a full backup is performed once a week every Sunday and where we stored a daily backup for the last 7 days and a monthly backups for the last 13 months.
```cron
0 0 * * Sun     mongobackup --basedir /backup --full --kind daily   && mongobackup delete --basedir /backup --entries '7-'
0 0 * * Mon-Sat mongobackup --basedir /backup --kind daily
0 0 1 * *       mongobackup --basedir /backup --full --kind monthly && mongobackup delete --basedir /backup --entries '13-'
```

## Releases

The project is in version **0.01** and is currently **not yet ready for production**.
If you are interested and want more information or want to participate, feel free to ping me ;)
