package energy

import (
	"ai-workshop/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	CreateEnergyUsage(usage *models.EnergyUsage) error
	GetByFacilityAndDateRange(facilityID string, startDate, endDate time.Time) ([]models.EnergyUsage, error)
	GetByTemperatureRange(facilityID string, minTemp, maxTemp float64) ([]models.EnergyUsage, error)
	CreateForecastBatch(ctx context.Context, forecasts []*models.EnergyForecast) error
	GetForecasts(ctx context.Context, facilityID string, startDate, endDate time.Time) ([]*models.EnergyForecast, error)
}

type PostgresRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) CreateEnergyUsage(usage *models.EnergyUsage) error {
	query := `
		INSERT INTO energy_usage 
		(facility_id, timestamp, energy_kwh, humidity_percent, temperature_celsius)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	fmt.Printf("Storing energy usage with humidity % with %v", usage.HumidityPercent)

	return r.db.QueryRow(
		query,
		usage.FacilityID,
		usage.Timestamp,
		usage.EnergyKWh,
		usage.HumidityPercent,
		usage.TemperatureCelsius,
	).Scan(&usage.ID)
}

func (r *PostgresRepository) GetByFacilityAndDateRange(facilityID string, startDate, endDate time.Time) ([]models.EnergyUsage, error) {
	query := `
		SELECT * FROM energy_usage
		WHERE facility_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp ASC
	`

	var usages []models.EnergyUsage
	err := r.db.Select(&usages, query, facilityID, startDate, endDate)
	return usages, err
}

func (r *PostgresRepository) GetByTemperatureRange(facilityID string, minTemp, maxTemp float64) ([]models.EnergyUsage, error) {
	query := `
		SELECT * FROM energy_usage
		WHERE facility_id = $1 AND temperature_celsius BETWEEN $2 AND $3
		ORDER BY temperature_celsius ASC
	`

	var usages []models.EnergyUsage
	err := r.db.Select(&usages, query, facilityID, minTemp, maxTemp)
	return usages, err
}

func (r *PostgresRepository) CreateForecastBatch(ctx context.Context, forecasts []*models.EnergyForecast) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO energy_forecasts (
			facility_id, forecast_date, predicted_kwh, 
			confidence_percent, lower_bound_kwh, upper_bound_kwh, humidity_percent, model_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (facility_id, forecast_date, model_type) 
		DO UPDATE SET 
			predicted_kwh = EXCLUDED.predicted_kwh,
			confidence_percent = EXCLUDED.confidence_percent,
			lower_bound_kwh = EXCLUDED.lower_bound_kwh,
			upper_bound_kwh = EXCLUDED.upper_bound_kwh
		RETURNING id, created_at
	`)

	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, forecast := range forecasts {
		var id string
		var createdAt time.Time
		err := stmt.QueryRowContext(
			ctx,
			forecast.FacilityID, forecast.ForecastDate, forecast.PredictedKwh,
			forecast.ConfidencePercent, forecast.LowerBoundKwh, forecast.UpperBoundKwh,
			forecast.HumidityPercent,
			forecast.ModelType,
		).Scan(&id, &createdAt)

		if err != nil {
			return err
		}

		forecast.ID = id
		forecast.CreatedAt = createdAt
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetForecasts(ctx context.Context, facilityID string, startDate, endDate time.Time) ([]*models.EnergyForecast, error) {
	query := `
		SELECT 
			id, facility_id, created_at, forecast_date, 
			predicted_kwh, confidence_percent, lower_bound_kwh, upper_bound_kwh, humidity_percent, model_type
		FROM energy_forecasts
		WHERE facility_id = $1 AND forecast_date BETWEEN $2 AND $3
		ORDER BY forecast_date ASC
	`

	var forecasts []*models.EnergyForecast
	err := r.db.SelectContext(ctx, &forecasts, query, facilityID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return forecasts, nil
}
