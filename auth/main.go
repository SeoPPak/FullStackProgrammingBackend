package main

import (
	"github.com/SeoPPak/FullStackProgrammingBackend/tree/master/auth/server"
)

func StartServer() {
	r := server.Setup()
	r.Run(":5000")
}
