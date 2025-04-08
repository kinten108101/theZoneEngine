from datetime import datetime
from Scheduler import *
from Time import *
from helpfunc import *
import argparse
import json

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--data", type=str, required=True)
    args = parser.parse_args()

    try:
        event = json.loads(args.data)
    except json.JSONDecodeError:
        print(json.dumps({"error": "Invalid JSON input"}))

    # Setup scheduler
    sleep_time = parse_time(normalize_time_input("9"))
    latest_sleep = parse_time(normalize_time_input("10"))
    wake = parse_time(normalize_time_input("7"))
    lunch_du = 45
    sched = Scheduler(sleep_time, latest_sleep, wake, lunch_du)

    date = event["date"]
    sched.sync(date, date)

    title = event["title"]
    start = None
    end = None
    des = event.get("description", "")
    is_static = event["FK"]
    is_static = (is_static == 1)
    # def create(self, title, date, start_time, end_time,des = "", duration=None,end_date = None, is_static=True):
    end_date = None
    duration = None

    if not is_static:
        end_date = event["end_date"]
        duration = timedelta(hours=event["dur"])
    else:
        start = parse_time(normalize_time_input(event["start_time"]))
        end = parse_time(normalize_time_input(event["end_time"]))

    created_event = sched.create(
        title=title,
        date=date,
        end_date=end_date,
        duration=duration,
        start_time=start,
        end_time=end,
        des=des,
        is_static=is_static
    )

    sched.insert_events_into_db(created_event)

    for e in created_event:
        print(json.dumps(e.to_json()))

