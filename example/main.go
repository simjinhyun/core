package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/simjinhyun/x"
)

func main() {
	app := x.NewApp()
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	fmt.Println("실행파일 경로:", exePath)
	dir := filepath.Dir(exePath)
	app.Conf["WebRoot"] = filepath.Join(dir, "../www")
	app.Run()
}
