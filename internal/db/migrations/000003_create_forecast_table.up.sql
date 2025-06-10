
CREATE TABLE energy_forecasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    facility_id VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    forecast_date TIMESTAMP NOT NULL,
    predicted_kwh FLOAT NOT NULL,
    confidence_percent FLOAT NOT NULL,
    lower_bound_kwh FLOAT,
    upper_bound_kwh FLOAT,
    model_type VARCHAR(50) NOT NULL,
    UNIQUE (facility_id, forecast_date, model_type)
);

CREATE INDEX idx_energy_forecasts_facility_date ON energy_forecasts (facility_id, forecast_date);
