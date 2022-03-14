package main

//imports needed to run the program
import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
)

//values used for microservice integration
const (
	REGION = "uksouth"
	URI    = "https://" + REGION + ".tts.speech.microsoft.com/" +
		"cognitiveservices/v1"
	KEY = "d76745e51adf4408b1f29d7a4362dc39"
)

// Client creation of struct to create client
type (
	Client struct {
		httpClient *http.Client
	}
)

// SSML Creation of the variables needed by the Azure API
type SSML struct {
	XMLName  xml.Name  `xml:"speak"`
	Version  string    `xml:"version,attr"`
	Language string    `xml:"xml:lang,attr"`
	Voice    SsmlVoice `xml:"voice"`
}

// SsmlVoice Creation of the variables needed by the Azure API
type SsmlVoice struct {
	XMLName  xml.Name `xml:"voice"`
	Language string   `xml:"xml:lang,attr"`
	Name     string   `xml:"name,attr"`
	Voice    string   `xml:",chardata"`
}

// textToSpeech Function to convert the text to speech thanks to the Microsoft API
func textToSpeech(text []byte) ([]byte, error) {
	//Client will forward all headers set on the initial Request
	client := &http.Client{}

	req, err := http.NewRequest("POST", URI, bytes.NewBuffer(text))
	if err != nil {
		errors.New("request error")
	}
	//set the header with the info needed by Azure
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("Ocp-Apim-Subscription-Key", KEY)
	req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")
	//send request to microservice
	rsp, err2 := client.Do(req)
	if err2 != nil {
		errors.New("request error")
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
		body, err3 := ioutil.ReadAll(rsp.Body)
		if err3 != nil {
			return nil, errors.New("error reading information")
		}
		return body, nil
	} else {
		return nil, errors.New("cannot convert text to speech")
	}
}

func tts(w http.ResponseWriter, r *http.Request) {
	//map values
	t := map[string]interface{}{}
	//If no error found, start function
	if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
		//if text correct, start decoding
		if text, ok := t["text"].(string); ok { // text variable -> What is the melting point of silver?
			//Use the struct to fill the information needed to configure the API
			var x SsmlVoice
			x.Language = "en-US"
			x.Name = "en-US-JennyNeural"
			x.Voice = text
			var y SSML
			y.Version = "1.0"
			y.Language = "en-US"
			y.Voice = x
			//create XML and put it in header with
			//Marshal returns XML encoding of y
			createXml, _ := xml.Marshal(y)
			byteEncodedText := []byte(xml.Header + string(createXml))

			if apiResult, err := textToSpeech(byteEncodedText); err == nil {
				//map the result to the string
				u := map[string]interface{}{"speech": apiResult}
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(u)
				if err != nil {
					//error message sent to header
					w.WriteHeader(http.StatusInternalServerError)
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
	r.HandleFunc("/tts", tts).Methods("POST")
	//call to serve with handler to handle requests on tcp network
	err := http.ListenAndServe(":3003", r)
	if err != nil {
		return
	}
}
