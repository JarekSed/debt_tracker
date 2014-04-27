package main

import (
    "bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
    "debt_tracker/data_model"
    "encoding/json"
)

var port = flag.Int("port", 8080, "Port to listen on.")
var plivo_number = flag.String("plivo_number", "14052659850", "Number registered with Plivo to send/recieve sms")
var plivo_auth_id = flag.String("plivo_auth_id", "MANTVJOGZKODI5ZGZLYW", "Auth ID registered with Plivo")
var plivo_auth_token = flag.String("plivo_auth_token", "ZjYzMjk3MTQwNWYwNWVkMWUwNGYxM2M3NDA4NWVl", "Auth token registered with Plivo")

func rootHandler(w http.ResponseWriter, r *http.Request, database *debt_tracker.Database) {
	r.ParseForm()
    fromNumber := r.Form["From"][0]
    sender, err := database.GetUserByPhoneNumber(fromNumber)
    if err != nil {
        log.Fatal("Error getting user from number: ", err)
    }else if sender == nil {
        SendMessage(fromNumber, "Who the fuck are you?")
        log.Print("Message from unknown number " ,fromNumber, ": ", r.Form["Body"])
    } else {
        fmt.Print(*sender, " sent that message!")
    }
}

func SendMessage(dest_number string, message string) {
    postUrl := fmt.Sprintf("https://%s:%s@api.plivo.com/v1/Account/%s/Message/", *plivo_auth_id, *plivo_auth_token, *plivo_auth_id)

    params := map[string]string {
        "src": *plivo_number,
        "dst": dest_number,
        "text": message,
    }
    body, err := json.Marshal(params)
    buf := bytes.NewBuffer(body)
    if err != nil {
        log.Fatal("error marshaling JSON: ", err)
    }
    // Submit form
    resp, err := http.Post(postUrl, "application/json", buf)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
}

func main() {
	flag.Parse()

    database, err := debt_tracker.ConnectToDatabase()
    if err != nil {
        log.Fatal("Error initializing data model:", err)
    }
	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
        rootHandler(w, r, database)
    })

	addr := fmt.Sprintf(":%v", *port)
	fmt.Println("Listening on", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("Couldn't listen: ", err)
	}
}
