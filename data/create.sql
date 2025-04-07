USE SCHED;
DROP TABLE days, months, years;
CREATE TABLE years (
    OBJECTIVE VARCHAR(255),
    Y YEAR,
    PRIMARY KEY (Y)
);
CREATE TABLE months (
	OBJECTIVE VARCHAR(255),
    M INT,
    Y YEAR,
    PRIMARY KEY(M,Y),
    FOREIGN KEY (Y) REFERENCES years(Y)
);
CREATE TABLE days (
	diary VARCHAR(255),
    dt DATE,
    M Int,
    Y YEAR,
    PRIMARY KEY(dt),
    FOREIGN KEY (M, Y) REFERENCES months(M, Y)
);
CREATE TABLE task (
	title VARCHAR(50),
    id INT AUTO_INCREMENT,
    start_time TIME,
    end_time Time,
    des VARCHAR(255),
    dt DATE,
    M Int,
    Y YEAR,
    FOREIGN KEY (dt) REFERENCES days(dt)
);