package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"time"
)

// Get the Port from the environment so we can run on Heroku
func getPort() string {
	var port = os.Getenv("PORT")
	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "4747"
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}

// checkChan is the channel that we'll use to send Services.
// This let's us kick-off a bunch of checks and get back the statuses
// async.
var checkChan = make(chan Service)

// Service is the struct that contains the information of where the
// service should be checked, as well as the last known status and date
// checked. The URL is ommitted from JSON marshalling.
type Service struct {
	Name       string    `json:"name,omitempty"`
	Status     string    `json:"status,omitempty"`
	Last_Check time.Time `json:"last_check,omitempty"`
	Url        string    `json:"-"`
	Connection string    `json:"-"`
}

// getServices sets up the services that we'd like to check and
// returns and array of those services.
func getServices() []Service {
	services := make([]Service, 3)
	services[0] = Service{
		Name:       "Web Frontend",
		Url:        "https://www.ircrelay.com",
		Connection: "HTTP",
	}
	services[1] = Service{
		Name:       "Provisioning API",
		Url:        "https://ircrelay-api-production.herokuapp.com",
		Connection: "HTTP",
	}
	services[2] = Service{
		Name:       "IRC Router",
		Url:        "irc.ircrelay.com:6667",
		Connection: "TCP",
	}
	return services
}

func failCheck(service Service) {
	service.Status = "Experiencing Issues"
	checkChan <- service
}

func passCheck(service Service) {
	service.Status = "Fully Operational"
	checkChan <- service
}

// checkService makes an http request to the service and
// sets a status on the Service struct. If the response status code
// is anything other than 200, it gets a failing status.
func checkService(service Service) {
	if service.Connection == "HTTP" {
		resp, err := http.Get(service.Url)
		now := time.Now().UTC()
		service.Last_Check = now
		if err != nil {
			fmt.Println("Check Failed:", err)
			failCheck(service)
		}
		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				passCheck(service)
			} else {
				failCheck(service)
			}
		} else {
			failCheck(service)
		}
	}
	if service.Connection == "TCP" {
		conn, err := net.Dial("tcp", service.Url)
		now := time.Now().UTC()
		service.Last_Check = now
		if err != nil {
			fmt.Println("Check Failed:", err)
			failCheck(service)
		}
		if conn != nil {
			conn.Close()
			passCheck(service)
		} else {
			failCheck(service)
		}
	}
}

// getChecks kicks off, and then waits for all of their messaages to come
// back down the checkChan. It returns an array of checked services.
func getChecks(services []Service) []Service {
	for _, service := range services {
		go checkService(service)
	}
	for i := 0; i < 3; i++ {
		services[i] = <-checkChan
	}
	return services
}

// IndexHandler handles requests to the index page. This will also
// catch all other requests, no matter what path. That's ok, in this case.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Println("Error occured while rendering template:", err)
	}
	w.Header().Set("Content-Type", "text/html")
	now := time.Now().UTC()
	t.Execute(w, now)
	fmt.Println("200", r.URL, now)
}

// CheckHandler returns a JSON response of the checked services.
// It does this by kicking off the checks and blocking until a response
// is returned.
func CheckHandler(w http.ResponseWriter, r *http.Request) {
	services := getServices()
	checked_services := getChecks(services)
	j, err := json.MarshalIndent(checked_services, "", "  ")
	if err != nil {
		fmt.Println("error encoding json:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(j))
	now := time.Now().UTC()
	fmt.Println("200", r.URL, now)
}

// main runs the http server, and set's up the http handlers and routes.
func main() {
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/check", CheckHandler)
	fmt.Println("Starting web service...")
	http.ListenAndServe(getPort(), nil)
}
