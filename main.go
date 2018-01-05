package main

import (
	// "drafter/app"
	// "net/http"
	"drafter/app"
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("hello")

	http.ListenAndServe(":2000", app.R)
}
