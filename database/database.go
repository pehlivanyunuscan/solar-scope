package database

import (
	"encoding/json"
	"fmt"
	"log"
	"solar-scope/internal/config"
	"solar-scope/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(cfg config.Config) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Istanbul",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established")

	// Modelleri otomatik olarak migrate et
	err = DB.AutoMigrate(
		&models.Forecast{},
		&models.EnergyBalance{},
		&models.BatteryPerformance{},
		&models.ActionRecommendation{},
	)
	if err != nil {
		log.Fatal("Migrate işlemi başarısız:", err)
	}

}

// SaveForecast, gelen payload'u veritabanına kaydeder
func SaveForecast(payload models.ForecastPayload) (*models.Forecast, error) {
	result := payload.Result
	// Zamanı string olarak al ve time.Time'a dönüştür
	parsedTime, err := time.Parse("2006-01-02T15:04:05.999999", payload.Timestamp)
	if err != nil {
		// Eğer zaman formatı hatalıysa, kaydı yapmadan hata döndür.
		return nil, fmt.Errorf("zaman formatı ayrıştırılamadı: %w", err)
	}

	recommendations := []models.ActionRecommendation{}
	for _, rec := range result.ActionRecommendations {
		recommendations = append(recommendations, models.ActionRecommendation{Recommendation: rec})
	}

	forecast := &models.Forecast{
		SessionID:     payload.SessionID,
		Timestamp:     parsedTime,
		ForecastDate:  result.Date,
		GeneralStatus: payload.GeneralStatus,
		EnergyBalance: models.EnergyBalance{
			TotalProductionKwh:  result.EnergyBalance.TotalProductionKwh,
			TotalConsumptionKwh: result.EnergyBalance.TotalConsumptionKwh,
			NetBatteryChangeWh:  result.EnergyBalance.NetBatteryChangeWh,
			StatusDescription:   result.EnergyBalance.StatusDescription,
		},
		BatteryPerformance: models.BatteryPerformance{
			InitialSoc:         result.BatteryPerformance.InitialSoc,
			MinSoc:             result.BatteryPerformance.MinSoc,
			MinSocTime:         result.BatteryPerformance.MinSocTime,
			MaxSoc:             result.BatteryPerformance.MaxSoc,
			MaxSocTime:         result.BatteryPerformance.MaxSocTime,
			EndOfDaySoc:        result.BatteryPerformance.EndOfDaySoc,
			TimeToFull:         result.BatteryPerformance.TimeToFull,
			FullChargeExpected: result.BatteryPerformance.FullChargeExpected,
		},
		ActionRecommendations: recommendations,
	}

	if err := DB.Create(forecast).Error; err != nil {
		return nil, err
	}
	return forecast, nil
}

func SaveResultToDB(result interface{}) {
	var dbPayload models.ForecastPayload

	resultBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshaling result: %v", err)
		return
	}

	err = json.Unmarshal(resultBytes, &dbPayload)
	if err != nil {
		log.Printf("Error unmarshaling to ForecastPayload: %v", err)
		return
	}

	if dbPayload.SessionID == "" {
		log.Println("No session_id in result, skipping DB save")
		return
	}

	_, err = SaveForecast(dbPayload)
	if err != nil {
		log.Printf("Error saving forecast to DB: %v", err)
		return
	}
	log.Println("Forecast saved to DB successfully")
}

// son tahminleri getirir
func GetRecentForecasts(limit int) ([]models.Forecast, error) {
	var forecasts []models.Forecast
	err := DB.Limit(limit).Order("timestamp desc").
		Preload("EnergyBalance").
		Preload("BatteryPerformance").
		Preload("ActionRecommendations").
		Find(&forecasts).Error
	return forecasts, err
}

// ID'ye göre belirli bir tahmini getirir
func GetForecastByID(id string) (*models.Forecast, error) {
	var forecast models.Forecast
	err := DB.Where("id = ?", id).
		Preload("EnergyBalance").
		Preload("BatteryPerformance").
		Preload("ActionRecommendations").
		First(&forecast).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Kayıt bulunamadı
		}
		return nil, err
	}
	return &forecast, nil
}
