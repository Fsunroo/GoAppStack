package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var dbMutex sync.RWMutex

func main() {
	dbConnection := "root:password@tcp(proxysql:6033)/sbtest"
	var err error
	db, err = sql.Open("mysql", dbConnection)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := createKeyValueTable(); err != nil {
		fmt.Printf("Failed to create key_value table: %v", err)
	}
	http.HandleFunc("/set", handleSet)
	http.HandleFunc("/get", handleGet)

	port := ":8080"
	fmt.Printf("Server listening on port %s...\n", port)
	http.ListenAndServe(port, nil)
}

func handleSet(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	values,ok := query["value"]
	keys,ok := query["key"]
	fmt.Println(query)
	if !ok || len(values[0]) < 1 {
		http.Error(w, "Missing 'value' parameter", http.StatusBadRequest)
		return
	}
	
	if !ok || len(keys[0]) < 1 {
		fmt.Println(keys)
		http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
		return
	}

	key := keys[0]
	value := values[0]

	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := db.Exec("INSERT INTO key_value (key_name, value) VALUES (?, ?) ON DUPLICATE KEY UPDATE value = ?", key, value, value)
	if err != nil {
		fmt.Print(err)
		http.Error(w, "Error storing data", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Stored: %s - %s\n", key, value)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
		return
	}

	key := keys[0]

	dbMutex.RLock()
	defer dbMutex.RUnlock()

	var value string
	err := db.QueryRow("SELECT value FROM key_value WHERE key_name = ?", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error retrieving data", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Value for key %s: %s\n", key, value)
}

func createKeyValueTable() error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS key_value (id INT AUTO_INCREMENT PRIMARY KEY,key_name VARCHAR(255) NOT NULL UNIQUE,value VARCHAR(255) NOT NULL);")
	return err
}