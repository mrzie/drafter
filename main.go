package main

import (
	"strconv"
	// "drafter/app"
	// "net/http"
	"drafter/app"
	. "drafter/setting"
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("hello")

	http.ListenAndServe(":"+strconv.Itoa(Settings.Port), app.R)
}
