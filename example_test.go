package mjsonpatch

import "fmt"

var T = []byte(`{
	 "interests": [
		 "basketball"
	 ],
	 "A": {
		 "a3": [
			 {}
		 ]
	 }
 }`)

var T2 = []byte(`{
	"name": "string",
	"interests": [
		"basketball",
		"pingpong"
	],
	"A": {
		"a1": "string",
		"a2": "int",
		"a3": [
			{
				"b1": 1,
				"b2": "2"
			},
			{
			   "b1": 3,
			   "b2": "4"
		   }
		]
	},
	"B": "string"
}`)

var patch = []byte(`[
		{
			"op": "add",
			"path": "/interests/0",
			"value": "football"
		},
		{
			"op": "add",
			"path": "/B",
			"value": "xiongxiaoxiao"
		},
		{
			"op": "replace",
			"path": "/A/a1",
			"value": "a1-replace"
		},
		{
			"op": "remove",
			"path": "/A/a3/1",
			"value": ""
		}
	]`)

func ExampleMongoOP() {
	tpl := NewTemplate(T)
	ps, _ := Patchs(patch)
	m, _ := MongoOP(tpl, ps)
	fmt.Printf("%+v", m)
	// Output:
	// map[$pull:map[A.a3:<nil>] $push:map[interests:map[$each:[football] $position:0]] $set:map[A.a1:a1-replace B:xiongxiaoxiao] $unset:map[A.a3.1:]]
}
