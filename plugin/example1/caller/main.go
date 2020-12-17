package main

import (
	"plugin"
)

func main() {
	plugin, err := plugin.Open("../plugin/test.so")
	if err != nil {
		panic(err)
	}

	foo, _ := plugin.Lookup("Foo")
	foo.(func())()
}
