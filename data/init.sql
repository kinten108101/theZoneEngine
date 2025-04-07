use Sched;

CREATE TABLE years (
    Y INT PRIMARY KEY,        -- Year as primary key
    objective VARCHAR(255)    -- Objective for the year
);

-- Table for storing months, with a foreign key to years
CREATE TABLE months (
    M INT,                    -- Month as part of composite key
    Y INT,                    -- Year as part of composite key
    objective VARCHAR(255),   -- Objective for the month
    PRIMARY KEY (M, Y),       -- Composite key (Month, Year)
    FOREIGN KEY (Y) REFERENCES years(Y) -- Foreign key to the years table
);

-- Table for storing days, with a foreign key to months (composite key)
CREATE TABLE days (
    M INT,                    -- Month as part of composite key
    Y INT,                    -- Year as part of composite key
    dt DATE,                  -- Date for the day
    diary TEXT,               -- Diary entry for the day
    PRIMARY KEY (M, Y, dt),   -- Composite key (Month, Year, Date)
    FOREIGN KEY (M, Y) REFERENCES months(M, Y) -- Foreign key to the months table
);

-- Table for storing tasks, associated with specific dates, months, and years
CREATE TABLE task (
    id INT PRIMARY KEY AUTO_INCREMENT,   -- Task ID (primary key)
    title VARCHAR(255),                  -- Task title
    dt DATE,                             -- Date of the task
    start_time TIME,                     -- Start time of the task
    end_time TIME,                       -- End time of the task
    des TEXT,                            -- Task description
    M INT,                               -- Month part of foreign key
    Y INT,                               -- Year part of foreign key
    FOREIGN KEY (M, Y) REFERENCES months(M, Y) -- Foreign key to months
);
