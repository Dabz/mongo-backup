/*
** log.go for log.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Fri 25 Dec 17:09:55 2015 gaspar_d
** Last update Mon  7 Mar 16:53:41 2016 gaspar_d
*/

package mongobackup

import (
  "encoding/json"
  "gopkg.in/mgo.v2/bson"
  "os"
  "time"
  "errors"
  "strconv"
)

const (
  HomeFileVersion = "0.0.1"
  SuffixInc       = '+'
  SuffixDec       = '-'
)

// Home file json representation
type HomeLog struct {
  Version  string        `json:"version"`
  Entries  []BackupEntry `json:"entries"`
  Sequence int           `json:"seq"`
}

// Home file backup entry json representation
type BackupEntry struct {
  Id         string              `json:"id"`
  Ts         time.Time           `json:"ts"`
  Source     string              `json:"source"`
  Dest       string              `json:"dest"`
  Kind       string              `json:"kind"`
  Type       string              `json:"type"`
  Compress   bool                `json:"compress"`
  FirstOplog bson.MongoTimestamp `json:"firstOplog"`
  LastOplog  bson.MongoTimestamp `json:"lastOplog"`
}

// Represent a home file
type BackupHistoryFile struct {
  content   HomeLog
  file      *os.File
  lastOplog bson.MongoTimestamp
}

// Read & Populate the homefile structure from a file
func (b *BackupHistoryFile) Read(reader *os.File) error {
  result := HomeLog{}
  b.file = reader
  dec := json.NewDecoder(reader)
  err := dec.Decode(&result)
  b.content = result

  if err != nil {
    return err
  }

  for _, obj := range b.content.Entries {
    if b.lastOplog == 0 {
      b.lastOplog = obj.LastOplog
    } else if b.lastOplog < obj.LastOplog {
      b.lastOplog = obj.LastOplog
    }
  }

  return nil
}

// create a new homelogfile and write it to the disk
func (b *BackupHistoryFile) Create(writer *os.File) error {
  b.content.Version  = HomeFileVersion
  b.content.Entries  = []BackupEntry{}
  b.content.Sequence = 0
  b.file = writer
  return nil
}

// add a new entry and flush it to disk
func (b *BackupHistoryFile) AddNewEntry(in BackupEntry) error {
  b.content.Entries   = append(b.content.Entries, in)
  b.content.Sequence += 1
  return nil
}

// remove an entry and flush it to disk
func (b *BackupHistoryFile) RemoveEntry(rm BackupEntry) error {
  entries := []BackupEntry{}

  for _, entry := range b.content.Entries {
    if entry.Id == rm.Id {
      continue
    }

    entries = append(entries, entry)
  }

  b.content.Entries = entries

  return nil
}

// flush the homelogfile to disk
func (b *BackupHistoryFile) Flush() error {
  buff, err := json.MarshalIndent(b.content, "", "  ")

  if err != nil {
    return err
  }

  b.file.Truncate(0)
  b.file.Seek(0, 0)
  _, err = b.file.Write(buff)

  return err
}

// return a backup associated to this speicifc id
func (b *BackupHistoryFile) GetBackupEntry(id string) *BackupEntry {
  for _, entry :=  range b.content.Entries {
    if entry.Id == id {
      return &entry
    }
  }

  return nil
}

// return the last full backup realized before a specific entry
func (b *BackupHistoryFile) GetLastFullBackup(etr BackupEntry) *BackupEntry {
  for _, entry :=  range b.content.Entries {
    if entry.Ts.Before(etr.Ts) && entry.Type == "full" && entry.Kind == etr.Kind {
      return &entry
    }
  }

  return nil
}

// return the next entry after the one provided
func (b *BackupHistoryFile) GetNextBackup(etr BackupEntry) BackupEntry {
  lastentry := BackupEntry{}

  for _, entry := range b.content.Entries {
    if lastentry.Id == etr.Id {
      return entry
    }
    lastentry = entry
  }

  return BackupEntry{}
}

// get the last entry before the requested date
// used to determine which snapshots to recover for pit
func (b *BackupHistoryFile) GetLastEntryAfter(ts time.Time) *BackupEntry {
  lastentry := BackupEntry{}
  for _, entry := range b.content.Entries {
    if entry.Ts.After(ts) {
      if lastentry.Id == "" {
        return nil
      }
      return &lastentry
    }

    lastentry = entry
  }

  return nil
}

// check that that incremental restoration of this backup will be consistent
func (b *BackupHistoryFile) CheckIncrementalConsistency(entry *BackupEntry) (error) {
  fullEntry := b.GetLastFullBackup(*entry)
  entries   := b.GetIncEntriesBetween(fullEntry, entry)
  lastval   := *fullEntry

  for _, e := range entries {
    if lastval.LastOplog <= e.FirstOplog {
      lastval = e
    } else {
      return errors.New("gap detected between " + lastval.Id + " and " + e.Id)
    }
  }

  return nil
}

// get all incremental BackupEntry between two specific entry
// used to realize point in time recovery and recreate the oplog
func (b *BackupHistoryFile) GetIncEntriesBetween(from, to *BackupEntry) []BackupEntry {
  results := []BackupEntry{}
  for _, entry :=  range b.content.Entries {
    if entry.Ts.After(from.Ts) && entry.Ts.Before(to.Ts) && entry.Kind == from.Kind {
      if (entry.Type == "inc") {
        results = append(results, entry)
      }
    }
  }

  if to.Kind == from.Kind && to.Type == "inc" {
    results = append(results, *to)
  }

  return results
}

// Find entries according to a criteria
func (b* BackupHistoryFile) FindEntriesFromCriteria(criteria string, input []BackupEntry) (error, []BackupEntry) {
  var (
    position int
    suffix   uint8
    err      error
    result   []BackupEntry
  )

  suffix       = 0
  criterialen := len(criteria)
  lastchar    := criteria[criterialen - 1]

  if lastchar == SuffixInc || lastchar == SuffixDec {
    suffix   = lastchar
    criteria = criteria[:criterialen - 1]
  }

  position, err = strconv.Atoi(criteria)
  if err != nil {
    return err, result
  }

  ilist := []BackupEntry{}
  if suffix == SuffixInc {
    ilist = input
  } else if suffix == SuffixDec {
    ilist = input
    for i, j := 0, len(ilist)-1; i < j; i, j = i+1, j-1 {
      ilist[i], ilist[j] = ilist[j], ilist[i]
    }
  }

  for i, entry := range ilist {
    if suffix == 0 && i == position {
      result = append(result, entry)
    } else if i >= position {
      result = append(result, entry)
    }
  }
  if suffix == SuffixDec {
    for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
      result[i], result[j] = result[j], result[i]
    }
  }

  return nil, result
}

// return all entries from this kind
func (b* BackupHistoryFile) FindEntriesFromKind(kind string, input []BackupEntry) (error, []BackupEntry) {
  result := []BackupEntry{}

  for _, entry := range input {
    if entry.Kind == kind {
      result = append(result, entry)
    }
  }

  return nil, result
}

// return entries according to a criteria (string0
// TODO should we use a lexer/parser?
func (b *BackupHistoryFile) FindEntries(criteria, kind string) (error, []BackupEntry) {
  var (
    result      []BackupEntry
    err         error
  )

  // filter on kind
  if kind != "" {
    err, result = b.FindEntriesFromKind(kind, b.content.Entries)
    if err != nil {
      return err, result
    }
  } else { // no kind
    result = b.content.Entries
  }

  // filter on criteria
  if criteria != "" {
    err, result = b.FindEntriesFromCriteria(criteria, result)
    if err != nil {
      return err, result
    }
  }

  return nil, result
}
