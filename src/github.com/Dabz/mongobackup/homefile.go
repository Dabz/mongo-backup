/*
** log.go for log.go
**
** Made by gaspar_d
** Login   gaspar_d <d.gasparina@gmail.com>
**
** Started on  Fri 25 Dec 17:09:55 2015 gaspar_d
** Last update Mon 28 Dec 22:44:45 2015 gaspar_d
 */

package main

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"os"
	"time"
)

const (
	HomeFileVersion = "0.0.1"
)

type HomeLog struct {
	Version string        `json:"version"`
	Entries []BackupEntry `json:"entries"`
}

type BackupEntry struct {
	Ts        time.Time           `json:"ts"`
	Source    string              `json:"source"`
	Dest      string              `json:"dest"`
	Kind      string              `json:"kind"`
	Type      string              `json:"type"`
	Compress  bool                `json:"compress"`
	LastOplog bson.MongoTimestamp `json:"lastOplog"`
}

type HomeLogFile struct {
	content   HomeLog
	file      *os.File
	lastOplog bson.MongoTimestamp
}

func (b *HomeLogFile) Read(reader *os.File) error {
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

func (b *HomeLogFile) Create(writer *os.File) error {
	b.content.Version = HomeFileVersion
	b.content.Entries = []BackupEntry{}
	b.file = writer
	err := b.Flush()

	return err
}

func (b *HomeLogFile) AddNewEntry(in BackupEntry) error {
	b.content.Entries = append(b.content.Entries, in)
	b.Flush()
	return nil
}

func (b *HomeLogFile) Flush() error {
	buff, err := json.MarshalIndent(b.content, "", "  ")

	if err != nil {
		return err
	}

	b.file.Seek(0, 0)
	_, err = b.file.Write(buff)

	return err
}
