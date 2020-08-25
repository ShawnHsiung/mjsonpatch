package mjsonpatch

import (
	"bytes"
	"encoding/json"
	"errors"
	"sync"
)

const (
	REMOVE  = "remove"
	ADD     = "add"
	REPLACE = "replace"
	MOVE    = "move"
	COPY    = "copy"
	TEST    = "test"
)

var (
	valPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

type Operation struct {
	OP    string          `json:"op,omitempty"`
	Path  string          `json:"path,omitempty"`
	Value json.RawMessage `json:"value,omitempty"`
}

type Patch []*Operation

func (op *Operation) path() (string, error) {
	if op.Path == "" {
		return "", errors.New("path is empty")
	}
	return op.Path, nil
}

func (op *Operation) action() (string, error) {
	if op.OP == "" {
		return "", errors.New("op is empty")
	}
	return op.OP, nil
}

func (o *Operation) valueInterface() (interface{}, error) {
	if len(o.Value) > 0 {

		buf := valPool.Get().(*bytes.Buffer)
		buf.Write(o.Value)
		defer func() {
			buf.Reset()
			valPool.Put(buf)
		}()
		dec := json.NewDecoder(buf)
		dec.UseNumber()

		var v interface{}
		err := dec.Decode(&v)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, errors.New("missing value field")
}
