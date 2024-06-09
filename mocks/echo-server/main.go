package bettertester

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	port  = flag.String("port", "8000", "port to listen on")
	name  = flag.String("name", "gocho", "server name")
	debug = flag.Bool("debug", false, "enable debug logging")
)

type Request struct {
	Datetime string
	Server   string
	Hostname string
	Method   string
	Path     string
	Query    map[string][]string
	Headers  map[string][]string
	Body     string
	Duration int64
	Status   int
}

func echo(w http.ResponseWriter, req *http.Request) {
	started := time.Now()
	if *debug {
		log.Print("[" + started.UTC().String() + "] " + req.Method + " " + req.URL.String())
	}

	r := Request{
		Datetime: started.UTC().String(),
		Server:   *name,
		Hostname: req.Host,
		Method:   req.Method,
		Path:     strings.Split(req.URL.String(), "?")[0],
		Query:    req.URL.Query(),
		Headers:  req.Header,
		Body:     "",
		Status:   200,
		Duration: time.Duration(0).Nanoseconds(),
	}
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("Body reading error: %v", err)
			return
		}
		r.Body = string(bodyBytes)
		defer req.Body.Close()
	}

	doDelay(r)
	r.Status = getStatus(r)
	w.WriteHeader(r.Status)

	r.Duration = time.Since(started).Nanoseconds()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "HEAD, OPTIONS, GET, PUT, POST, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if req.Method == "OPTIONS" || req.Method == "HEAD" || r.Status == 204 {
		w.Header().Set("Content-Length", "0")
		w.Write([]byte{})
	} else {
		rb, _ := json.Marshal(r)
		w.Header().Set("Content-Type", "application/json")
		w.Write(rb)
	}

	log.Printf("Returning response: %v\n", r)

}

func doDelay(r Request) {
	value, ok := r.Query["delay"]
	if ok && len(value) > 0 {
		delay, _ := time.ParseDuration(value[0] + "s")
		log.Printf("Delaying response by %v seconds\n", delay)
		time.Sleep(delay)
	}
}

func getStatus(r Request) int {
	status := 200
	value, ok := r.Query["status"]
	if ok && len(value) > 0 {
		status, _ = strconv.Atoi(value[0])
		if *debug {
			log.Printf("Returning explicit status: %v\n", status)
		}
	}
	return status
}

func main() {
	flag.Parse()
	log.Print("Running echo-ish server on " + *port)
	http.HandleFunc("/", echo)
	err := http.ListenAndServe(":"+*port, nil)
	log.Printf("%v", err)
}
