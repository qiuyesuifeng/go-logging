// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const FORMAT_DATE_TIME string = "2006-01-02 15:04:05"

// RotateLogBackend utilizes the standard log module.
type RotateLogBackend struct {
	Logger *log.Logger

	DailyRolling bool
	HourRolling  bool

	FileName  string
	LogSuffix string

	fd   *os.File
	lock sync.Mutex
}

// NewRotateLogBackend creates a new RotateLogBackend.
func NewRotateLogBackend(fileName string, prefix string, flag int, rotate string) (*RotateLogBackend, error) {
	var dailyRolling bool
	var hourRolling bool

	var suffix string
	if rotate == "day" {
		dailyRolling = true
		suffix = genDayTime(time.Now())
	} else if rotate == "hour" {
		hourRolling = true
		suffix = genHourTime(time.Now())
	}

	backend := &RotateLogBackend{DailyRolling: dailyRolling, HourRolling: hourRolling, FileName: fileName, LogSuffix: suffix}

	fd, err := backend.createLogFile()
	if err != nil {
		return nil, err
	}

	backend.fd = fd
	backend.Logger = log.New(backend.fd, prefix, flag)
	return backend, nil
}

func (b *RotateLogBackend) Log(level Level, calldepth int, rec *Record) error {
	err := b.doCheckRotate()
	if err != nil {
		return err
	}

	return b.Logger.Output(calldepth+2, rec.Formatted(calldepth+1))
}

func (b *RotateLogBackend) createLogFile() (*os.File, error) {
	fd, err := os.OpenFile(b.FileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0664)
	return fd, err
}

func (b *RotateLogBackend) doCheckRotate() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	var suffix string
	if b.DailyRolling {
		suffix = genDayTime(time.Now())
	} else if b.HourRolling {
		suffix = genHourTime(time.Now())
	}

	// if suffix is not equal b.LogSuffix, then rotate
	if suffix != b.LogSuffix {
		err := b.doRotate(suffix)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *RotateLogBackend) doRotate(suffix string) error {
	// convert xxx.log to xxx.log.20150129
	oldFileName := b.FileName + "." + b.LogSuffix
	err := os.Rename(b.FileName, oldFileName)
	if err != nil {
		return fmt.Errorf("[doRotate][Rename][Error]%s\n", err)
	}

	fd, err := b.createLogFile()
	if err != nil {
		return fmt.Errorf("[doRotate][createLogFile][Error]%s\n", err)
	}

	// Is this ok?
	oldFd := b.fd
	defer oldFd.Close()

	b.fd = fd
	b.LogSuffix = suffix

	prefix := b.Logger.Prefix()
	flag := b.Logger.Flags()
	b.Logger = log.New(b.fd, prefix, flag)
	return nil
}

func genDayTime(t time.Time) string {
	now := t.Format(FORMAT_DATE_TIME)
	year := now[0:4]
	month := now[5:7]
	day := now[8:10]

	ret := year + month + day
	return ret
}

func genHourTime(t time.Time) string {
	now := t.Format(FORMAT_DATE_TIME)
	year := now[0:4]
	month := now[5:7]
	day := now[8:10]
	hour := now[11:13]

	ret := year + month + day + hour
	return ret
}
