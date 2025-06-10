package energy

import (
	"ai-workshop/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateEnergyUsage handles the creation of a new energy usage record
func (h *Handler) CreateEnergyUsage(c *gin.Context) {

	var energyUsage *models.EnergyUsage

	if err := c.ShouldBindJSON(&energyUsage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	if err := h.service.CreateEnergyUsage(energyUsage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create energy usage record: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Energy usage record created successfully",
	})
}

// GetByDateRange handles retrieving energy usage by date range
func (h *Handler) GetByDateRange(c *gin.Context) {
	var request struct {
		FacilityID string `json:"facility_id" binding:"required"`
		StartDate  string `json:"start_date" binding:"required"`
		EndDate    string `json:"end_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	startDate, err := time.Parse(time.RFC3339, request.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Use RFC3339 format (e.g., 2025-03-01T09:00:00Z)"})
		return
	}

	endDate, err := time.Parse(time.RFC3339, request.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Use RFC3339 format (e.g., 2025-03-01T17:00:00Z)"})
		return
	}

	usages, err := h.service.GetByDateRange(request.FacilityID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve energy usage data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": usages,
	})
}

// GetByTemperatureRange handles retrieving energy usage by temperature range
func (h *Handler) GetByTemperatureRange(c *gin.Context) {
	facilityID := c.Query("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "facility_id is required"})
		return
	}

	minTempStr := c.Query("min_temp")
	maxTempStr := c.Query("max_temp")

	if minTempStr == "" || maxTempStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_temp and max_temp are required"})
		return
	}

	minTemp, err := strconv.ParseFloat(minTempStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_temp format"})
		return
	}

	maxTemp, err := strconv.ParseFloat(maxTempStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max_temp format"})
		return
	}

	usages, err := h.service.GetByTemperatureRange(facilityID, minTemp, maxTemp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve energy usage data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": usages,
	})
}

func (h *Handler) StoreForecast(c *gin.Context) {
	var req *models.StoreForecastRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.StoreForecastBatch(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store forecasts: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Forecasts stored successfully",
	})
}

func (h *Handler) GetForecasts(c *gin.Context) {
	facilityID := c.Query("facility_id")
	if facilityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "facility_id is required"})
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	forecasts, err := h.service.GetForecasts(c.Request.Context(), facilityID, startDateStr, endDateStr)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve forecasts: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   forecasts,
	})
}
