import random
from datetime import datetime, timedelta
from Time import *
from helpfunc import *

class GeneticScheduler:
    def __init__(self,  sleep_time, latest_sleep, sleep_du, lunch_du):
        self.sched = []
        self.free_slots = []
        self.generations = 100
        self.population_size = 50
        self.mutation_rate = 0.1
        self.sleep_time = sleep_time
        self.latest_sleep = latest_sleep
        self.sleep_du = sleep_du
        self.lunch_du = lunch_du

    
    def get_free_slots(self, start_time, end_time):
        start_time = parse_time(normalize_time_input(start_time))
        end_time = parse_time(normalize_time_input(end_time))
        free_slots = []
        current_time = start_time

        for event in sorted(self.sched, key=lambda e: e.start_time):
            if event.start_time >= end_time:
                break  
            if event.start_time > current_time:
                free_slots.append((current_time.strftime('%H:%M'), event.start_time.strftime('%H:%M')))

            if event.end_time > current_time:
                current_time = event.end_time

        if current_time < end_time:
            free_slots.append((current_time.strftime('%H:%M'), end_time.strftime('%H:%M')))

        self.free_slots = free_slots
    
    def is_overlapping(self, new_event):
        for event in self.sched:
            if event.Dyna == None and not (new_event.end_time <= event.start_time or new_event.start_time >= event.end_time):
                return True
        return False
    
    def add_event(self, e):
        if isinstance(e, Event):
            if self.is_overlapping(e):
                response = input("Warning: The event has an overlapping timeline with an existing static event.\nDo you want to proceed? (yes/no): ")
                if response.lower()[0] != 'y':
                    print("Event was not added.")
                    return
            self.sched.append(e)
            print("Static Event added successfully.")
        elif isinstance(e, Dynamax):
            print("Invalid event.")
    
    def fitness(self, schedule):
        penalty = 0
        lunch_time = None
        sleep_time = None
        
        for event in reversed(schedule):
            if event.end_time:  
                sleep_time = event.end_time
                break 
        
        if lunch_time:
            penalty += self.lunch_penalty(self.lunch_du)[0]
        penalty += self.sleep_penalty(sleep_time) if sleep_time else 10000
        
        return -penalty 
    
    def lunch_penalty(self, lunch_du):
        lunch_window = (parse_time("10:30"), parse_time("13:30"))
        best_penalty = float("inf")
        best_lunch_start = None
        ideal = parse_time("12:00")
        for start, end in self.free_slots:
            start, end = parse_time(start), parse_time(end)
            
            if end <= lunch_window[0] or start >= lunch_window[1]:
                continue  

            start = max(start, lunch_window[0])
            end = min(end, lunch_window[1])

            slot_duration = (end - start).total_seconds() // 60
            lunch_start = max(start, min(ideal - timedelta(minutes=lunch_du // 2), end - timedelta(minutes=lunch_du)))

            center = lunch_start + timedelta(minutes=lunch_du // 2)
            center_penalty = abs((center - ideal).total_seconds() // 60)
            duration_penalty = abs(slot_duration - lunch_du)

            total_penalty = center_penalty + duration_penalty

            if total_penalty < best_penalty:
                best_penalty = total_penalty
                best_lunch_start = lunch_start
            

        return best_penalty, best_lunch_start.strftime("%H:%M") if best_lunch_start else None
    
    
    def sleep_penalty(self, sleep_time):
        if(sleep_time <= self.sleep_time):
                return 0 
        return ( min(self.latest_sleep - self.sleep_time,sleep_time - self.sleep_time)* 1.2 
                     + max(min(parse_time("01:00"),sleep_time - self.latest_sleep), 0)*2 
                     + max(sleep_time - parse_time("01:00") - self.latest_sleep,0))
    
    def generate_population(self):
        population = []
        for _ in range(self.population_size):
            schedule = []
            for slot in self.free_slots:
                start, end = slot
                duration = random.randint(30, 120)
                if start + timedelta(minutes=duration) <= end:
                    schedule.append(Event("Task", start, start + timedelta(minutes=duration)))
            population.append(schedule)
        return population
    
    def select_population(self, population, fitness_scores):
        sorted_population = sorted(zip(population, fitness_scores), key=lambda x: x[1], reverse=True)
        return [x[0] for x in sorted_population[:self.population_size // 2]]
    
    def crossover(self, parent1, parent2):
        split = len(parent1) // 2
        return parent1[:split] + parent2[split:]
    
    def mutate(self, schedule):
        if random.random() < self.mutation_rate:
            idx = random.randint(0, len(schedule) - 1)
            slot = random.choice(self.free_slots)
            new_start = slot[0]
            new_end = new_start + timedelta(minutes=random.randint(30, 120))
            schedule[idx] = Event("Task", new_start, new_end)
        return schedule
    
    def run(self):
        population = self.generate_population()
        
        for _ in range(self.generations):
            fitness_scores = [self.fitness(schedule) for schedule in population]
            population = self.select_population(population, fitness_scores)
            next_generation = []
            
            while len(next_generation) < self.population_size:
                parent1, parent2 = random.sample(population, 2)
                child = self.crossover(parent1, parent2)
                child = self.mutate(child)
                next_generation.append(child)
            
            population = next_generation
        
        best_schedule = max(population, key=self.fitness)
        return best_schedule
