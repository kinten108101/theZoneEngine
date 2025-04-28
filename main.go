package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"
	"net/http"
	"math/rand"
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
	}else if day != "" {
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
func deleteAll(w http.ResponseWriter, r *http.Request){
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		_, err := db.Exec("DELETE FROM task")
		if err != nil {
			http.Error(w, "Failed to clear data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("Bruh bruh lmao delete ALLLLLLLL")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("All event data cleared."))
		return
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

func handleRoot(w http.ResponseWriter, r *http.Request)() {
	output, err := php.Exec("health.php")
	if err != nil {
		log.Println("Health check failed: %v\nOutput: %s", err, output)
	} else {
	log.Println("I dont know what does it do, but it works:", string(output))
	}
	fmt.Fprintf(w, output)

	today := time.Now()
    day := today.Day()

    q := r.URL.Query()
    q.Set("day", fmt.Sprintf("%d", day))
    r.URL.RawQuery = q.Encode()

    readEvents(w, r)

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
func randomTime() string {
	// Generate random hour (8:00 AM to 5:00 PM)
	hour := rand.Intn(10) + 8 // Hour between 8 and 17
	minute := rand.Intn(60)   // Minute between 0 and 59

	// Format time as "HH:MM AM/PM"
	t := time.Date(2025, time.May, 1, hour, minute, 0, 0, time.Local)
	return t.Format("03:04 PM")
}

// Calculate end time based on start time and random duration
func calculateEndTime(startTime string, duration float32) string {
	// Parse start time
	start, _ := time.Parse("03:04 PM", startTime)

	// Add duration (convert to hours and minutes)
	durationInMinutes := int(duration * 60)
	endTime := start.Add(time.Duration(durationInMinutes) * time.Minute)

	// Format end time as "HH:MM AM/PM"
	return endTime.Format("03:04 PM")
}

func useMock(w http.ResponseWriter, r *http.Request) {
	eventTitles := []string{
		"Team Meeting", "Client Meeting", "Code Review", "Brainstorming Session", "Development Work",
		"Feature Demo", "Task Assignment", "Sprint Planning", "Code Debugging", "Documentation Work",
	}

	eventDescriptions := []string{
		"Discuss project milestones and progress", "Review pull requests and improvements", "Work on new feature implementation",
		"Brainstorm ideas for the next sprint", "Update project documentation and diagrams", "Demo new feature to the team",
		"Assign tasks for the next sprint", "Debug and test the newly implemented features", "Plan tasks for the next phase",
		"Casual lunch and networking",
	}

	// Define the date range (27th April to 7th June)
	startDate := "2025-04-27"
	endDate := "2025-06-07"

	// Parse the start and end dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		http.Error(w, "Failed to parse start date", http.StatusInternalServerError)
		return
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		http.Error(w, "Failed to parse end date", http.StatusInternalServerError)
		return
	}

	// Generate 30 events
	var events []Event
	eventID := 1
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		for i := 0; i < 2; i++ { // Two events per day for simplicity
			// Random start time and duration
			startTime := randomTime()
			duration := float32(rand.Intn(3)+1) // Random duration between 1 and 3 hours

			event := Event{
				ID:          eventID,
				Title:       eventTitles[rand.Intn(len(eventTitles))],
				Date:        current.Format("2006-01-02"),
				StartTime:   startTime,
				EndTime:     calculateEndTime(startTime, duration),
				Description: eventDescriptions[rand.Intn(len(eventDescriptions))],
				Dyna:        rand.Intn(2), // Random Dyna value (0 or 1)
				Duration:    duration,
			}
			events = append(events, event)
			eventID++
		}
	}

	// Generate mock day data with events
	var days []Day
	for i := 0; i < len(events); i++ {
		// Check if the date already exists in days, if not, create it
		date := events[i].Date
		var day Day
		dayFound := false
		for j := range days {
			if days[j].Date == date {
				day = days[j]
				dayFound = true
				break
			}
		}

		// Add event to the day
		if !dayFound {
			day = Day{
				Date:   date,
				Events: []Event{events[i]},
				Diary:  "Productive day working on tasks and collaborating with the team.",
			}
			days = append(days, day)
		} else {
			day.Events = append(day.Events, events[i])
		}
	}

	// Convert to JSON
	response, err := json.Marshal(days)
	if err != nil {
		http.Error(w, "Failed to generate mock data", http.StatusInternalServerError)
		return
	}

	// Set the content type and send the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
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
	mux.HandleFunc("/event/clear", corsMiddleware(loggingMiddleware(deleteAll)))
	mux.HandleFunc("/mock",corsMiddleware(loggingMiddleware(useMock)))
	// Configure server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start server
	log.Println("Server is running on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
