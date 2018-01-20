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

package buffer

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Jeffail/benthos/lib/types"
)

//------------------------------------------------------------------------------

func TestNoneBufferClose(t *testing.T) {
	empty, err := NewEmpty(NewConfig(), nil, nil)
	if err != nil {
		t.Error(err)
		return
	}

	msgChan := make(chan types.Message)
	resChan := make(chan types.Response)

	if err = empty.StartReceiving(msgChan); err != nil {
		t.Error(err)
		return
	}
	if err = empty.StartReceiving(msgChan); err == nil {
		t.Error("received nil, expected error from double msg assignment")
		return
	}
	if err = empty.StartListening(resChan); err != nil {
		t.Error(err)
		return
	}
	if err = empty.StartListening(resChan); err == nil {
		t.Error("received nil, expected error from double res assignment")
		return
	}

	empty.CloseAsync()
	if err = empty.WaitForClose(time.Second); err != nil {
		t.Error(err)
	}
}

func TestNoneBufferBasic(t *testing.T) {
	nThreads, nMessages := 5, 100

	empty, err := NewEmpty(NewConfig(), nil, nil)
	if err != nil {
		t.Error(err)
		return
	}

	msgChan := make(chan types.Message)
	resChan := make(chan types.Response)

	if err = empty.StartListening(resChan); err != nil {
		t.Error(err)
		return
	}
	if err = empty.StartReceiving(msgChan); err != nil {
		t.Error(err)
		return
	}

	go func() {
		for msg := range empty.MessageChan() {
			resChan <- types.NewSimpleResponse(errors.New(string(msg.Parts[0])))
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(nThreads)

	for i := 0; i < nThreads; i++ {
		go func(nThread int) {
			for j := 0; j < nMessages; j++ {
				msg := fmt.Sprintf("Hello World %v %v", nThread, j)
				msgChan <- types.Message{
					Parts: [][]byte{[]byte(msg)},
				}
				select {
				case res := <-empty.ResponseChan():
					if actual := res.Error().Error(); msg != actual {
						t.Errorf("Wrong result: %v != %v", msg, actual)
					}
				case <-time.After(time.Second):
					t.Error("Timed out waiting for response")
				}
			}
			wg.Done()
		}(i)
	}

	wg.Wait()

	close(msgChan)
	if err = empty.WaitForClose(time.Second); err != nil {
		t.Error(err)
	}
}

//------------------------------------------------------------------------------
