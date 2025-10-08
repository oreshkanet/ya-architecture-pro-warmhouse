-- Create the database if it doesn't exist
CREATE DATABASE devices;

-- Connect to the database
\c devices;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    house_id UUID NOT NULL,
    name TEXT NOT NULL,
    type TEXT CHECK (type IN ('room', 'outdoor', 'floor')),
    parent_location_id UUID REFERENCES locations(id)
);

CREATE INDEX IF NOT EXISTS idx_locations_user_id ON locations(user_id);
CREATE INDEX IF NOT EXISTS idx_locations_house_id ON locations(house_id);
CREATE INDEX IF NOT EXISTS idx_locations_parent_location_id ON locations(parent_location_id);

CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    name TEXT NOT NULL,
    device_type TEXT NOT NULL CHECK (device_type IN ('warm', 'light', 'door', 'video', 'universal')),
    serial_number TEXT,
    location_id UUID NOT NULL REFERENCES locations(id),
    status TEXT DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'disabled')),
    firmware_version TEXT,
    connected_at TIMESTAMPTZ,
    last_seen TIMESTAMPTZ
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_devices_user_id ON devices(user_id);
CREATE INDEX IF NOT EXISTS idx_devices_location_id ON devices(location_id);