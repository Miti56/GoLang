package main

//imports needed to run the program
import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"io"
	"net/http"
)

// Client creation of struct to create client
type (
	Client struct {
		httpClient *http.Client
	}
)

//values used for microservice integration
const (
	REGION = "uksouth"
	URI    = "https://" + REGION + ".stt.speech.microsoft.com/" +
		"speech/recognition/conversation/cognitiveservices/v1?" +
		"language=en-US"
	KEY = "d76745e51adf4408b1f29d7a4362dc39"
)

//function used to convert speech to text and communicate API
func convertText(speech []byte) (string, error) {
	//map values
	t := map[string]interface{}{}
	//Client will forward all headers set on the initial Request
	client := &http.Client{}
	//send speech to API
	req, err := http.NewRequest("POST", URI, bytes.NewReader(speech))
	if err != nil {
		errors.New("error occurred")
	}
	//Give API information about information sent
	req.Header.Set("Content-Type",
		"audio/wav;codecs=audio/pcm;samplerate=16000")
	//Give information to API about our key
	req.Header.Set("Ocp-Apim-Subscription-Key", KEY)
	//send request to microservice
	rsp, err2 := client.Do(req)
	if err2 != nil {
		errors.New("error occurred")
	}

	//client must close connection with server when finished
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			errors.New("connection not closed")
		}
	}(rsp.Body)

	//If response from API is good, start
	if rsp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(rsp.Body).Decode(&t); err == nil {
			return t["DisplayText"].(string), nil
		}
		return "", errors.New("cannot convert to speech to text")
	} else {
		return "", errors.New("bad response from API")
	}
}

//function handling speech to text
func stt(w http.ResponseWriter, r *http.Request) {
	t := map[string]interface{}{}
	//If no error found, start function
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		//if speech.wav correct, start decoding
		if speech, ok := t["speech"].(string); ok {
			//convert speech into base64
			decodedSpeech, _ := base64.StdEncoding.DecodeString(speech)
			//If no error found, send user the API results
			if apiResult, err := convertText(decodedSpeech); err == nil {
				//map the result to the string
				u := map[string]interface{}{"text": apiResult}
				w.WriteHeader(http.StatusOK)
				//send result to the user
				err := json.NewEncoder(w).Encode(u)
				if err != nil {
					//error message sent to header
					w.WriteHeader(http.StatusBadRequest)
				}
			} else {
				//error message sent to header
				w.WriteHeader(http.StatusInternalServerError)
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
	r.HandleFunc("/stt", stt).Methods("POST")
	//call to serve with handler to handle requests on tcp network
	err := http.ListenAndServe(":3002", r)
	if err != nil {
		errors.New("error")
	}
}
