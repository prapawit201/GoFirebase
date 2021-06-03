package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect" //check type like typeOf
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/prapawit201/GoFirebase/models"
	"github.com/twinj/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type App struct {
	Router *mux.Router
	client *firestore.Client
	ctx    context.Context
}

func main() {
	godotenv.Load()
	route := App{}
	route.Init()
	route.Run()
}

func (route *App) Init() {

	route.ctx = context.Background()

	sa := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(route.ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	route.client, err = app.Firestore(route.ctx)
	if err != nil {
		log.Fatalln(err)
	}
	route.Router = mux.NewRouter()
	route.initializeRoutes()
	fmt.Println("Successfully connected at port : " + route.GetPort())
}

func (route *App) GetPort() string {
	var port = os.Getenv("MyPort")
	if port == "" {
		port = "5000"
	}
	return ":" + port
}

func (route *App) Run() {
	log.Fatal(http.ListenAndServe(route.GetPort(), route.Router))
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func (route *App) initializeRoutes() {
	route.Router.HandleFunc("/", route.Home).Methods("GET")
	route.Router.HandleFunc("/{id}", route.FetchDataByIds).Methods("GET")
	route.Router.HandleFunc("/create", route.CreateBook).Methods("POST")
}

func (route *App) Home(w http.ResponseWriter, r *http.Request) {
	BooksData := []models.Books{}

	iter := route.client.Collection("books").Documents(route.ctx)
	for {
		BookData := models.Books{}
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			respondWithJSON(w, http.StatusInternalServerError, "Something wrong, please try again.")
		}

		mapstructure.Decode(doc.Data(), &BookData)
		BooksData = append(BooksData, BookData)
	}
	respondWithJSON(w, http.StatusOK, BooksData)
}

func (route *App) FetchDataByIds(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	paramId, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Failed to iterate: %v", err)
	}
	BooksData := []models.Books{}
	fmt.Println(reflect.TypeOf(paramId))
	iter := route.client.Collection("books").Where("id", "==", paramId).Documents(route.ctx)
	for {
		BookData := models.Books{}
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		mapstructure.Decode(doc.Data(), &BookData)
		BooksData = append(BooksData, BookData)
	}
	respondWithJSON(w, http.StatusOK, BooksData)
}

func (route *App) CreateBook(w http.ResponseWriter, r *http.Request) {
	uid := uuid.NewV4()
	splitID := strings.Split(uid.String(), "-")
	id := splitID[0] + splitID[1] + splitID[2] + splitID[3] + splitID[4]

	BookData := models.Books{}

	Decoder := json.NewDecoder(r.Body)
	err := Decoder.Decode(&BookData)

	BookData.Id = id
	fmt.Println(BookData)
	if err != nil {
		log.Printf("error: %s", err)
	}

	_, _, err = route.client.Collection("books").Add(route.ctx, BookData)
	if err != nil {
		log.Printf("An error has occurred: %s", err)
	}

	respondWithJSON(w, http.StatusCreated, "Create book success!")
}
