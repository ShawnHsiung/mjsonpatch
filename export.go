package mjsonpatch

import (
	"encoding/json"
	"errors"
)

func Patchs(data []byte) (*Patch, error) {

	if len(data) == 0 {
		return nil, errors.New("no patchs")
	}

	var p Patch
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func MongoOP(t *Template, patchs *Patch) (map[string]interface{}, error) {
	result := Object{}
	for _, op := range *patchs {
		t.MongoOP(op, &result)
	}
	return result, nil
}
