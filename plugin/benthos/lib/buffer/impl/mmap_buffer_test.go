// Copyright (c) 2014 Ashley Jeffs
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

package impl

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/Jeffail/benthos/lib/types"
	"github.com/Jeffail/benthos/lib/util/service/log"
	"github.com/Jeffail/benthos/lib/util/service/metrics"
)

var logConfig = log.LoggerConfig{
	LogLevel: "NONE",
}

func cleanUpMmapDir(dir string) {
	os.RemoveAll(dir)
}

func TestMmapBufferInterface(t *testing.T) {
	b := &MmapBuffer{}
	if c := Buffer(b); c == nil {
		t.Error("MmapBuffer does not satisfy the Buffer interface")
	}
}

func TestMmapBufferBasic(t *testing.T) {
	dir, err := ioutil.TempDir("", "benthos_test_")
	if err != nil {
		t.Error(err)
		return
	}

	defer cleanUpMmapDir(dir)

	n := 100

	conf := NewMmapBufferConfig()
	conf.FileSize = 100000
	conf.Path = dir

	block, err := NewMmapBuffer(conf, log.NewLogger(os.Stdout, logConfig), metrics.DudType{})
	if err != nil {
		t.Error(err)
		return
	}
	defer block.Close()

	for i := 0; i < n; i++ {
		if _, err := block.PushMessage(types.Message{
			Parts: [][]byte{
				[]byte("hello"),
				[]byte("world"),
				[]byte("12345"),
				[]byte(fmt.Sprintf("test%v", i)),
			},
		}); err != nil {
			t.Error(err)
			return
		}
	}

	for i := 0; i < n; i++ {
		m, err := block.NextMessage()
		if err != nil {
			t.Error(err)
			return
		}
		if len(m.Parts) != 4 {
			t.Errorf("Wrong # parts, %v != %v", len(m.Parts), 4)
		} else if expected, actual := fmt.Sprintf("test%v", i), string(m.Parts[3]); expected != actual {
			t.Errorf("Wrong order of messages, %v != %v", expected, actual)
		}
		if _, err := block.ShiftMessage(); err != nil {
			t.Error(err)
			return
		}
	}
}

func TestMmapBufferBacklogCounter(t *testing.T) {
	dir, err := ioutil.TempDir("", "benthos_test_")
	if err != nil {
		t.Error(err)
		return
	}

	defer cleanUpMmapDir(dir)

	conf := NewMmapBufferConfig()
	conf.FileSize = 100000
	conf.Path = dir

	block, err := NewMmapBuffer(conf, log.NewLogger(os.Stdout, logConfig), metrics.DudType{})
	if err != nil {
		t.Error(err)
		return
	}
	defer block.Close()

	if _, err := block.PushMessage(types.Message{
		Parts: [][]byte{[]byte("1234")}, // 4 bytes + 4 bytes
	}); err != nil {
		t.Error(err)
		return
	}

	if expected, actual := 16, block.backlog(); expected != actual {
		t.Errorf("Wrong backlog count: %v != %v", expected, actual)
	}

	if _, err := block.PushMessage(types.Message{
		Parts: [][]byte{
			[]byte("1234"),
			[]byte("1234"),
		}, // ( 4 bytes + 4 bytes ) * 2
	}); err != nil {
		t.Error(err)
		return
	}

	if expected, actual := 40, block.backlog(); expected != actual {
		t.Errorf("Wrong backlog count: %v != %v", expected, actual)
	}

	if _, err := block.ShiftMessage(); err != nil {
		t.Error(err)
		return
	}

	if expected, actual := 24, block.backlog(); expected != actual {
		t.Errorf("Wrong backlog count: %v != %v", expected, actual)
	}

	if _, err := block.ShiftMessage(); err != nil {
		t.Error(err)
		return
	}

	if expected, actual := 0, block.backlog(); expected != actual {
		t.Errorf("Wrong backlog count: %v != %v", expected, actual)
	}
}

func TestMmapBufferLoopingRandom(t *testing.T) {
	dir, err := ioutil.TempDir("", "benthos_test_")
	if err != nil {
		t.Error(err)
		return
	}

	defer cleanUpMmapDir(dir)

	conf := NewMmapBufferConfig()
	conf.FileSize = 8000
	conf.Path = dir

	block, err := NewMmapBuffer(conf, log.NewLogger(os.Stdout, logConfig), metrics.DudType{})
	if err != nil {
		t.Error(err)
		return
	}
	defer block.Close()

	n, iter := 50, 5

	for j := 0; j < iter; j++ {
		for i := 0; i < n; i++ {
			b := make([]byte, rand.Int()%100)
			for k := range b {
				b[k] = '0'
			}
			if _, err := block.PushMessage(types.Message{
				Parts: [][]byte{
					b,
					[]byte(fmt.Sprintf("test%v", i)),
				},
			}); err != nil {
				t.Error(err)
				return
			}
		}

		for i := 0; i < n; i++ {
			m, err := block.NextMessage()
			if err != nil {
				t.Error(err)
				return
			}
			if len(m.Parts) != 2 {
				t.Errorf("Wrong # parts, %v != %v", len(m.Parts), 4)
			} else if expected, actual := fmt.Sprintf("test%v", i), string(m.Parts[1]); expected != actual {
				t.Errorf("Wrong order of messages, %v != %v", expected, actual)
			}
			if _, err := block.ShiftMessage(); err != nil {
				t.Error(err)
				return
			}
		}
	}
}

func TestMmapBufferMultiFiles(t *testing.T) {
	dir, err := ioutil.TempDir("", "benthos_test_")
	if err != nil {
		t.Error(err)
		return
	}

	defer cleanUpMmapDir(dir)

	n := 10000

	conf := NewMmapBufferConfig()
	conf.FileSize = 1000
	conf.Path = dir

	block, err := NewMmapBuffer(conf, log.NewLogger(os.Stdout, logConfig), metrics.DudType{})
	if err != nil {
		t.Error(err)
		return
	}
	defer block.Close()

	for i := 0; i < n; i++ {
		if _, err := block.PushMessage(types.Message{
			Parts: [][]byte{
				[]byte("hello"),
				[]byte("world"),
				[]byte("12345"),
				[]byte(fmt.Sprintf("test%v", i)),
			},
		}); err != nil {
			t.Error(err)
			return
		}
	}

	for i := 0; i < n; i++ {
		m, err := block.NextMessage()
		if err != nil {
			t.Error(err)
			return
		}
		if len(m.Parts) != 4 {
			t.Errorf("Wrong # parts, %v != %v", len(m.Parts), 4)
		} else if expected, actual := fmt.Sprintf("test%v", i), string(m.Parts[3]); expected != actual {
			t.Errorf("Wrong order of messages, %v != %v", expected, actual)
		}
		if _, err := block.ShiftMessage(); err != nil {
			t.Error(err)
			return
		}
	}
}

func TestMmapBufferRecoverFiles(t *testing.T) {
	dir, err := ioutil.TempDir("", "benthos_test_")
	if err != nil {
		t.Error(err)
		return
	}

	defer cleanUpMmapDir(dir)

	n := 10000

	conf := NewMmapBufferConfig()
	conf.FileSize = 1000
	conf.Path = dir

	// Write a load of data
	block, err := NewMmapBuffer(conf, log.NewLogger(os.Stdout, logConfig), metrics.DudType{})
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < n; i++ {
		if _, err := block.PushMessage(types.Message{
			Parts: [][]byte{
				[]byte("hello"),
				[]byte("world"),
				[]byte("12345"),
				[]byte(fmt.Sprintf("test%v", i)),
			},
		}); err != nil {
			t.Error(err)
			return
		}
	}

	// Close down any handlers we have.
	block.Close()

	// Read the data back
	block, err = NewMmapBuffer(conf, log.NewLogger(os.Stdout, logConfig), metrics.DudType{})
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < n; i++ {
		m, err := block.NextMessage()
		if err != nil {
			t.Error(err)
			return
		}
		if len(m.Parts) != 4 {
			t.Errorf("Wrong # parts, %v != %v", len(m.Parts), 4)
		} else if expected, actual := fmt.Sprintf("test%v", i), string(m.Parts[3]); expected != actual {
			t.Errorf("Wrong order of messages, %v != %v", expected, actual)
		}
		if _, err := block.ShiftMessage(); err != nil {
			t.Error(err)
			return
		}
	}

	block.Close()
}

func TestMmapBufferRejectLargeMessage(t *testing.T) {
	dir, err := ioutil.TempDir("", "benthos_test_")
	if err != nil {
		t.Error(err)
		return
	}

	defer cleanUpMmapDir(dir)

	tMsg := types.Message{Parts: make([][]byte, 1)}
	tMsg.Parts[0] = []byte("hello world this message is too long!")

	conf := NewMmapBufferConfig()
	conf.FileSize = 10
	conf.Path = dir

	// Write a load of data
	block, err := NewMmapBuffer(conf, log.NewLogger(os.Stdout, logConfig), metrics.DudType{})
	if err != nil {
		t.Error(err)
		return
	}
	_, err = block.PushMessage(tMsg)
	if exp, actual := types.ErrMessageTooLarge, err; exp != actual {
		t.Errorf("Unexpected error: %v != %v", exp, actual)
	}
}

func BenchmarkMmapBufferBasic(b *testing.B) {
	dir, err := ioutil.TempDir("", "benthos_test_")
	if err != nil {
		b.Error(err)
		return
	}

	defer cleanUpMmapDir(dir)

	conf := NewMmapBufferConfig()
	conf.FileSize = 1000
	conf.Path = dir

	block, err := NewMmapBuffer(conf, log.NewLogger(os.Stdout, logConfig), metrics.DudType{})
	if err != nil {
		b.Error(err)
		return
	}
	defer block.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := block.PushMessage(types.Message{
			Parts: [][]byte{
				[]byte("hello"),
				[]byte("world"),
				[]byte("12345"),
				[]byte(fmt.Sprintf("test%v", i)),
			},
		}); err != nil {
			b.Error(err)
			return
		}
	}

	for i := 0; i < b.N; i++ {
		m, err := block.NextMessage()
		if err != nil {
			b.Error(err)
			return
		}
		if len(m.Parts) != 4 {
			b.Errorf("Wrong # parts, %v != %v", len(m.Parts), 4)
		} else if expected, actual := fmt.Sprintf("test%v", i), string(m.Parts[3]); expected != actual {
			b.Errorf("Wrong order of messages, %v != %v", expected, actual)
		}
		if _, err := block.ShiftMessage(); err != nil {
			b.Error(err)
			return
		}
	}
}
