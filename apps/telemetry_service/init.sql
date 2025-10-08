CREATE TABLE IF NOT EXISTS telemetry (
    id String,
    sensor_id String,
    device_id String,
    type String,
    unit String,
    value String,
    timestamp DateTime
) ENGINE = MergeTree()
ORDER BY (sensor_id, timestamp)