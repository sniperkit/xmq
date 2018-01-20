// Copyright (c) 2017 Ashley Jeffs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package processor

import (
	"os"
	"testing"

	"github.com/Jeffail/benthos/lib/types"
	"github.com/Jeffail/benthos/lib/util/service/log"
	"github.com/Jeffail/benthos/lib/util/service/metrics"
)

func TestSample10Percent(t *testing.T) {
	conf := NewConfig()
	conf.Sample.Retain = 0.1

	testLog := log.NewLogger(os.Stdout, log.LoggerConfig{LogLevel: "NONE"})
	proc, err := NewSample(conf, testLog, metrics.DudType{})
	if err != nil {
		t.Error(err)
		return
	}

	total := 100000
	totalSampled := 0
	margin := 0.01
	for i := 0; i < total; i++ {
		msgIn := types.NewMessage()
		msgOut, _, propagate := proc.ProcessMessage(&msgIn)
		if propagate {
			if &msgIn != msgOut {
				t.Error("Message told to propagate but not given")
			}
			totalSampled++
		}
	}

	act, exp := float64(totalSampled)/float64(total), conf.Sample.Retain
	var sampleError float64
	if exp > act {
		sampleError = (exp - act) / exp
	} else {
		sampleError = (act - exp) / exp
	}
	if sampleError > margin {
		t.Errorf("Sample error greater than margin: %v != %v", act, exp)
	}
}

func TestSample24Percent(t *testing.T) {
	conf := NewConfig()
	conf.Sample.Retain = 0.24

	testLog := log.NewLogger(os.Stdout, log.LoggerConfig{LogLevel: "NONE"})
	proc, err := NewSample(conf, testLog, metrics.DudType{})
	if err != nil {
		t.Error(err)
		return
	}

	total := 100000
	totalSampled := 0
	margin := 0.01
	for i := 0; i < total; i++ {
		msgIn := types.NewMessage()
		msgOut, _, propagate := proc.ProcessMessage(&msgIn)
		if propagate {
			if &msgIn != msgOut {
				t.Error("Message told to propagate but not given")
			}
			totalSampled++
		}
	}

	act, exp := float64(totalSampled)/float64(total), conf.Sample.Retain
	var sampleError float64
	if exp > act {
		sampleError = (exp - act) / exp
	} else {
		sampleError = (act - exp) / exp
	}
	if sampleError > margin {
		t.Errorf("Sample error greater than margin: %v != %v", act, exp)
	}
}
