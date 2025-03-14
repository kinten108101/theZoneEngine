package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"

	_ "github.com/go-sql-driver/mysql"
)
type Day struct{
	list []Event 
	event string
}
type Month struct{
	list []Day 
	Diary string
}
type Event struct {
	ID          int    `json:"id,omitempty"`
	Title       string `json:"title"`
	Date        string `json:"date"`
	Time        string `json:"time,omitempty"`
	Description string `json:"description,omitempty"`
	static bool
}

func initDB() {
	var db *sql.DB
	var err error
	connStr := "Assume we have some shit going on"
	var db, err = sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	fmt.Println("Connected to MySQL")
	return db
}

func getEvents(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	month := r.URL.Query().Get("month")

	if date != "" {
		if !isValidDate(date) {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
		getDayViewEvents(w, date)
	} else if month != "" {
		if !isValidMonth(month) {
			http.Error(w, "Invalid month format", http.StatusBadRequest)
			return
		}
		getMonthViewEvents(w, month)
	} else {
		getDayViewEvents(w, currentDate())
	}
	
}


func getMonthViewEvents(w http.ResponseWriter, month string) {
	rows, err := db.Query("SELECT id, title, DATE_FORMAT(date, '%Y-%m-%d') FROM events WHERE date LIKE ?", month+"%")
	if err != nil {
		http.Error(w, "Error fetching month events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Date); err != nil {
			http.Error(w, "Error reading event data", http.StatusInternalServerError)
			return
		}
		events = append(events, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// Day view: Full details (sections, notifications)
func getDayViewEvents(w http.ResponseWriter, date string) {
	rows, err := db.Query("SELECT id, title, DATE_FORMAT(date, '%Y-%m-%d'), TIME_FORMAT(time, '%H:%i'), description, noti FROM events WHERE date = ?", date)
	if err != nil {
		http.Error(w, "Error fetching day events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Date, &e.Time, &e.Description, &e.Noti); err != nil {
			http.Error(w, "Error reading event data", http.StatusInternalServerError)
			return
		}
		events = append(events, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}


func createEvent(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var e Event
	if err := json.Unmarshal(body, &e); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO events (title, date, time, description, noti) VALUES (?, ?, ?, ?, ?)",
		e.Title, e.Date, e.Time, e.Description, e.Noti)
	if err != nil {
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Event created successfully")
}


func updateEvent(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var e Event
	if err := json.Unmarshal(body, &e); err != nil || e.ID == 0 {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE events SET title=?, date=?, time=?, description=?, noti=? WHERE id=?",
		e.Title, e.Date, e.Time, e.Description, e.Noti, e.ID)
	if err != nil {
		http.Error(w, "Failed to update event", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Event updated successfully")
}

func deleteEvent(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing event ID", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM events WHERE id=?", id)
	if err != nil {
		http.Error(w, "Failed to delete event", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Event deleted successfully")
}

func main() {
	var db *sql.DB
	var err error
	db := initDB()
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if runtime.GOOS == "windows" {
			http.Error(w, "This route isn't supported on Windows. Still Panicking", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintln(w, "The Zone API is running 🚀")
	})
	healthOutput, error := php.Exec("health.php")
	if error != nil {
		log.Fatal(error)
	}
	fmt.Fprintf(response, healthOutput)
	
	http.HandleFunc("/events", getEvents)     
	http.HandleFunc("/event/create", createEvent)
	http.HandleFunc("/event/update", updateEvent)
	http.HandleFunc("/event/delete", deleteEvent)

	log.Println("Server is running on port 8089...")
	log.Fatal(http.ListenAndServe(":8089", nil))
}
