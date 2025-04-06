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
    def __init__(self, title, start_time, end_time, Dyna = None):
        self.title = title
        self.start_time = start_time
        self.end_time = end_time
        self.Dyna = Dyna
        self.duration = end_time - start_time

    def __str__(self):
        return f"Title: {self.title}, Start: {self.start_time.strftime('%H:%M')}, End: {self.end_time.strftime('%H:%M')}, Duration: {self.duration}"