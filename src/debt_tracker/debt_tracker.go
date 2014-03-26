package main

import (
    "flag"
	"fmt"
    "net/http"
    "log"
)

var port = flag.Int("port", 8080, "Port to listen on.")

func root_handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "request to %s\n<br/>", r.URL.Path[1:])
    fmt.Fprintf(w, "%v\n<br/>", r)
    r.ParseForm()
    fmt.Printf("request to %s\n", r.URL.Path[1:])
    fmt.Printf( "%v\n", r)
    fmt.Printf( "%v\n", r.Form["Text"])
}

func main() {
	flag.Parse()
    http.HandleFunc("/", root_handler)
    addr := fmt.Sprintf(":%v", *port)
    fmt.Println("Listening on", addr);
    err := http.ListenAndServe(addr , nil)
    if err != nil {
        log.Fatal("Couldn't listen: ", err)
    }
}
