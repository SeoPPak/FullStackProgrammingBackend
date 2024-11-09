package main

import (
	"oauth/server"
)

func main() {
	r := server.Setup()
	r.Run(":5000")
}
