package main

import (
	"fmt"

	"github.com/simjinhyun/x"
)

func initialize() {
	fmt.Println("Initialize: DB 연결, 캐시 준비 등")
}
func finalize() {
	fmt.Println("Finalize: 리소스 정리, 로그 flush 등")
}
func handler(c *x.Context) {
	fmt.Fprintln(c.Res, "YEP")
}
func OnShutdownErr(err error) {
	fmt.Println(err)
}
func main() {
	x.NewApp(
		initialize,
		handler,
		finalize,
		OnShutdownErr,
	).Run()
}
