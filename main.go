package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//Person is a default data-object
type Person struct {
	ID        string   `json:"id,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Address   *Address `json:"address,omitempty"`
}

//Address cointains data bases on a simple address
type Address struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

var people []Person

func getAllUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	json.NewEncoder(w).Encode(people)
}

func getUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	for i := range people {
		if people[i].ID == ps.ByName("ID") {
			json.NewEncoder(w).Encode(people[i])
			break
		}
	}
}

func newUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	p := Person{}
	err := decoder.Decode(&p)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	people = append(people, p)
}

func main() {
	people = append(people, Person{ID: "1", Firstname: "John", Lastname: "Doe", Address: &Address{City: "City X", State: "State X"}})
	people = append(people, Person{ID: "2", Firstname: "Koko", Lastname: "Doe", Address: &Address{City: "City Z", State: "State Y"}})
	people = append(people, Person{ID: "3", Firstname: "Francis", Lastname: "Sunday"})

	router := httprouter.New()
	router.GET("/users", getAllUsers)
	router.GET("/users/:ID", getUser)
	router.POST("/users", newUser)

	log.Fatal(http.ListenAndServe(":8080", router))
}
