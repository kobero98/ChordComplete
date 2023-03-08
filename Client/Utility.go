package main

type Node struct {
	Name       string
	Ip         []string
	Port       int
	PortExtern int
	Index      int
}
type ReplyRegistration struct {
	Precedente Node
	Successivo Node
	NumNod     int
}
