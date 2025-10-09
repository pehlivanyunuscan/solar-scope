package models

import (
	"time"

	"gorm.io/gorm"
)

// Forecast ana tabloyu temsil eder. Bütün ilişkiler bu model üzerinden kurulur.
type Forecast struct {
	gorm.Model
	SessionID             string    `json:"session_id"`
	Timestamp             time.Time `json:"timestamp"`
	ForecastDate          string    `json:"date"`
	GeneralStatus         string    `json:"general_status"`
	EnergyBalance         EnergyBalance
	BatteryPerformance    BatteryPerformance
	ActionRecommendations []ActionRecommendation `gorm:"constraint:OnDelete:CASCADE;"`
}

type EnergyBalance struct {
	gorm.Model
	ForecastID          uint    `json:"-"`
	TotalProductionKwh  float64 `json:"total_production_kwh"`
	TotalConsumptionKwh float64 `json:"total_consumption_kwh"`
	NetBatteryChangeWh  float64 `json:"net_battery_change_wh"`
	StatusDescription   string  `json:"status_description"`
}

type BatteryPerformance struct {
	gorm.Model
	ForecastID         uint    `json:"-"`
	InitialSoc         float64 `json:"initial_soc"`
	MinSoc             float64 `json:"min_soc"`
	MinSocTime         string  `json:"min_soc_time"`
	MaxSoc             float64 `json:"max_soc"`
	MaxSocTime         string  `json:"max_soc_time"`
	EndOfDaySoc        float64 `json:"end_of_day_soc"`
	TimeToFull         string  `json:"time_to_full"`
	FullChargeExpected bool    `json:"full_charge_expected"`
}

type ActionRecommendation struct {
	gorm.Model
	ForecastID     uint   `json:"-"`
	Recommendation string `json:"recommendation"`
}

// Gelen JSON verisini parse etmek için yardımcı bir struct
type ForecastPayload struct {
	Result struct {
		ActionRecommendations []string `json:"action_recommendations"`
		BatteryPerformance    struct {
			EndOfDaySoc        float64 `json:"end_of_day_soc"`
			FullChargeExpected bool    `json:"full_charge_expected"`
			InitialSoc         float64 `json:"initial_soc"`
			MaxSoc             float64 `json:"max_soc"`
			MaxSocTime         string  `json:"max_soc_time"`
			MinSoc             float64 `json:"min_soc"`
			MinSocTime         string  `json:"min_soc_time"`
			TimeToFull         string  `json:"time_to_full"`
		} `json:"battery_performance"`
		Date          string `json:"date"`
		EnergyBalance struct {
			NetBatteryChangeWh  float64 `json:"net_battery_change_wh"`
			StatusDescription   string  `json:"status_description"`
			TotalConsumptionKwh float64 `json:"total_consumption_kwh"`
			TotalProductionKwh  float64 `json:"total_production_kwh"`
		} `json:"energy_balance"`
	} `json:"result"`

	GeneralStatus string `json:"general_status"`
	SessionID     string `json:"session_id"`
	Timestamp     string `json:"timestamp"`
}
