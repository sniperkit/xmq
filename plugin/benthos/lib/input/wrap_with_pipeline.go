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

package input

import (
	"time"

	"github.com/Jeffail/benthos/lib/pipeline"
	"github.com/Jeffail/benthos/lib/types"
	"github.com/Jeffail/benthos/lib/util"
)

//------------------------------------------------------------------------------

// WithPipeline is a type that wraps both an input type and a pipeline type
// by routing the input through the pipeline, and implements the input.Type
// interface in order to act like an ordinary input.
type WithPipeline struct {
	in   Type
	pipe pipeline.Type
}

// WrapWithPipeline routes an input directly into a processing pipeline and
// returns a type that manages both and acts like an ordinary input.
func WrapWithPipeline(in Type, pipeConstructor pipeline.ConstructorFunc) (*WithPipeline, error) {
	pipe, err := pipeConstructor()
	if err != nil {
		return nil, err
	}
	if err = util.Couple(in, pipe); err != nil {
		return nil, err
	}
	return &WithPipeline{
		in:   in,
		pipe: pipe,
	}, nil
}

// WrapWithPipelines wraps an input with a variadic number of pipelines.
func WrapWithPipelines(in Type, pipeConstructors ...pipeline.ConstructorFunc) (*WithPipeline, error) {
	var err error
	var pipe *WithPipeline
	for i, ctor := range pipeConstructors {
		if i == 0 {
			if pipe, err = WrapWithPipeline(in, ctor); err != nil {
				return nil, err
			}
		} else if pipe, err = WrapWithPipeline(pipe, ctor); err != nil {
			return nil, err
		}
	}
	return pipe, nil
}

//------------------------------------------------------------------------------

// MessageChan returns the channel used for consuming messages from this input.
func (i *WithPipeline) MessageChan() <-chan types.Message {
	return i.pipe.MessageChan()
}

// StartListening starts the type listening to a response channel from a
// consumer.
func (i *WithPipeline) StartListening(resChan <-chan types.Response) error {
	return i.pipe.StartListening(resChan)
}

//------------------------------------------------------------------------------

// CloseAsync triggers a closure of this object but does not block.
func (i *WithPipeline) CloseAsync() {
	i.in.CloseAsync()
}

// WaitForClose is a blocking call to wait until the object has finished closing
// down and cleaning up resources.
func (i *WithPipeline) WaitForClose(timeout time.Duration) error {
	return i.pipe.WaitForClose(timeout)
}

//------------------------------------------------------------------------------
