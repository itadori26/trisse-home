-- Daily usage per file system for the disk forecast (#108), 90 days.
CREATE TABLE system_disk_daily (
    mount     TEXT NOT NULL,
    day       TEXT NOT NULL,
    min_used  INTEGER NOT NULL,
    max_used  INTEGER NOT NULL,
    last_used INTEGER NOT NULL,
    total     INTEGER NOT NULL,
    PRIMARY KEY (mount, day)
);

CREATE TABLE system_forecast_settings (
    id          INTEGER PRIMARY KEY CHECK (id = 1),
    window_days INTEGER NOT NULL,
    alert_days  INTEGER NOT NULL
);
