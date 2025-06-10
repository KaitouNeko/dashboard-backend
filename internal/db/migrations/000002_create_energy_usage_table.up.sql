CREATE TABLE energy_usage (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  facility_id VARCHAR(50) NOT NULL,
  timestamp TIMESTAMP NOT NULL,
  energy_kwh FLOAT NOT NULL,
  temperature_celsius FLOAT NOT NULL
);

-- Create index for faster time-based queries
CREATE INDEX idx_energy_usage_timestamp ON energy_usage(timestamp);
-- Create index for facility-specific queries
CREATE INDEX idx_energy_usage_facility ON energy_usage(facility_id);
