package main

import (
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"github.com/snipem/gt7fuel/lib"
	"github.com/snipem/gt7tools/lib/dump"
	"testing"
)

func Benchmark_Run(b *testing.B) {

	//for i := 0; i < b.N; i++ {

	gt7c := gt7.NewGT7Communication("255.255.255.255")

	dumpFilePath := "../gt7testdata/watkinsglen.gob.gz"

	gt7dump, err := dump.NewGT7Dump(dumpFilePath, gt7c)
	if err != nil {
		panic(err)
	}
	gt7dump.DataSendFrequency = 0 // full throttle data sending
	WaitTime = 0
	gt7stats = lib.NewStats()

	go gt7dump.Run()

	raceTimeInMinutes := 25
	go LogRace(gt7c, gt7stats, &raceTimeInMinutes)

	loggedMessages := 0
	maxMessages := 10000

	for loggedMessages <= maxMessages {
		gt7stats.GetRealTimeMessage()
		loggedMessages++
	}

	gt7stats.ShallRun = false
	//}

}
