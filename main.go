package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"
	"net/http"
	"os"
	"bytes"
	"strings"
	"os/exec"
	"thezone/engine/lib/php"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)
const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04"
)
type Event struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Date        string  `json:"date"`
	EndDate	    string  `json:"end_date,omitempty"`
	StartTime   string  `json:"start_time,omitempty"`
	EndTime     string  `json:"end_time,omitempty"`
	Description string  `json:"description,omitempty"`
	Dyna        int     `json:"FK,omitempty"`
	Duration    float32 `json:"dur,omitempty"`
}

type Day struct {
	Date        string  `json:"date,omitempty"`  
	Events      []Event `json:"events,omitempty"`
	Diary       string  `json:"diary,omitempty"`
}

type Month struct {
	Days        []Day   `json:"days"`
	Objective   string  `json:"objective,omitempty"`
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
func validateEvent(e Event) error {
	if e.EndDate != "" {
		startDate, err := time.Parse(dateLayout, e.Date)
		if err != nil {
			return fmt.Errorf("invalid start date format: %v", err)
		}
		endDate, err := time.Parse(dateLayout, e.EndDate)
		if err != nil {
			return fmt.Errorf("invalid end date format: %v", err)
		}
		if endDate.Before(startDate) {
			return fmt.Errorf("end date must be after start date")
		}
	}

	if e.StartTime != "" && e.EndTime != "" {
		startTime, err := time.Parse(timeLayout, e.StartTime)
		if err != nil {
			return fmt.Errorf("invalid start time format: %v", err)
		}
		endTime, err := time.Parse(timeLayout, e.EndTime)
		if err != nil {
			return fmt.Errorf("invalid end time format: %v", err)
		}
		if !endTime.After(startTime) {
			return fmt.Errorf("End time must be after start time")
		}
	}
	return nil
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
	w.Header().Set("Content-Type", "application/json")
	var e Event
	if err := json.Unmarshal(body, &e); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	fmt.Printf("Received event: %+v\n", e)
	if err := validateEvent(e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonData, err := json.Marshal(e)
	if err != nil {
		http.Error(w, "Failed to marshal event", http.StatusInternalServerError)
		return
	}
	cmd := exec.Command("python3", "Sched/op.py", "--data", string(jsonData))
	var out bytes.Buffer
	var stderr bytes.Buffer
	
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	
	err = cmd.Run()
	if err != nil {
		log.Println("Python script failed:", err)
		log.Println("stderr:", stderr.String()) // prints error output from Python
		http.Error(w, "Python script failed: "+stderr.String(), http.StatusInternalServerError)
		return
	}
	
	log.Println("stdout:", out.String())
		
	// Split the output into lines
	outputLines := strings.Split(out.String(), "\n")

	var jsonLines []string
	for _, line := range outputLines {
		// Check if the line contains a valid JSON object (i.e., contains '{' and '}')
		if strings.Contains(line, "{") && strings.Contains(line, "}") {
			jsonLines = append(jsonLines, line)
		}
	}

	// If no valid JSON lines found
	if len(jsonLines) == 0 {
		http.Error(w, "No valid JSON found in the Python output", http.StatusInternalServerError)
		return
	}

	// Log all the found JSON lines
	log.Println("Found JSON lines from Python output:", jsonLines)


	var events []Event

	for _, jsonLine := range jsonLines {
		var event Event

		// Unmarshal each JSON line into the event struct
		if err := json.Unmarshal([]byte(jsonLine), &event); err != nil {
			http.Error(w, "Invalid JSON from Python: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Append to events slice
		events = append(events, event)
	}

	// Optionally, return the events as a response
	response, err := json.Marshal(events)
	if err != nil {
		http.Error(w, "Failed to marshal events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

func readEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	month := query.Get("month")
	day := query.Get("day")
	var year = time.Now().Year()
	if month == "" && day == "" {
		day = time.Now().Format("2025-04-07")
	}

	w.Header().Set("Content-Type", "application/json")

	if useMockData {
		if month != "" {
			mockMonth := Month{
				Days: []Day{
					{Date: "01-" + month + "-2024"},
					{Date: "02-" + month + "-2024"},
				},
				Objective: "Mock objective for month",
			}
			json.NewEncoder(w).Encode(mockMonth)
			return
		}

		if day != "" {
			mockDay := Day{
				Events: []Event{
					{
						ID:          99,
						Title:       "Mock Event for Day",
						Date:        day,
						StartTime:   "09:00",
						EndTime:     "10:00",
						Description: "Mocked event details",
					},
				},
				Diary: "This is a mock diary entry.",
			}
			json.NewEncoder(w).Encode(mockDay)
			return
		}
	}

	// ---- REAL DATABASE HANDLING ----
	if month != "" {
		var objective string
		err := db.QueryRow("SELECT objective FROM months WHERE M = ? and Y = ?", month, year).Scan(&objective)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "Error fetching month", http.StatusInternalServerError)
			return
		}

		rows, err := db.Query("SELECT dt, diary FROM days WHERE M = ? and Y = ?", month,year)
		if err != nil {
			http.Error(w, "Error fetching days for month", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var days []Day
		for rows.Next() {
			var d Day
			if err := rows.Scan(&d.Date, &d.Diary); err != nil {
				http.Error(w, "Failed to scan day", http.StatusInternalServerError)
				return
			}
			log.Println("Fetched day:", d.Date)
			days = append(days, d)
		}
		
		monthData := Month{
			Days:      days,
			Objective: objective,
		}
		json.NewEncoder(w).Encode(monthData)
		return
	}

	if day != "" {
		day, err := time.Parse("2006-01-02", day)
		var diary string
		_ = db.QueryRow("SELECT diary FROM days WHERE dt = ?", day).Scan(&diary)

		rows, err := db.Query("SELECT id, title, dt, start_time, end_time, des FROM task WHERE dt = ?", day)
		if err != nil {
			log.Println("Day fetching error", err)
			http.Error(w, "Error fetching events for day", http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		var events []Event
		for rows.Next() {
			var e Event
			if err := rows.Scan(&e.ID, &e.Title, &e.Date, &e.StartTime, &e.EndTime, &e.Description); err != nil {
				http.Error(w, "Failed to scan event", http.StatusInternalServerError)
				return
			}
			log.Println("Event Title:", e.Title)
			events = append(events, e)
		}

		dayData := Day{
			Events: events,
			Diary:  diary,
		}
		json.NewEncoder(w).Encode(dayData)
		return
	}
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
	result, err := db.Exec("DELETE FROM task WHERE id = ?", id)
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

	result, err := db.Exec("UPDATE task SET title=?, dt=?, start_time=?, end_time=?, des=? WHERE id=?", 
		e.Title, e.Date, e.StartTime, e.EndTime, e.Description, e.ID)
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

func handleRoot(response http.ResponseWriter, request *http.Request)() {
	output, err := php.Exec("health.php")
	if err != nil {
		log.Println("Health check failed: %v\nOutput: %s", err, output)
	} else {
	log.Println("I dont know what does it do, but it works:", string(output))
	}
	fmt.Fprintf(response, output)
}

func eventRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		readEvents(w, r)
	case http.MethodPost:
		createEvent(w, r)
	case http.MethodPut:
		updateEvent(w, r)
	case http.MethodDelete:
		deleteEvent(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	initDB()
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	// Register routes with middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/",corsMiddleware(loggingMiddleware(handleRoot)))
	mux.HandleFunc("/event", corsMiddleware(loggingMiddleware(eventRouter)))
	// Configure server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start server
	log.Println("Server is running on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
