package main

import (
  "fmt"
  "log"
  "net/http"
  "strconv"
  "encoding/json"

  "github.com/gorilla/mux"
  
  mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
  CONN_HOST     = "localhost"
  CONN_PORT     = "8080"
  MONGO_DB_URL  = "127.0.0.1"
)

var session *mgo.Session
var connectionError error

type Employee struct {
  Id    int     `json:"uid"`
  Name  string  `json:"name"`
}

func init() {
  session, connectionError = mgo.Dial(MONGO_DB_URL)
  if connectionError != nil {
    log.Fatal("error connecting to database :: ", connectionError)
  }
  session.SetMode(mgo.Monotonic, true)
}

func welcome(w http.ResponseWriter, r *http.Request) {
  log.Print("First connection to DB")
  fmt.Fprintf(w, "Welcome to this app.")
}

func readDocuments(w http.ResponseWriter, r *http.Request) {
	log.Print("reading documents from database")
	var employees []Employee

	collection := session.DB("mydb").C("employee")

	err := collection.Find(bson.M{}).All(&employees)
	if err != nil {
		log.Print("error occurred while reading documents from database :: ", err)
		return
	}
	json.NewEncoder(w).Encode(employees)
}


func createDocument(w http.ResponseWriter, r *http.Request) {
  vals := r.URL.Query()
  name, nameOk := vals["name"]
  id, idOk := vals["id"]
  if nameOk && idOk {
    employeeId, err := strconv.Atoi(id[0])
    if err != nil {
      log.Print("Error converting string id to int :: ", err)
      return 
    }
    log.Print("Inserting document in the DB for name :: ", name[0])
    collection := session.DB("mydb").C("employee")
    err = collection.Insert(&Employee{employeeId, name[0]})
    if err != nil {
      log.Print("Error occured while inserting document in database :: ", err)
      return
    }
    fmt.Fprintf(w, "Last created document id is :: %s", id[0])
  } else {
    fmt.Fprintf(w, "Error occured while creating document in database for name :: %s", name[0])
  }
}

func updateDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	vals := r.URL.Query()
	name, ok := vals["name"]

	if ok {
		employeeId, err := strconv.Atoi(id)
		if err != nil {
			log.Print("error converting string id to int :: ", err)
			return
		}
		log.Print("going to update document in database for id :: ", id)
		collection := session.DB("mydb").C("employee")
		var changeInfo *mgo.ChangeInfo
		changeInfo, err = collection.Upsert(bson.M{"id": employeeId}, &Employee{employeeId, name[0]})
		if err != nil {
			log.Print("error occurred while updating record in database :: ", err)
			return
		}
		fmt.Fprintf(w, "Number of documents updated in database are :: %d", changeInfo.Updated)
	} else {
		fmt.Fprintf(w, "Error occurred while updating document in database for id :: %s", id)
	}
}

func deleteDocument(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	name, ok := vals["name"]
	if ok {
		log.Print("going to delete document in database for name :: ", name[0])
		collection := session.DB("mydb").C("employee")
		removeErr := collection.Remove(bson.M{"name": name[0]})
		if removeErr != nil {
			log.Print("error removing document from database :: ", removeErr)
			return
		}
		fmt.Fprintf(w, "Document with name %s is deleted from database", name[0])
	} else {
		fmt.Fprintf(w, "Error occurred while deleting document in database for name :: %s", name[0])
	}
}

func main() {
  router := mux.NewRouter()
  router.HandleFunc("/", welcome).Methods("GET")
  router.HandleFunc("/employees", readDocuments).Methods("GET")
  router.HandleFunc("/employees", createDocument).Methods("POST")
  router.HandleFunc("/employees", updateDocument).Methods("PUT")
  router.HandleFunc("/employees", deleteDocument).Methods("DELETE")
  fmt.Println("Bonjour!")
  fmt.Println("Listening on : "+CONN_HOST+":"+CONN_PORT)
  defer session.Close()
  err := http.ListenAndServe(CONN_HOST+":"+CONN_PORT, router)
  if err != nil {
    log.Fatal("Error starting http server :: ", err)
    return
  } 
}
