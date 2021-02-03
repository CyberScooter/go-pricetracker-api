package main

import (
	"fmt"
	"html/template"
	"encoding/json"
	"net/http"
	"log"
	"os"
	"regexp"
	"tracker-api-backend-go/api/tracker"
	"tracker-api-backend-go/api/structs"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/dpapathanasiou/go-recaptcha"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/mail.v2"

)


var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")


type Message struct {
	Email   string
	Content string
  }

var db *sql.DB

func setup() error {
    var err error
    db, err = sql.Open("mysql", "root:" + os.Getenv("DATABASE_PASS") + "@tcp(127.0.0.1:3306)/trackerapi")
    if err != nil {
        log.Fatal(err)
	}
	return err

	// Other setup-related activities
}

func (msg *Message) Deliver() error {
	email := mail.NewMessage()
	email.SetHeader("To", os.Getenv("EMAIL_USERNAME"))
	email.SetHeader("From", os.Getenv("EMAIL_USERNAME"))
	email.SetHeader("Reply-To", msg.Email)
	email.SetHeader("Subject", "New message via Contact Form")
	email.SetBody("text/plain", msg.Content)
  
	username := os.Getenv("EMAIL_USERNAME")
	password := os.Getenv("EMAIL_PASSWORD")
  
	return mail.NewDialer("smtp.gmail.com", 587, username, password).DialAndSend(email)
  }  

// Define our struct
type authenticationMiddleware struct {
	tokenUsers map[string]string
}

type UserAPI struct {
	Email   string    `json:"email"`
    APIKey string `json:"apiKey"`
}

// Initialize it somewhere
// runs every 5 minutes
func (amw *authenticationMiddleware) Populate() {
	for {
		// Execute the query
		results, err := db.Query("SELECT * FROM UserAPI")
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		for results.Next() {
			var userapi UserAPI
			// for each row, scan the result into our tag composite object
			err = results.Scan(&userapi.Email, &userapi.APIKey)
			if err != nil {
				panic(err.Error()) // proper error handling instead of panic in your app
			}
			// and then print out the tag's Name attribute
			amw.tokenUsers[userapi.APIKey] = userapi.Email
			// log.Printf(userapi.Email)
			// log.Printf(userapi.APIKey)
		}
		// amw.tokenUsers["00000000"] = "user0"
		// amw.tokenUsers["aaaaaaaa"] = "userA"
		// amw.tokenUsers["05f717e5"] = "randomUser"
		// amw.tokenUsers["deadbeef"] = "user0"
		fmt.Println("ran")
		time.Sleep(300 * time.Second)
	}

}

// Middleware function, which will be called for each request
func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
		if r.URL.Path == "/pricetracker" && r.Method == http.MethodPost{
			if err := r.ParseForm(); err != nil {
				fmt.Fprintf(w, "ParseForm() err: %v", err)
				return
			}
			recaptchaResponse, responseFound := r.Form["g-recaptcha-response"]
			if responseFound {
				rs, err := recaptcha.Confirm("127.0.0.1", recaptchaResponse[0])
				if err != nil {
					log.Println("recaptcha server error", err)
					http.Error(w, "Forbidden", http.StatusForbidden)
				}else {
					if rs == false{
						http.Error(w, "reCAPTCHA not completed", 403)
					}else {
						if emailRegex.MatchString(r.Form["email"][0]) && r.Form["name"][0] != "" {
							next.ServeHTTP(w, r)
						}else {
							http.Error(w, "Email in wrong format or name not filled in", 406)
						}
						
					}
					
				}
				
			}
		}else if r.URL.Path == "/pricetracker" && r.Method == http.MethodGet{
			next.ServeHTTP(w,r)
		}else {
			token := mux.Vars(r)["api-key"]

			if user, found := amw.tokenUsers[token]; found {
				// We found the token in our map
				log.Printf("Authenticated user %s\n", user)
				// Pass down the request to the next middleware (or final handler)
				next.ServeHTTP(w, r)
			} else {
				// Write an error and stop the handler chain
				http.Error(w, "Forbidden", http.StatusForbidden)
			}
		}

		// else if r.URL.Path == "/pricetracker"{
		// 	next.ServeHTTP(w,r)

    })
}

func priceTrackerAPIHandler(w http.ResponseWriter, r *http.Request){
	name := r.URL.Query().Get("item")


	if name != "" {
		var items []structs.Product = tracker.Track(name)
		if(items[0].Found == true){
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(items)
		}else {
			// w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			http.Error(w, "Not found", 406)
		}
	}else {
		w.WriteHeader(http.StatusNotFound)
		http.Error(w, "No data entered, add '?item=' query string to the end", 406)
	}
}

func indexPageHandler(w http.ResponseWriter, r *http.Request){

	tmpl := template.Must(template.ParseFiles("./api/forms/registration.html"))
	

	if r.Method != http.MethodPost {
		tmpl.Execute(w, struct { Success bool; Key string}{false , os.Getenv("RECAPTCHA_SITE")})
		return
	}

	// do things with email
	email := r.FormValue("email")
	
	msg := &Message{
		Email:   email,
		Content: fmt.Sprintf("User: %s under email: %s has shown an interest about the tracker API please reply", r.FormValue("name"), email),
	}

	if err := msg.Deliver(); err != nil {
		log.Println(err)
		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
		return
	}

	//if successful return
	tmpl.Execute(w, struct{ Success bool}{true})
}



func main(){
	// load .env file from given path
	// we keep it empty it will load .env from current directory
	envErr := godotenv.Load(".env")
	if envErr != nil {
		log.Fatalf("Error loading .env file")
	}


	err := setup()
    if err != nil {
        log.Fatal(err)
	}
	
	recaptcha.Init(os.Getenv("RECAPTCHA_SECRET"))

	r := mux.NewRouter()

	r.HandleFunc("/pricetracker", indexPageHandler)
	r.HandleFunc("/{api-key}/pricetracker", priceTrackerAPIHandler).Methods("GET")

	amw := authenticationMiddleware{make(map[string]string)}
	go amw.Populate()
	r.Use(amw.Middleware)
	
	log.Fatal(http.ListenAndServe(":8081", r))
	
}