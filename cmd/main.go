package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const STATE_FILE = "data/state.json"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found")
	}
	ctx := context.TODO()

	s, err := LoadServerState(ctx)
	if err != nil {
		log.Fatalf("Error loading server state: %v", err)
	}

	http.HandleFunc("/", s.HandleIndex)

	http.HandleFunc("/app.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./www/app.css")
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println(http.ListenAndServe(":8069", nil))
}

type Server struct {
	Title     string
	CacheBust string
	Templates *template.Template
	Data      *JsonData
}

// All the persistent data about the game will go here
type JsonData struct {
	Temp string
}

type IndexStruct struct {
	*Server
}

// This function will be useful later when extracting data from client requests
func ReadAndUnmarshal(w http.ResponseWriter, r *http.Request, reqBody interface{}) error {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return err
	}
	defer r.Body.Close()

	err = json.Unmarshal(bytes, reqBody)
	if err != nil {
		log.Printf("json: %v", string(bytes))
		log.Printf("Error parsing json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	return nil
}

func (s Server) SaveState() error {
	data, err := json.Marshal(s.Data)
	if err != nil {
		log.Println("Error marshalling state")
		return err
	}
	log.Println("Saving state")
	if err := os.WriteFile(STATE_FILE, data, 0666); err != nil {
		log.Println("Error writing state state")
		return err
	}
	return nil
}

func LoadServerState(context context.Context) (*Server, error) {
	var serverState Server
	var err error

	app_title := os.Getenv("TITLE")

	var state JsonData
	if _, err := os.Stat(STATE_FILE); err != nil {
		log.Println("Initialising state")
		state = JsonData{
			Temp: "Temp string",
		}
	} else {
		data, err := os.ReadFile(STATE_FILE)
		if err != nil {
			log.Printf("Couldn't read state file: %v", err)
			return nil, err
		}
		if err = json.Unmarshal(data, &state); err != nil {
			log.Printf("Couldn't unmarshal state file: %v", err)
			return nil, err
		}
	}

	serverState = Server{
		Title:     app_title,
		CacheBust: fmt.Sprintf("%v", time.Now().UnixNano()),
		Templates: template.New("").Funcs(template.FuncMap{}),
		Data:      &state,
	}
	serverState.Templates, err = serverState.Templates.ParseGlob("./www/*.html")
	if err != nil {
		return nil, err
	}
	return &serverState, nil
}

func (s *Server) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	err := s.Templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		log.Printf("Error executing template: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	data := IndexStruct{
		Server: s,
	}
	s.renderTemplate(w, "index", data)
}
