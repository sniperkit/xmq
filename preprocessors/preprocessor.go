package preprocessors

import (
	"fmt"
	"strings"

	"github.com/corpix/loggers"
	"github.com/corpix/loggers/logger/prefixwrapper"

	"github.com/cryptounicorns/gluttony/preprocessors/preprocessor/lua"
	"github.com/cryptounicorns/gluttony/preprocessors/preprocessor/none"
)

type Preprocessor interface {
	Preprocess(interface{}) (interface{}, error)
	Close() error
}

func New(c Config, l loggers.Logger) (Preprocessor, error) {
	var (
		t   = strings.ToLower(c.Type)
		log = prefixwrapper.New(
			fmt.Sprintf("Preprocessor(%s): ", t),
			l,
		)
	)

	switch t {
	case lua.Name:
		return lua.New(c.Lua, log)
	case none.Name:
		return none.New(c.None, log)
	default:
		return nil, NewErrUnknownPreprocessorType(c.Type)
	}
}
