package main

//imports needed to run the program
import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"net/url"
)

// Client creation of struct to create client
type (
	Client struct {
		httpClient *http.Client
	}
)

//constant used during the program
const (
	URI = "http://api.wolframalpha.com/v1/result"
	KEY = "293GEK-XJHEWR29AJ"
)

//Creation of the function alpha
func alpha(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		// query --> What is the melting point of silver?
		if query, ok := t["text"].(string); ok {
			if answer, err := Service(query); err == nil {
				u := map[string]interface{}{"text": answer}
				//let know user everything running smoothly
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(u)
				if err != nil {
					//error message sent to header
					w.WriteHeader(http.StatusBadRequest)
				}
			} else {
				//error message sent to header
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			//error message sent to header
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		//error message sent to header
		w.WriteHeader(http.StatusBadRequest)
	}
}

func Service(words string) (interface{}, error) {
	//Client will forward all headers set on the initial Request
	client := &http.Client{}
	//http://api.wolframalpha.com/v1/result?appid=293GEK-XJHEWR29AJ&i=What+is+the+melting+point+of+silver%3F
	//Expected output after joining all the parts into a single string
	link := URI + "?appid=" + KEY + "&i=" + url.QueryEscape(words)
	//check that request is correct then proceed
	if req, err := http.NewRequest("GET", link, nil); err == nil {
		//check answer is correct then proceed
		if rsp, err := client.Do(req); err == nil {
			//check that everything is working
			if rsp.StatusCode == http.StatusOK {
				//read the answer from Api Response
				answer, err := io.ReadAll(rsp.Body)
				if err != nil {
					log.Fatalln(err)
				}
				return string(answer), nil
			}
		}
	}
	return nil, errors.New("api error")
}

//function that handles
func main() {
	//matches incoming requests against a list of registered routes and calls a handler for the route that matches the URL
	r := mux.NewRouter()
	// string to let know the program what to expect + protocol used
	r.HandleFunc("/alpha", alpha).Methods("POST")
	//call to serve with handler to handle requests on tcp network
	err := http.ListenAndServe(":3001", r)
	if err != nil {
		log.Fatalln(err)
	}
}
