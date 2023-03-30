package main

import "time"

func main() {
	go ServeFunc()
	ClientFunc()
	time.Sleep(time.Second * 10)
}
