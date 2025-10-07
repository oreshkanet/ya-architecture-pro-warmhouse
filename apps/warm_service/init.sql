-- Create the database if it doesn't exist
CREATE DATABASE warm;

-- Connect to the database
\c warm;

CREATE TABLE IF NOT EXISTS warm_sensors (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    location TEXT NOT NULL,
    serial_number TEXT UNIQUE NOT NULL,
    is_on BOOLEAN NOT NULL DEFAULT false,
    current_temperature DOUBLE PRECISION NOT NULL DEFAULT 0,
    target_temperature DOUBLE PRECISION NOT NULL DEFAULT 20,
    status TEXT NOT NULL CHECK (status IN ('online', 'offline', 'error')),
    firmware_version TEXT,
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL,
    is_legacy BOOLEAN NOT NULL DEFAULT false,
    url TEXT NOT NULL
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_warm_sensors_location ON warm_sensors(location);
CREATE INDEX IF NOT EXISTS idx_warm_sensors_status ON warm_sensors(status);