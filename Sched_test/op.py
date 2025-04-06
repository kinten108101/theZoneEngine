from datetime import datetime
from Scheduler import *

sched = []
sleep_du = 8
sleep_time = 0
latest_sleep = 0
lunch_du = 1



def main():
    sleep_time = parse_time(normalize_time_input(input("Your preferred sleeping time: ")))
    latest_sleep = parse_time(normalize_time_input(input("Latest sleeping time: ")))
    sleep_du = input("How long do you want to sleep: ")
    lunch_du = input("How long do you want to break during lunch time: ")
    sched = Scheduler(sleep_time,latest_sleep,sleep_du,lunch_du)
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
    print("Sleep_time: ", sleep_time)
    print("Latest Sleep_time: ", latest_sleep)
    print("Sleep Duration: ", sleep_du)
    print(":unch Duration: ", lunch_du)
    
    

if __name__ == "__main__":
    main()

