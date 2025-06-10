package models

import (
	"time"

	"github.com/google/uuid"
)

// --- Core Entities ---
// for storing primary energy usage data
type EnergyUsage struct {
	ID                 uuid.UUID `db:"id" json:"id,omitempty"`
	FacilityID         string    `db:"facility_id" json:"facility_id" binding:"required"`
	Timestamp          time.Time `db:"timestamp" json:"timestamp" binding:"required"`
	EnergyKWh          float64   `db:"energy_kwh" json:"energy_kwh" binding:"required"`
	HumidityPercent    float64   `db:"humidity_percent" json:"humidity_percent" binding:"required"`
	TemperatureCelsius float64   `db:"temperature_celsius" json:"temperature_celsius" binding:"required"`
}

// data for storing ai results
type EnergyForecast struct {
	ID                string    `db:"id" json:"id"`
	FacilityID        string    `db:"facility_id" json:"facility_id"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	ForecastDate      time.Time `db:"forecast_date" json:"forecast_date"`
	PredictedKwh      float64   `db:"predicted_kwh" json:"predicted_kwh"`
	ConfidencePercent float64   `db:"confidence_percent" json:"confidence_percent"`
	LowerBoundKwh     float64   `db:"lower_bound_kwh" json:"lower_bound_kwh"`
	UpperBoundKwh     float64   `db:"upper_bound_kwh" json:"upper_bound_kwh"`
	HumidityPercent   float64   `db:"humidity_percent" json:"humidity_percent,omitempty"`
	ModelType         string    `db:"model_type" json:"model_type"`
}

// --- Request / Response Models ---

type StoreForecastRequest struct {
	FacilityID string              `json:"facility_id" binding:"required"`
	ModelType  string              `json:"model_type" binding:"required"`
	Forecasts  []ForecastDataPoint `json:"forecasts" binding:"required,min=1"`
}

type ForecastDataPoint struct {
	ForecastDate      string  `json:"forecast_date" binding:"required"`
	PredictedKwh      float64 `json:"predicted_kwh" binding:"required"`
	ConfidencePercent float64 `json:"confidence_percent" binding:"required"`
	HumidityPercent   float64 `json:"humidity_percent" binding:"required"`
	LowerBoundKwh     float64 `json:"lower_bound_kwh"`
	UpperBoundKwh     float64 `json:"upper_bound_kwh"`
}

// wrapper for forecast data response
type ForecastResponse struct {
	Status string             `json:"status"`
	Data   []ForecastDataItem `json:"data"`
}

// single forecast data item for response
type ForecastDataItem struct {
	ID                string  `json:"id"`
	FacilityID        string  `json:"facility_id"`
	ForecastDate      string  `json:"forecast_date"`
	PredictedKwh      float64 `json:"predicted_kwh"`
	ConfidencePercent float64 `json:"confidence_percent"`
	LowerBoundKwh     float64 `json:"lower_bound_kwh"`
	UpperBoundKwh     float64 `json:"upper_bound_kwh"`
	HumidityPercent   float64 `json:"humidity_percent"`
	CreatedAt         string  `json:"created_at"`
	ModelType         string  `json:"model_type"`
}
