package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"os/exec"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Event struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Date        string `json:"date"`
	start_time        string `json:"start_time"`
	end_time        string `json:"end_time"`
	Description string `json:"description,omitempty"`
	Dyna      int   `json:"fk_id,omitempty"`
}

type Day struct {
	Events []Event `json:"events"`
	Diary  string  `json:"diary,omitempty"`
}

type Month struct {
	Days      []Day  `json:"days"`
	Objective string `json:"objective,omitempty"`
}

var db *sql.DB
var useMockData bool

func initDB() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found. Using default DB config")
	}

	connStr := os.Getenv("DBstring")
	if connStr == "None" || connStr == "" {
		log.Println("No database found. Running in mock data mode.")
		useMockData = true
		return
	}

	db, err = sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Database ping failed:", err)
	}
	log.Println("Connected to MySQL")
}

// CORS Middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	}
}

// Logging Middleware
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	}
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var e Event
	if err := json.Unmarshal(body, &e); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if useMockData {
		// In mock mode, return the event details back to the user
		e.ID = 999 // Mock ID
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(e)
		return
	}

	// Database insertion
	result, err := db.Exec("INSERT INTO events (title, date, start_time, end_time, description) VALUES (?, ?, ?, ?)", 
		e.Title, e.Date, e.start_time, e.end_time , e.Description)
	if err != nil {
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	// Get the ID of the newly inserted event
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to retrieve event ID", http.StatusInternalServerError)
		return
	}

	e.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}

func readEvents(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if useMockData {
		// Return constant mock data
		mockEvents := []Event{
			{
				ID:          1,
				Title:       "Mock Event 1",
				Date:        "28 -03- 2024",
				start_time:        "10:00",
				end_time: "12:00",
				Description: "This is a mock event for demonstration",
				// Dyna:      false,
			},
			{
				ID:          2,
				Title:       "Mock Event 2",
				Date:        "09-11-2011",
				start_time:        "10:00",
				end_time: "12:00",
				Description: "This is a mock event for demonstration",
				// Dyna:      false,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockEvents)
		return
	}

	rows, err := db.Query("SELECT id, title, date, time, description FROM events")
	if err != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Date, &e.start_time, &e.end_time, &e.Description); err != nil {
			http.Error(w, "Failed to scan event", http.StatusInternalServerError)
			return
		}
		events = append(events, e)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(events)
}


func deleteEvent(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from query parameter or path
	id := r.URL.Query().Get("id")
	if id == "" {
		// Try extracting from path if not in query
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) > 2 {
			id = pathParts[len(pathParts)-1]
		}
	}

	if id == "" {
		http.Error(w, "Missing event ID", http.StatusBadRequest)
		return
	}

	if useMockData {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("Mock mode: Event %s deleted successfully", id),
		})
		return
	}

	// Database deletion
	result, err := db.Exec("DELETE FROM events WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to delete event", http.StatusInternalServerError)
		return
	}

	// Check if any rows were actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking deletion", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "No event found with given ID", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Event deleted successfully",
	})
}

func updateEvent(w http.ResponseWriter, r *http.Request) {
	// Ensure only PUT method is accepted
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var e Event
	if err := json.Unmarshal(body, &e); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate that ID is provided
	if e.ID == 0 {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	if useMockData {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(e)
		return
	}

	result, err := db.Exec("UPDATE events SET title=?, date=?, start_time=?, end_time=?, description=? WHERE id=?", 
		e.Title, e.Date, e.start_time, e.end_time, e.Description, e.ID)
	if err != nil {
		http.Error(w, "Failed to update event", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking update", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "No event found with given ID", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Event updated successfully",
	})
}
func health() {
	cmd := exec.Command("php", "health.php")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Health check failed: %v\nOutput: %s", err, output)
	}
	log.Println("I dont know what does it do, but it works:", string(output))
}

func main() {
	health()
	initDB()
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	// Create a new serve mux
	mux := http.NewServeMux()

	// Register routes with middleware
	mux.HandleFunc("/event/create", corsMiddleware(loggingMiddleware(createEvent)))
	mux.HandleFunc("/event/read", corsMiddleware(loggingMiddleware(readEvents)))
	mux.HandleFunc("/event/delete", corsMiddleware(loggingMiddleware(deleteEvent)))
	mux.HandleFunc("/event/update", corsMiddleware(loggingMiddleware(updateEvent)))

	// Configure server
	server := &http.Server{
		Addr:    ":6969",
		Handler: mux,
	}

	// Start server
	log.Println("Server is running on http://localhost:6969")
	log.Fatal(server.ListenAndServe())
}