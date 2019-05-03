package main

import (
	"fmt"
	"log"
	"net/http"
)

func getToken(w http.ResponseWriter, r *http.Request) {
	log.Printf("%#v", r)

	token := maps.GetToken()
	fmt.Fprintf(w, "%s", token)
}
