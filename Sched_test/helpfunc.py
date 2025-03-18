from datetime import datetime

def normalize_time_input(time_input):
    if len(time_input) == 1:
        time_input ='0'+ time_input
    time_str = time_input.replace(':', '').ljust(4, '0')
    if len(time_str) == 4 and time_str.isdigit():
        return f"{time_str[:2]}:{time_str[2:]}"
    else:
        raise ValueError("Invalid time format. Please enter time as 'HHMM' or 'HH:MM'.")
    


def parse_time(t):
    return datetime.strptime(t, "%H:%M")