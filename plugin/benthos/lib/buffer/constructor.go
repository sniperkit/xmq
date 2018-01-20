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
	"bytes"
	"sort"
	"strings"

	"github.com/Jeffail/benthos/lib/buffer/impl"
	"github.com/Jeffail/benthos/lib/types"
	"github.com/Jeffail/benthos/lib/util/service/log"
	"github.com/Jeffail/benthos/lib/util/service/metrics"
)

//------------------------------------------------------------------------------

// typeSpec is a constructor and usage description for each buffer type.
type typeSpec struct {
	constructor func(conf Config, log log.Modular, stats metrics.Type) (Type, error)
	description string
}

var constructors = map[string]typeSpec{}

//------------------------------------------------------------------------------

// Config is the all encompassing configuration struct for all input types.
type Config struct {
	Type   string                `json:"type" yaml:"type"`
	Mmap   impl.MmapBufferConfig `json:"mmap_file" yaml:"mmap_file"`
	Memory impl.MemoryConfig     `json:"memory" yaml:"memory"`
}

// NewConfig returns a configuration struct fully populated with default values.
func NewConfig() Config {
	return Config{
		Type:   "none",
		Mmap:   impl.NewMmapBufferConfig(),
		Memory: impl.NewMemoryConfig(),
	}
}

//------------------------------------------------------------------------------

// Descriptions returns a formatted string of collated descriptions of each type.
func Descriptions() string {
	// Order our buffer types alphabetically
	names := []string{}
	for name := range constructors {
		names = append(names, name)
	}
	sort.Strings(names)

	buf := bytes.Buffer{}
	buf.WriteString("BUFFERS\n")
	buf.WriteString(strings.Repeat("=", 7))
	buf.WriteString("\n\n")
	buf.WriteString("This document has been generated with `benthos --list-buffers`.")
	buf.WriteString("\n\n")

	// Append each description
	for i, name := range names {
		buf.WriteString("## ")
		buf.WriteString("`" + name + "`")
		buf.WriteString("\n")
		buf.WriteString(constructors[name].description)
		if i != (len(names) - 1) {
			buf.WriteString("\n\n")
		}
	}
	return buf.String()
}

// New creates an input type based on an input configuration.
func New(conf Config, log log.Modular, stats metrics.Type) (Type, error) {
	if c, ok := constructors[conf.Type]; ok {
		return c.constructor(conf, log, stats)
	}
	return nil, types.ErrInvalidBufferType
}

//------------------------------------------------------------------------------
