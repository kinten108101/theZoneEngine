class Dynamax():
    def __init__(self, title, start_time, end_time, duration):
        self.title = title
        self.start_time = start_time
        self.end_time = end_time
        self.list = []
        self.duration = duration
        
    def __str__(self):
        return f"Title: {self.title}, Start: {self.start_time.strftime('%H:%M')}, End: {self.end_time.strftime('%H:%M')}"
    
class Event:
    def __init__(self, ID, title, date, start_time, end_time, description="", Dyna=None):
        self.ID = ID
        self.title = title
        self.date = date  # "YYYY-MM-DD"
        self.start_time = start_time
        self.end_time = end_time
        self.description = description
        self.Dyna = Dyna
        self.duration = end_time - start_time

    def __str__(self):
        return f"{self.title} ({self.start_time.strftime('%H:%M')} - {self.end_time.strftime('%H:%M')})"

class Day:
    def __init__(self, date):
        self.date = date  
        self.events = []
        self.diary = ""

class Month:
    def __init__(self, objective=""):
        self.days = {}
        self.objective = objective

    def get_day(self, date_str):
        if date_str not in self.days:
            self.days[date_str] = Day(date_str)
        return self.days[date_str]
