package main

import (
    "database/sql"
    "encoding/json"
    "github.com/gorilla/mux"
    "log"
    "net/http"
    "strconv"
)

import _ "github.com/go-sql-driver/mysql"

type Contact struct {
  ID            int64     `json:"id,omitempty"`
  FirstName     string  `json:"first_name,omitempty"`
  LastName      string  `json:"last_name,omitempty"`
  Age           int     `json:"age,omitempty"`
  MobileNumber  string  `json:"mobile_number,omitempty"`
}

func GetContact(db *sql.DB) func (http.ResponseWriter, *http.Request) {
  return func (w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, _ := strconv.ParseInt(params["id"], 10, 64)
    contact, _ := QueryContactByID(db, id)
    json.NewEncoder(w).Encode(contact)
  }
}

func QueryContactByID(db *sql.DB, id int64) (Contact, error) {
  var contact Contact
  contact.ID = id
  err := db.QueryRow("SELECT first_name, last_name, age, mobile_number FROM contacts WHERE id=?", id).Scan(&contact.FirstName, &contact.LastName, &contact.Age, &contact.MobileNumber)
  if err != nil {
    log.Fatal(err)
  }
  return contact, err
}

func GetContacts(db *sql.DB) func (http.ResponseWriter, *http.Request) {
  return func (w http.ResponseWriter, r *http.Request) {
    contacts, err := QueryContacts(db)
    if err != nil {
      log.Fatal(err)
    }
    json.NewEncoder(w).Encode(contacts)
  }
}

func QueryContacts(db *sql.DB) ([]Contact, error) {
  var contacts []Contact
  rows, err := db.Query("select id, first_name, last_name, age, mobile_number from contacts")
  if err != nil {
  	log.Fatal(err)
  }
  defer rows.Close()
  for rows.Next() {
    var contact Contact
  	err := rows.Scan(&contact.ID, &contact.FirstName, &contact.LastName, &contact.Age, &contact.MobileNumber)
  	if err != nil {
  		log.Fatal(err)
  	}
  	contacts = append(contacts, contact)
  }
  err = rows.Err()
  if err != nil {
  	log.Fatal(err)
  }
  return contacts, err
}

func CreateContact(db *sql.DB) func (http.ResponseWriter, *http.Request) {
  return func (w http.ResponseWriter, r *http.Request) {
    var contact Contact
    _ = json.NewDecoder(r.Body).Decode(&contact)
    InsertContact(db, &contact)
    w.WriteHeader(201)
    json.NewEncoder(w).Encode(contact)
  }
}

func InsertContact(db *sql.DB, contact *Contact) error {
  res, err := db.Exec(
    "INSERT INTO contacts(first_name, last_name, age, mobile_number) VALUES(?, ?, ?, ?)",
    contact.FirstName, contact.LastName, contact.Age, contact.MobileNumber)
  contact.ID, err = res.LastInsertId()
  return err
}

func RemoveContact(db *sql.DB) func (http.ResponseWriter, *http.Request) {
  return func (w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, _ := strconv.ParseInt(params["id"], 10, 64)
    DeleteContact(db, id)
    w.WriteHeader(204)
  }
}

func DeleteContact(db *sql.DB, id int64) error {
  _, err := db.Exec(
    "DELETE FROM contacts WHERE id = ?", id)
  return err
}

func main() {
  /**
   CREATE DATABASE phonebook;
   CREATE TABLE contacts (
     id int not null auto_increment,
     first_name varchar(50),
     last_name varchar(50),
     age int,
     mobile_number varchar(20),
     primary key(id)
   );
   */
  db, err := sql.Open("mysql", "dbuser:changeme@tcp(localhost:3306)/phonebook")
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  router := mux.NewRouter()
  router.HandleFunc("/contacts", GetContacts(db)).Methods("GET")
  router.HandleFunc("/contacts/{id}", GetContact(db)).Methods("GET")
  router.HandleFunc("/contacts", CreateContact(db)).Methods("POST")
  router.HandleFunc("/contacts/{id}", RemoveContact(db)).Methods("DELETE")

  // TODO: derive the port number from arguments
  log.Fatal(http.ListenAndServe(":8000", router))
}
