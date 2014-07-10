package main

import (
	"math/rand"
	"runtime"
	"time"

	server "./server"
)

func main() {
	runtime.GOMAXPROCS(4)
	rand.Seed(time.Now().Unix())

	server.Play()
}
