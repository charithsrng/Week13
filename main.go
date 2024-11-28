package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Response struct {
	CurrentTime string `json:"current_time"`
	Timezone    string `json:"timezone"`
	Message     string `json:"message"`
}

type LoggedTimesResponse struct {
	Times []string `json:"times"`
}

var db *sql.DB

func currentTimeHandler(w http.ResponseWriter, r *http.Request) {

	now := time.Now()

	torontoLocation, err := time.LoadLocation("America/Toronto")
	if err != nil {
		http.Error(w, "Failed to load Toronto timezone", http.StatusInternalServerError)
		return
	}

	torontoTime := now.In(torontoLocation)
	fmt.Println(torontoTime)
	location, err := time.LoadLocation("America/Toronto")
	currentTime := time.Now().In(location)
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	_, err = db.Exec("INSERT INTO time_log (logged_time) VALUES (?)", formattedTime)
	if err != nil {
		log.Printf("Error inserting into database: %v", err)
		http.Error(w, "Failed to insert into database", http.StatusInternalServerError)
		return
	}

	// Create response
	response := Response{
		CurrentTime: torontoTime.Format(time.RFC1123),
		Timezone:    "America/Toronto",
		Message:     "Current time logged successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getLoggedTimesHandler(w http.ResponseWriter, r *http.Request) {
	// Query all logged times from the time_log table
	rows, err := db.Query("SELECT logged_time FROM time_log")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		http.Error(w, "Failed to retrieve logged times", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Collect results into a slice
	var times []string
	for rows.Next() {
		var loggedTime time.Time
		if err := rows.Scan(&loggedTime); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Failed to process database results", http.StatusInternalServerError)
			return
		}
		times = append(times, loggedTime.Format(time.RFC1123))
	}

	// Create response
	response := LoggedTimesResponse{Times: times}

	// Set headers and return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Connect to MySQL database
	var err error
	dsn := "root:Welcome123@tcp(127.0.0.1:3306)/Test"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error verifying connection to the database: %v", err)
	}
	log.Println("Connected to the database successfully!")

	http.HandleFunc("/current-time", currentTimeHandler)
	http.HandleFunc("/get-logged-times", getLoggedTimesHandler)

	port := "8080"
	log.Printf("Server is running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
