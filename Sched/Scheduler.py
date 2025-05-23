import random
from datetime import datetime, timedelta
from Time import *
from helpfunc import *
from copy import deepcopy
import sqlite3
import os
from dotenv import load_dotenv
load_dotenv()
from mysql.connector import connect, Error
from mysql.connector.cursor import MySQLCursorDict
import re

class Scheduler:
    def insert_events_into_db(self, events):
        cursor = self.db.cursor()

        insert_query = """
            INSERT INTO task (title, dt, start_time, end_time, des) 
            VALUES (%s, %s, %s, %s, %s)
        """



        for ev in events:
            cursor.execute(insert_query, (
                ev.title,
                ev.date,
                ev.start_time.strftime("%H:%M"),
                ev.end_time.strftime("%H:%M"),
                ev.description
            ))
            ev.id=cursor.lastrowid

        self.db.commit()
        cursor.close()

    

    def sync(self, start_date : str, end_date : str):
        start_dt = datetime.strptime(start_date, "%Y-%m-%d")
        end_dt = datetime.strptime(end_date, "%Y-%m-%d")
        days_range = [
            (start_dt + timedelta(days=i)).strftime("%Y-%m-%d")
            for i in range((end_dt - start_dt).days + 1)
        ]

        cursor = self.db.cursor(dictionary=True)  # Use dictionary cursor for MySQL
        for day in days_range:
            cursor.execute(
                "SELECT id, title, dt, start_time, end_time, des FROM task WHERE dt = %s", (day,)
            )
            rows = cursor.fetchall()

            for row in rows:
                self.create(
                    row["title"],
                    row["dt"],
                    row["start_time"],
                    row["end_time"],
                    des=row["des"],
                    is_static=True
                )

        cursor.close()
        return True
    
    def __init__(self, sleep_time, latest_sleep, wake, lunch_du):
        db_string = os.getenv("DBstring")
        if not db_string:
            raise ValueError("Environment variable 'DBstring' is not set")

        match = re.match(r'(\w+):([^@]+)@tcp\(([^:]+):(\d+)\)/(\w+)', db_string)
        if not match:
            raise ValueError("Invalid DBstring format")

        user, password, host, port, dbname = match.groups()

        self.db = connect(
            user=user,
            password=password,
            host=host,
            port=int(port),
            database=dbname
        )
        self.month = Month()
        self.free_slots = {}
        self.sleep_time = sleep_time
        self.latest_sleep = latest_sleep
        self.wake = wake
        self.lunch_du = lunch_du

    def add_event(self, e, d=None):
        if d==None:
            d = self.month.get_day(e.date)
        for existing in d.events:
            if not (e.end_time <= existing.start_time or e.start_time >= existing.end_time):
                print("Event is overlapped.")
                return False
        d.events.append(e)
        d.events.sort(key=lambda ev: ev.start_time)
        print("Event added.")
        return True

    def create(self, title, date, start_time, end_time,des = "", duration=None,end_date = None, is_static=True):
        events = None
        if is_static:
            events = [Event( title, date, start_time, end_time, des)]
            self.add_event(events[0])
        else:
            dyna = Dynamax(title, start_time, end_time, duration,des)
            success, events = self.autoCalendar(dyna, date, end_date)
            if success:
                print("Dynamic Event added.")
            else:
                print("Failed to schedule Dynamic Event.")
        return events

    def get_free_slots(self, date):
        d = self.month.get_day(date)
        events = sorted(d.events, key=lambda e: e.start_time)
        free = []

        start_time = self.wake  
        end_time = parse_time("23:59")
        cur = start_time

        for e in events:
            if e.start_time > cur:
                free.append((cur.strftime('%H:%M'), e.start_time.strftime('%H:%M')))
            cur = max(cur, e.end_time)

        if cur < end_time:
            free.append((cur.strftime('%H:%M'), end_time.strftime('%H:%M')))

        self.free_slots[date] = free

# _____________________________________________________________________To do section____________________________________________________________________
    def get_free_slots_from_events(self, events):
        events = sorted(events, key=lambda e: e.start_time)
        free = []
        cur = self.wake
        end_day = parse_time("23:59")

        for e in events:
            if e.start_time > cur:
                free.append((cur, e.start_time))
            cur = max(cur, e.end_time)

        if cur < end_day:
            free.append((cur, end_day))
        return free

    def sleep_penalty(self, sleep_time):
        if sleep_time <= self.sleep_time:
            return 0

        return int(
            min(self.latest_sleep - self.sleep_time, sleep_time - self.sleep_time).total_seconds() / 60 * 1.2 +
            max(min(timedelta(hours=1), sleep_time - self.latest_sleep), timedelta(hours=0)).total_seconds() / 60 * 2 +
            max(sleep_time - timedelta(hours=1) - self.latest_sleep, timedelta(hours=0)).total_seconds() / 60
        )

    def autoCalendar(self, dyna, start_date, end_date):
        start_dt = datetime.strptime(start_date, "%Y-%m-%d")
        end_dt = datetime.strptime(end_date, "%Y-%m-%d") if end_date else start_dt
        days_range = [(start_dt + timedelta(days=i)).strftime("%Y-%m-%d") 
                    for i in range((end_dt - start_dt).days + 1)]

        temp_days = {day: deepcopy(self.month.get_day(day).events) for day in days_range}
        best_sched = self.generate_sched(dyna, temp_days, days_range)

        for e in best_sched:
            self.month.get_day(e.date).events.append(e)
            self.month.get_day(e.date).events.sort(key=lambda x: x.start_time)
        print(best_sched)
        return True, best_sched

    def generate_sched(self, dyna, day_events, days_range):
        candidates = []
        max_attempts = 10

        for _ in range(max_attempts):
            temp_day_events = {day: day_events[day].copy() for day in day_events}
            sched = []
            dur_left = dyna.duration.total_seconds() // 60

            days_to_use = min(len(days_range), max(1, int(dur_left / 120)))
            selected_days = random.sample(days_range, min(days_to_use, len(days_range)))
            time_per_day = dur_left / len(selected_days)
            day_allocations = {day: 0 for day in selected_days}
            attempts = 0
            max_loop_attempts = 500

            while dur_left > 0 and attempts < max_loop_attempts:
                attempts += 1

                eligible_days = [day for day in selected_days if day_allocations[day] < time_per_day]
                if not eligible_days:
                    eligible_days = selected_days.copy()
                if not eligible_days:
                    break  # Prevent crash if all days are removed

                day = min(eligible_days, key=lambda d: day_allocations[d])
                available = self.get_free_slots_from_events(temp_day_events[day])
                available.sort(key=lambda slot: slot[0])

                # Skip days that have no 30+ min slots
                if not available or all((end - start).total_seconds() / 60 < 30 for start, end in available):
                    if day in selected_days and len(selected_days) > 1:
                        selected_days.remove(day)
                        time_per_day = dur_left / len(selected_days)
                    continue


                # Check for gaps and split events if needed
                for start, end in available:
                    slot_duration = (end - start).total_seconds() / 60
                    if slot_duration < 30:
                        continue

                    ideal_block = min(random.choice([45, 60, 75, 90, 105, 120]), dur_left)
                    block_size = min(ideal_block, slot_duration -30)
                    if block_size <= 0:
                        continue

                    block_td = timedelta(minutes=block_size)
                    add = min((slot_duration -block_size) / 2 ,0) 
                    add = timedelta(minutes=add)
                    end_time = start + block_td+add

                    # Calculate the perfect middle of the slot
  

                    # Create event with adjusted start time to be in the middle of the available slot
                    e = Event(dyna.title, day, start+add, end_time,"",dyna)
                    sched.append(e)
                    

                    dur_left -= block_size
                    temp_day_events[day].append(e)
                    temp_day_events[day].sort(key=lambda x: x.start_time)
                    day_allocations[day] += block_size

                    if dur_left <= 0:
                        break

                if dur_left <= 0:
                    break

            original_duration = dyna.duration.total_seconds() // 60
            if sched and (original_duration - dur_left) >= (original_duration * 0.75):
                candidates.append(sched)

        if not candidates:
            return []

        return max(candidates, key=lambda s: self.fitness(s))


    def fitness(self, schedule):
        if not schedule:
            return float("-inf")
            
        base_penalty = 0
        day_map = {}
        for e in schedule:
            day_map.setdefault(e.date, []).append(e)

        for date, events in day_map.items():
            all_day_events = self.month.get_day(date).events + [e for e in events if e not in self.month.get_day(date).events]
            free = self.get_free_slots_from_events(all_day_events)
            lunch_pen, _ = self.lunch_penalty_dynamic(free, self.lunch_du)
            base_penalty += lunch_pen
            
            if events:
                last_event = max(events, key=lambda e: e.end_time)
                base_penalty += self.sleep_penalty(last_event.end_time)
        
        # Distribution fitness components
        dist_score = 100
        days_used = len(day_map)
        dist_score += days_used * 20
        
        # Check distribution within each day
        for day, events in day_map.items():
            events.sort(key=lambda e: e.start_time)
            
            # Check for clumping
            for i in range(1, len(events)):
                time_gap = (events[i].start_time - events[i-1].end_time).total_seconds() / 60
                if time_gap < 30:
                    dist_score -= 15  # Penalty for events less than 30 minutes apart
                elif time_gap < 60:
                    dist_score -= 5   # Smaller penalty for events less than 1 hour apart
                else:
                    dist_score += 10  # Bonus for well-spaced events
            
            # Check time variety
            morning = any(e.start_time.hour < 12 for e in events)
            afternoon = any(12 <= e.start_time.hour < 17 for e in events)
            evening = any(e.start_time.hour >= 17 for e in events)
            time_variety_score = (morning + afternoon + evening) * 15
            dist_score += time_variety_score
        
        # Balance between days
        if days_used > 1:
            import statistics
            events_per_day = [len(events) for events in day_map.values()]
            std_dev = statistics.stdev(events_per_day) if len(events_per_day) > 1 else 0
            dist_score -= std_dev * 10
        
        # Combine base penalty with distribution score
        # Higher distribution score is better, lower penalty is better
        final_score = dist_score - base_penalty
        
        return final_score

    def lunch_penalty_dynamic(self, free_slots, lunch_du):
        lunch_window = (parse_time("10:30"), parse_time("13:30"))
        best_penalty = float("inf")
        best_lunch_start = None
        ideal = parse_time("12:00")

        for start, end in free_slots:
            if end <= lunch_window[0] or start >= lunch_window[1]:
                continue

            start = max(start, lunch_window[0])
            end = min(end, lunch_window[1])

            slot_duration = (end - start).total_seconds() // 60
            if slot_duration < lunch_du:
                continue

            lunch_start = max(start, min(ideal - timedelta(minutes=lunch_du // 2), end - timedelta(minutes=lunch_du)))
            center = lunch_start + timedelta(minutes=lunch_du // 2)

            center_penalty = abs((center - ideal).total_seconds() // 60)
            duration_penalty = abs(slot_duration - lunch_du)

            total_penalty = center_penalty + duration_penalty

            if total_penalty < best_penalty:
                best_penalty = total_penalty
                best_lunch_start = lunch_start

        return best_penalty if best_lunch_start else 1000, best_lunch_start.strftime("%H:%M") if best_lunch_start else None