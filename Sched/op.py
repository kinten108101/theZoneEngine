from datetime import datetime
from Scheduler import *
from Time import *
from helpfunc import *

def run_tc(sched):
    print("\nRunning test cases...\n")

    sched.create("Math Class", "2025-04-07", "08:00", "10:00")
    sched.create("Overlap Test", "2025-04-07", "13:30", "15:00")
    sched.create("Study AI", "2025-04-07", "08:00", "20:00",timedelta(hours=10, minutes=30),"2025-04-09" , is_static=False)

    sched.get_free_slots("2025-04-07")
    print("Free Slots on 2025-04-07:", sched.free_slots["2025-04-07"])

    for date, day in sched.month.days.items():
        print(f"\nDate: {date}")
        for event in day.events:
            print(" -", event)

def main():

    sleep_time = parse_time(normalize_time_input("9"))
    latest_sleep = parse_time(normalize_time_input("10"))
    wake = parse_time(normalize_time_input("7"))
    lunch_du = int(45)

    sched = Scheduler(sleep_time, latest_sleep, wake, lunch_du)
    sched.
    while True:
        print("\nMenu")
        print("1. Add event")
        print("2. Show free slots")
        print("3. Show all events")
        print("4. Run test cases")
        print("5. Exit")
        
        choice = input("Your choice: ")

        if choice == '5':
            break
        elif choice == '1':
            title = input("Enter event name: ")
            date = input("Enter date (YYYY-MM-DD): ")
            start_input = input("Enter start time (e.g., '11', '1130', '11:30'): ")
            end_input = input("Enter end time (e.g., '13', '1300', '13:00'): ")
            is_static_input = input("Is this a static event? (y/n): ").lower()
            is_static = (is_static_input[0] != 'n')

            try:
                start_time = normalize_time_input(start_input)
                end_time = normalize_time_input(end_input)
                sched.create(title, date, start_time, end_time, is_static=is_static)
            except ValueError as ve:
                print(f"Error: {ve}")
        elif choice == '2':
            date = input("Enter date to check free slots (YYYY-MM-DD): ")
            sched.get_free_slots(date)
            print("Free slots:", sched.free_slots.get(date, []))
        elif choice == '3':
            for date, day in sched.month.days.items():
                print(f"\nDate: {date}")
                for event in day.events:
                    print(" -", event)
        elif choice == '4':
            run_tc(sched)
        else:
            print("Invalid choice.")

    print("\nFinished.")

if __name__ == "__main__":
    main()