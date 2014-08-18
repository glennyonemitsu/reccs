package main

type CommandParameter struct {
	Name     string
	Type     string // "string", "integer", "collection" (which is a string), "binary"
	Required bool
}
