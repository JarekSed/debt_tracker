package main

import (
    "bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
    "strings"
    "strconv"
    "debt_tracker/data_model"
    "encoding/json"
    "regexp"
)

var port = flag.Int("port", 8080, "Port to listen on.")
var plivo_number = flag.String("plivo_number", "14052659850", "Number registered with Plivo to send/recieve sms")
var plivo_auth_id = flag.String("plivo_auth_id", "MANTVJOGZKODI5ZGZLYW", "Auth ID registered with Plivo")
var plivo_auth_token = flag.String("plivo_auth_token", "ZjYzMjk3MTQwNWYwNWVkMWUwNGYxM2M3NDA4NWVl", "Auth token registered with Plivo")

func rootHandler(w http.ResponseWriter, r *http.Request, database *debt_tracker.Database) {
	r.ParseForm()
    fromNumber := r.Form["From"][0]
    message := strings.Join(r.Form["Text"], " ")
    sender, err := database.GetUserByPhoneNumber(fromNumber)
    if err != nil {
        log.Fatal("Error getting user from number: ", err)
    }else if sender == nil {
        handleMessageFromUnknownNumber(fromNumber, message, database)
    } else {
        handleMessageFromUser(sender, message, database)
    }
}

// The sender's balance increases when he gets pays someone, or someone else owes him money.
func isSenderBalanceIncreaseMessage(message string) bool {
    // TODO: don't compile regex every time.
    var wasPaidRegex = regexp.MustCompile(`(Owes Me)|(I Paid)|(I Gave)`)
    return wasPaidRegex.MatchString(message)
}

// The sender's balance decreases when he gets paid, or he owes someone money
func isSenderBalanceDecreaseMessage(message string) bool {
    // TODO: don't compile regex every time.
    var wasPaidRegex = regexp.MustCompile(`(I Owe)|(Paid Me)|(Gave Me)`)
    return wasPaidRegex.MatchString(message)
}


func handleMessageFromUser(p *debt_tracker.Person, message string, database *debt_tracker.Database) {
    message = strings.Title(message)
    log.Print("Message from ", p.FullName(), ": ", message)

    if isSenderBalanceIncreaseMessage(message) {
        log.Println(p.FullName(), " is owed money")
    } else if isSenderBalanceDecreaseMessage(message) {
        log.Println(p.FullName(), " owes someone money")
    } else {
        log.Println("Unknown message...")
    }

}

func handleMessageFromUnknownNumber(fromNumber string, message string, database *debt_tracker.Database) {
    log.Print("Message from unknown number " ,fromNumber, ": ", message)
    words := strings.Fields(strings.Title(message))
    number, err := strconv.ParseUint(fromNumber, 10, 64)
    if err != nil {
        log.Fatal("Invalid from number: ", fromNumber)
    }
    // If this was a register request, register the new user
    // TODO: verify the number
    if len(words) == 3 && strings.ToLower(words[0]) == "register" {
        p := debt_tracker.Person{
            FirstName: words[1],
            LastName: words[2],
            PhoneNumber: number,
        }
        err := database.RegisterUser(p)
        if err != nil {
            log.Fatal("Error Registering user: ", err)
        } else {
            SendMessage(strconv.FormatUint(p.PhoneNumber, 10), fmt.Sprintf("Welcome %s!", p.FullName()))
            log.Printf("Added %s with phone number: %v\n", p.FullName(), p.PhoneNumber)
        }
    } else {
        SendMessage(fromNumber, "Who the fuck are you?")
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
