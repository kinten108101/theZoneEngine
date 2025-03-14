package main

import (
	"fmt"
	"time"
)

// Event represents a scheduled event
type Event struct {
	Title      string
	Start_Time time.Time
	End_Time   time.Time
	Static     bool
}

var sched []Event = []Event{}

func parseTime(t string) time.Time {
	layout := "15:04" 
	parsedTime, _ := time.Parse(layout, t)
	return parsedTime
}

func isOverlapping(newEvent Event) bool {
	for _, event := range sched {
		if event.Static && !(newEvent.End_Time.Before(event.Start_Time) || newEvent.Start_Time.After(event.End_Time)) {
			return true
		}
	}
	return false
}

func sort(e Event) {
	if e.Static {
		if isOverlapping(e) {
			var response string
			fmt.Println("Warning: The event has an overlapping timeline with an existing static event.")
			fmt.Print("Do you want to proceed? (yes/no): ")
			fmt.Scanln(&response)

			if response != "yes" {
				fmt.Println("Event was not added.")
				return
			}
		}
		sched = append(sched, e)
		fmt.Println("Event added successfully.")
	}else{
		
	}
}

func main() {

	e1 := Event{
		Title:      "Work",
		Start_Time: parseTime("11:00"),
		End_Time:   parseTime("13:00"),
		Static:     true,
	}
	sort(e1)

	e2 := Event{
		Title:      "Meeting",
		Start_Time: parseTime("12:30"),
		End_Time:   parseTime("14:00"),
		Static:     true,
	}
	sort(e2)

	fmt.Println("\nFinal Schedule:")
	for _, event := range sched {
		fmt.Printf("Title: %s, Start: %s, End: %s, Static: %t\n",
			event.Title, event.Start_Time.Format("15:04"), event.End_Time.Format("15:04"), event.Static)
	}
}
