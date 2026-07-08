package postgres

import "gorm.io/gorm"

type UserRepository struct{ db *gorm.DB }
type RegionRepository struct{ db *gorm.DB }
type StationRepository struct{ db *gorm.DB }
type AnalyticsRepository struct{ db *gorm.DB }
type AlertRepository struct{ db *gorm.DB }
type BackfillRepository struct{ db *gorm.DB }
type ActualWeatherRepository struct{ db *gorm.DB }
