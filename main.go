package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var checkChan = make(chan Service)

type Service struct {
	Name       string    `json:name,omitempty`
	Status     string    `json:status,omitempty`
	Last_Check time.Time `json:last_check,omitempty`
	Url        string    `json:-`
}

func getServices() []Service {
	services := make([]Service, 3)
	services[0] = Service{
		Name: "Web Frontend",
		Url:  "http://www.google.com",
	}
	services[1] = Service{
		Name: "Provisioning API",
		Url:  "http://www.google.com",
	}
	services[2] = Service{
		Name: "IRC Router",
		Url:  "http://www.google.com",
	}
	return services
}

func checkService(service Service) {
	resp, err := http.Get(service.Url)
	defer resp.Body.Close()
	now := time.Now().UTC()
	service.Last_Check = now
	if err != nil {
		fmt.Println("Check Failed:", err)
	}
	if resp.StatusCode == 200 {
		service.Status = "Fully Operational"
		checkChan <- service
	} else {
		service.Status = "Experiencing Issues"
		checkChan <- service
	}
}

func getChecks(services []Service) []Service {
	for _, service := range services {
		go checkService(service)
	}
	for i := 0; i < 3; i++ {
		services[i] = <-checkChan
	}
	return services
}

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
	services := getServices()
	checked_services := getChecks(services)
	j, err := json.MarshalIndent(checked_services, "", "  ")
	if err != nil {
		fmt.Println("error encoding json:", err)
	}
	fmt.Fprintf(w, string(j))
}

func main() {
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/check", CheckHandler)
	fmt.Println("Starting web service...")
	http.ListenAndServe(":4242", nil)
}
