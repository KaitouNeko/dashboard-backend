package energy

import (
	"ai-workshop/internal/models"
	"context"
	"errors"
	"fmt"
	"time"
)

type Service interface {
	CreateEnergyUsage(usage *models.EnergyUsage) error
	GetByDateRange(facilityID string, startDate, endDate time.Time) ([]models.EnergyUsage, error)
	GetByTemperatureRange(facilityID string, minTemp, maxTemp float64) ([]models.EnergyUsage, error)
	StoreForecastBatch(ctx context.Context, req *models.StoreForecastRequest) error
	GetForecasts(ctx context.Context, facilityID, startDateStr, endDatStr string) (*models.ForecastResponse, error)
}

type EnergyService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &EnergyService{
		repo: repo,
	}
}

func (s *EnergyService) CreateEnergyUsage(usage *models.EnergyUsage) error {
	if usage.FacilityID == "" {
		return errors.New("facility ID is required")
	}

	if usage.EnergyKWh <= 0 {
		return errors.New("energy usage must be positive")
	}

	return s.repo.CreateEnergyUsage(usage)
}

func (s *EnergyService) GetByDateRange(facilityID string, startDate, endDate time.Time) ([]models.EnergyUsage, error) {
	if facilityID == "" {
		return nil, errors.New("facility ID is required")
	}

	if startDate.After(endDate) {
		return nil, errors.New("start date must be before end date")
	}

	return s.repo.GetByFacilityAndDateRange(facilityID, startDate, endDate)
}

func (s *EnergyService) GetByTemperatureRange(facilityID string, minTemp, maxTemp float64) ([]models.EnergyUsage, error) {
	if facilityID == "" {
		return nil, errors.New("facility ID is required")
	}

	if minTemp > maxTemp {
		return nil, errors.New("minimum temperature must be less than or equal to maximum temperature")
	}

	return s.repo.GetByTemperatureRange(facilityID, minTemp, maxTemp)
}

/**
* Stores forceasts in a batch after AI has generated them.
**/
func (s *EnergyService) StoreForecastBatch(ctx context.Context, req *models.StoreForecastRequest) error {

	// convert request to repository model
	forecasts := make([]*models.EnergyForecast, 0, len(req.Forecasts))
	for _, f := range req.Forecasts {
		forecastDate, err := time.Parse("2006-01-02", f.ForecastDate)
		if err != nil {
			return err
		}

		fmt.Printf("\nhumidity percent when creating energy forecast: %f\n\n", f.HumidityPercent)

		forecast := &models.EnergyForecast{
			FacilityID:        req.FacilityID,
			ForecastDate:      forecastDate,
			PredictedKwh:      f.PredictedKwh,
			ConfidencePercent: f.ConfidencePercent,
			LowerBoundKwh:     f.LowerBoundKwh,
			UpperBoundKwh:     f.UpperBoundKwh,
			HumidityPercent:   f.HumidityPercent,
			ModelType:         req.ModelType,
		}
		forecasts = append(forecasts, forecast)
	}
	return s.repo.CreateForecastBatch(ctx, forecasts)
}

/**
* Gets the energy forecast data.
**/
func (s *EnergyService) GetForecasts(ctx context.Context, facilityID, startDateStr, endDateStr string) (*models.ForecastResponse, error) {
	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format, use YYYY-MM-DD: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format, use YYYY-MM-DD: %w", err)
	}

	// Get forecasts from repository
	forecasts, err := s.repo.GetForecasts(ctx, facilityID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	response := &models.ForecastResponse{
		Status: "success",
		Data:   make([]models.ForecastDataItem, 0, len(forecasts)),
	}

	for _, f := range forecasts {
		item := models.ForecastDataItem{
			ID:                f.ID,
			FacilityID:        f.FacilityID,
			ForecastDate:      f.ForecastDate.Format("2006-01-02"),
			PredictedKwh:      f.PredictedKwh,
			ConfidencePercent: f.ConfidencePercent,
			LowerBoundKwh:     f.LowerBoundKwh,
			UpperBoundKwh:     f.UpperBoundKwh,
			HumidityPercent:   f.HumidityPercent,
			CreatedAt:         f.CreatedAt.Format(time.RFC3339),
			ModelType:         f.ModelType,
		}
		response.Data = append(response.Data, item)
	}

	return response, nil
}
