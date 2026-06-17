package service

import (
	"time"
)

// SubscriptionCacheData represents cached subscription data
type SubscriptionCacheData struct {
	Status             string
	ExpiresAt          time.Time
	DailyUsage         float64
	WeeklyUsage        float64
	MonthlyUsage       float64
	DailyWindowStart   *time.Time
	WeeklyWindowStart  *time.Time
	MonthlyWindowStart *time.Time
	Version            int64
}
