package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Println("Error occured while rendering template:", err)
	}
	w.Header().Set("Content-Type", "text/html")
	now := time.Now().UTC()
	t.Execute(w, now)
	fmt.Println("200", now)
}

func CheckHandler(w http.ResponseWriter, r *http.Request) {
	// Checks go here.
}

func main() {
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/check", CheckHandler)
	fmt.Println("Starting web service...")
	http.ListenAndServe(":4242", nil)
}
