package main

import (
	"log/slog"

	"github.com/simjinhyun/x"
)

func main() {
	a := x.NewApp("")
	a.SetLevel(slog.LevelDebug)
	a.Run("localhost:7000")
}
