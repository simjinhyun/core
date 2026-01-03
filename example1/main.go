package main

import (
	"github.com/simjinhyun/x"
)

func main() {
	a := x.NewApp("")
	a.Run("localhost:7000", 5)
}
