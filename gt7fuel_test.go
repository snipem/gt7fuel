package main

import (
	"testing"
)

func Benchmark_Run(b *testing.B) {

	WaitTime = 0
	for i := 0; i < b.N; i++ {
		run(raceTimeInMinutes, false, "", "../gt7testdata/watkinsglen.gob.gz")
	}

}
