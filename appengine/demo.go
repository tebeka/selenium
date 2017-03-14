package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/tebeka/selenium"
)

func main() {
	http.HandleFunc("/", screenshotHandle)
	http.HandleFunc("/hello", helloHandle)
	http.HandleFunc("/_ah/health", healthCheckHandler)
	log.Print("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func screenshotHandle(w http.ResponseWriter, r *http.Request) {
	selenium.SetDebug(true)
	var output bytes.Buffer
	const (
		port            = 8081
		webDriverPath   = "selenium-server-standalone-3.3.1.jar"
		geckoDriverPath = "geckodriver-v0.15.0-linux64"
	)
	s, err := selenium.NewSeleniumService(webDriverPath, port, []selenium.ServiceOption{
		selenium.StartFrameBuffer(),
		selenium.Output(&output),
		selenium.GeckoDriver(geckoDriverPath),
	}...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting the WebDriver server with binary %q: %v", webDriverPath, err), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := s.Stop(); err != nil {
			http.Error(w, fmt.Sprintf("Error stopping the ChromeDriver service: %v", err), http.StatusInternalServerError)
		}
	}()

	caps := selenium.Capabilities{
		"browserName": "firefox",
	}
	addr := fmt.Sprintf("http://127.0.0.1:%d/wd/hub", port)
	wd, err := selenium.NewRemote(caps, addr)
	if err != nil {
		http.Error(w, fmt.Sprintf("NewRemote(%+v, %q) returned error: %v", caps, addr, err), http.StatusInternalServerError)
		return
	}
	defer wd.Quit()

	u := "http://www.google.com"
	if err := wd.Get(u); err != nil {
		http.Error(w, fmt.Sprintf("wd.Get(%q) returned error: %v", u, err), http.StatusInternalServerError)
		return
	}
	data, err := wd.Screenshot()
	if err != nil {
		http.Error(w, fmt.Sprintf("wd.Screenshot() returned error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	if _, err := w.Write(data); err != nil {
		http.Error(w, fmt.Sprintf("Write() returned error: %v", err), http.StatusInternalServerError)
		return
	}
}

func helloHandle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Hello world!")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}
