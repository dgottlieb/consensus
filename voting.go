package main

import (
	server "./server"
	"math/rand"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(4)
	rand.Seed(time.Now().Unix())

	server.Play()
}
