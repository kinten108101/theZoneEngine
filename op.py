from datetime import datetime
import optuna

class Dynamax():
    def __init__(self, title, start_time, end_time, duration):
        self.title = title
        self.start_time = start_time
        self.end_time = end_time
        self.list = []
        self.duration = duration
    def __str__(self):
        return f"Title: {self.title}, Start: {self.start_time.strftime('%H:%M')}, End: {self.end_time.strftime('%H:%M')}"
    
class Event():
    def __init__(self, title, start_time, end_time, duration = 0, Static = None):
        self.title = title
        self.start_time = start_time
        self.end_time = end_time
        self.Static = None
        self.duration = duration if duration else end_time - start_time

    def __str__(self):
        return f"Title: {self.title}, Start: {self.start_time.strftime('%H:%M')}, End: {self.end_time.strftime('%H:%M')}, Duration: {self.duration}"

sched = []
sleep_du = 8
sleep_time = 0
latest_sleep = 0

def get_free_slots(start_time, end_time):
    start_time = parse_time(normalize_time_input(start_time))
    end_time = parse_time(normalize_time_input(end_time))
    free_slots = []
    current_time = start_time

    for event in sorted(sched, key=lambda e: e.start_time):
        if event.start_time >= end_time:
            break  
        if event.start_time > current_time:
            free_slots.append((current_time.strftime('%H:%M'), event.start_time.strftime('%H:%M')))

        if (event.end_time > current_time):
            current_time = event.end_time

    if current_time < end_time:
        free_slots.append((current_time.strftime('%H:%M'), end_time.strftime('%H:%M')))

    return free_slots


def parse_time(t):
    return datetime.strptime(t, "%H:%M")

def is_overlapping(new_event):
    for event in sched:
        if event.Static and not (new_event.end_time <= event.start_time or new_event.start_time >= event.end_time):
            return True
    return False

def add_event(e):
    if isinstance(e, Event):
        if is_overlapping(e):
            response = input("Warning: The event has an overlapping timeline with an existing static event.\nDo you want to proceed? (yes/no): ")
            if response.lower()[0] != 'y':
                print("Event was not added.")
                return
        sched.append(e)
        print("Static Event added successfully.")
    else:
        print("We need to do something")
        # todo

def normalize_time_input(time_input):
    if len(time_input) == 1:
        time_input ='0'+ time_input
    time_str = time_input.replace(':', '').ljust(4, '0')
    print(time_str)
    if len(time_str) == 4 and time_str.isdigit():
        return f"{time_str[:2]}:{time_str[2:]}"
    else:
        raise ValueError("Invalid time format. Please enter time as 'HHMM' or 'HH:MM'.")


def main():
    sleep_time = parse_time(normalize_time_input(input("Your preferred sleeping time: ")))
    latest_sleep = parse_time(normalize_time_input(input("Latest sleeping time")))

    while True:
        print("Menu")
        print("1. Add events")
        print("2. exit")
        choice = input("Your choice:  ")
        if choice == '2':
            break
        title = input("Enter Event name: ")
        start_input = input("Enter start time (e.g., '11', '1130', '11:30'): ")
        end_input = input("Enter end time (e.g., '13', '1300', '13:00'): ")
        is_static_input = input("Is this a static event? (y/n): ").lower()
        is_static = (is_static_input[0].lower() != 'n')

        try:
            start_time = parse_time(normalize_time_input(start_input))
            end_time = parse_time(normalize_time_input(end_input))
            if start_time >= end_time:
                print("Error: Start time must be before end time.")
                continue
        except ValueError as ve:
            print(f"Error: {ve}")
            continue

        event = Event(title, start_time, end_time) if is_static else Dynamax(title, start_time, end_time)
        add_event(event)

    print("\nFinal Schedule:")
    for event in sched:
        print(event)
    print("\nFree slot:")
    print(get_free_slots("00:00", "23:59"))
    

if __name__ == "__main__":
    main()

