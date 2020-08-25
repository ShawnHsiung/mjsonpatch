package mjsonpatch

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

const (
	Dict = iota
	Array
)

type Template map[string]interface{}
type Object = map[string]interface{}

func NewTemplate(tpl []byte) *Template {
	var o Template
	err := json.Unmarshal(tpl, &o)
	if err != nil {
		panic(err)
	}
	return &o
}

/*
[
  { "op": "replace", "path": "/baz", "value": "boo" },
  { "op": "add", "path": "/hello", "value": ["world"] },
  { "op": "remove", "path": "/foo" }
]
*/
func (o *Template) MongoOP(op *Operation, result *Object) error {

	if op == nil {
		return errors.New("op is nil")
	}
	if result == nil {
		return errors.New("result is nil")
	}
	action, err := op.action()
	if err != nil {
		return err
	}
	path, err := op.path()
	if err != nil {
		return err
	}
	path = trimPath(path)

	switch action {
	case REMOVE:
		return o.remove(result, path)
	case ADD:
		return o.add(result, path, op)
	case REPLACE:
		return o.replace(result, path, op)
	default:
		return errors.New("unsupport op: " + action)
	}
}

func (o *Template) remove(out *Object, path string) error {

	unset, ok := (*out)["$unset"]
	if !ok {
		unset = Object{}
	}

	remove, _ := unset.(Object)
	remove[path] = ""

	parts := strings.Split(path, ".")
	parent := strings.Join(parts[:len(parts)-1], ".")
	t := o.kind(parent)
	if t == Array {
		pull, ok := (*out)["$pull"]
		if !ok {
			pull = Object{}
		}
		pullArr, _ := pull.(Object)
		pullArr[parent] = nil
		(*out)["$pull"] = pullArr
	}

	(*out)["$unset"] = remove
	return nil
}

func (o *Template) add(out *Object, path string, value *Operation) error {

	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return o.replace(out, path, value)
	}

	lastE := parts[len(parts)-1]
	parent := strings.Join(parts[:len(parts)-1], ".")
	parentKind := o.kind(parent)
	if parentKind == Dict && o.kind(path) == Dict {
		return o.replace(out, path, value)
	}

	v, err := value.valueInterface()
	if err != nil {
		return err
	}

	pos := -1
	if lastE != "-" {
		pos, err = strconv.Atoi(lastE)
		if err != nil {
			parent = path // add field
			pos = -1
		}
	}

	push, ok := (*out)["$push"]
	if !ok {
		push = Object{}
	}
	push2, _ := push.(Object)

	switch reflect.TypeOf(v).Kind() {
	case reflect.Slice, reflect.Array:
		if pos == -1 {
			push2[parent] = Object{
				"$each": v,
			}
		} else {
			push2[parent] = Object{
				"$each":     v,
				"$position": pos,
			}
		}
	default:
		if pos == -1 {
			push2[parent] = v
		} else {
			push2[parent] = Object{
				"$each":     []interface{}{v},
				"$position": pos,
			}
		}
	}

	(*out)["$push"] = push2
	return nil
}

func (o *Template) replace(out *Object, path string, value *Operation) error {
	v, err := value.valueInterface()
	if err != nil {
		return err
	}

	set, ok := (*out)["$set"]
	if !ok {
		set1 := Object{
			path: v,
		}
		(*out)["$set"] = set1
		return nil
	}
	set2, _ := set.(Object)
	set2[path] = v
	(*out)["$set"] = set2
	return nil
}

func (o *Template) kind(path string) int {
	parts := strings.Split(path, ".")
	var obj interface{}
	obj = Object(*o)

	l := len(parts)
	for i := 0; i <= l; i++ {
		switch v := obj.(type) {
		case map[string]interface{}:
			if i == l {
				return Dict
			}
			if _, ok := v[parts[i]]; ok {
				obj = v[parts[i]]
			} else { // WARN
				obj = Object{}
			}
		case []interface{}:
			if i == l {
				return Array
			}
			if len(v) == 0 {
				panic("no element in array of path: " + path)
			}
			// base on template
			obj = v[0]
		}
	}
	return Dict
}

func trimPath(path string) string {
	path = strings.TrimPrefix(path, "/")
	path = strings.ReplaceAll(path, "/", ".")
	path = strings.ReplaceAll(path, "~1", "/")
	path = strings.ReplaceAll(path, "~0", "~")
	return path
}
