package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

//Person is a default data-object
type Person struct {
	ID        string   `json:"id,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Balance   int      `json:"balance,omitempty"`
	Address   *Address `json:"address,omitempty"`
}

//Address cointains data bases on a simple address
type Address struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

//Transaction contains the needed data for a simple transaction from account A to account B with amount C
type Transaction struct {
	Amount   int    `json:"amount"`
	TargetID string `json:"targetid"`
}

var people []Person
var tokensignkey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

func testRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "Hello %v! \n", r.TLS.PeerCertificates[0].Subject.CommonName)
}

func createToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"foo": "bar",
		"nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	tokenString, _ := token.SignedString(tokensignkey)

	fmt.Fprintf(w, tokenString)
}

func checkToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	parsedToken, _ := jwt.Parse(r.Header.Get("Authorization"), func(_ *jwt.Token) (interface{}, error) {
		return &tokensignkey.PublicKey, nil
	})

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		fmt.Fprintf(w, "token is valid\n")
		fmt.Fprintf(w, "%s, %f", claims["foo"], claims["nbf"])

	} else {
		fmt.Fprintf(w, "invalid")
	}
}

func getUserByID(id string) (p *Person) {
	for i := range people {
		if people[i].ID == id {
			p = &people[i]
			break
		}
	}
	return
}

func getAllUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	json.NewEncoder(w).Encode(people)
}

func getUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	u := getUserByID(ps.ByName("ID"))
	json.NewEncoder(w).Encode(u)
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

func deleteUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	for i := range people {
		if people[i].ID == ps.ByName("ID") {
			people = append(people[:i], people[i+1:]...)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func newTransaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.TLS.PeerCertificates[0].Subject.CommonName != ps.ByName("ID") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	t := Transaction{}
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	doTransaction(ps.ByName("ID"), t)
}

func doTransaction(fromID string, t Transaction) (e error) {
	fromUser := getUserByID(fromID)
	toUser := getUserByID(t.TargetID)

	if fromUser.Balance > t.Amount {
		fromUser.Balance -= t.Amount
		toUser.Balance += t.Amount
		e = nil
		return
	}

	return
}

func main() {
	people = append(people, Person{ID: "alice", Firstname: "Alice", Lastname: "Doe", Balance: 200, Address: &Address{City: "City X", State: "State X"}})
	people = append(people, Person{ID: "bob", Firstname: "Bob", Lastname: "Doe", Balance: 300, Address: &Address{City: "City Z", State: "State Y"}})
	people = append(people, Person{ID: "charlie", Firstname: "Charlie", Lastname: "Doe", Balance: 400})

	router := httprouter.New()
	router.GET("/users", getAllUsers)
	router.GET("/users/:ID", getUser)
	router.POST("/users", newUser)
	router.DELETE("/users/:ID", deleteUser)
	router.POST("/transaction/:ID", newTransaction)
	router.GET("/test", testRequest)
	router.GET("/authenticate", createToken)
	router.GET("/testtoken", checkToken)

	certPath := "c:\\temp\\server.pem"
	keyPath := "c:\\temp\\server.key"

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequestClientCert,
		},
	}

	log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
}
