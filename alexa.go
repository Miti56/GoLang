package main

//imports needed to run the program
import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
)

// Client creation of struct to create client
type (
	Client struct {
		httpClient *http.Client
	}
)

//alexa Function in charge of connecting all previous programs to create an answer.wav
func alexa(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		if speech, ok := t["speech"].(string); ok {
			//map values
			u := map[string]string{"speech": speech}
			//Marshal returns the JSON encoding of u
			jsonValue, _ := json.Marshal(u)

			requestToStt, err := http.NewRequest("POST", "http://localhost:3002/stt", bytes.NewBuffer(jsonValue))
			if err != nil {
				errors.New("error")
			}
			//Client will forward all headers set on the initial Request
			client := &http.Client{}
			responseFromStt, err2 := client.Do(requestToStt)
			if err2 != nil {
				errors.New("error")
			}

			a := map[string]interface{}{}
			if err := json.NewDecoder(responseFromStt.Body).Decode(&a); err == nil {
				jsonValue1, _ := json.Marshal(a)
				//sending request to alpha
				requestToAlpha, err3 := http.NewRequest("POST", "http://localhost:3001/alpha", bytes.NewBuffer(jsonValue1))
				if err3 != nil {
					panic(err3)
				}
				//Client will forward all headers set on the initial Request
				client := &http.Client{}
				//getting response from alpha
				responseFromAlpha, err4 := client.Do(requestToAlpha)
				if err4 != nil {
					errors.New("error")
				}
				//map values
				p := map[string]interface{}{}
				if err := json.NewDecoder(responseFromAlpha.Body).Decode(&p); err == nil {
					//Marshal returns the JSON encoding of p
					jsonValue2, _ := json.Marshal(p)
					//sending request to tts
					requestToTts, err5 := http.NewRequest("POST", "http://localhost:3003/tts", bytes.NewBuffer(jsonValue2))
					if err5 != nil {
						errors.New("error request TTS")
					}
					//Client will forward all headers set on the initial Request
					client := &http.Client{}
					//sending request from tts
					responseFromTts, err6 := client.Do(requestToTts)
					if err6 != nil {
						errors.New("error response")
					}
					//map values
					t := map[string]interface{}{}
					//handling response from TTS if no faults found
					if err := json.NewDecoder(responseFromTts.Body).Decode(&t); err == nil {
						if speech1, ok := t["speech"].(string); ok {
							//mapping values answer
							y := map[string]string{"speech": speech1}
							w.WriteHeader(http.StatusOK)
							err := json.NewEncoder(w).Encode(y)
							if err != nil {
								errors.New("error")
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

//function that handles the functions
func main() {
	//matches incoming requests against a list of registered routes and calls a handler for the route that matches the URL
	r := mux.NewRouter()
	// string to let know the program what to expect + protocol used
	r.HandleFunc("/alexa", alexa).Methods("POST")
	//call to serve with handler to handle requests on tcp network
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		errors.New("error")
	}
}
