package utils

import (
	"time"
)

// SeasonInfo contains information about the current NFL season
type SeasonInfo struct {
	Year         int
	CurrentWeek  int
	IsOffseason  bool
	IsPreseason  bool
	IsRegular    bool
	IsPostseason bool
}

// GetCurrentSeason returns information about the current NFL season
// NFL season runs from early September through early February
func GetCurrentSeason() SeasonInfo {
	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	info := SeasonInfo{}

	// Determine season year
	// NFL season year is based on when it starts (e.g., 2024 season runs Sep 2024 - Feb 2025)
	if month >= time.September {
		info.Year = year
	} else if month <= time.February {
		// January/February belongs to previous year's season
		info.Year = year - 1
	} else {
		// March-August is offseason, but return current year
		info.Year = year
		info.IsOffseason = true
		return info
	}

	// Determine week and phase of season
	// Regular season: ~18 weeks starting first week of September
	// Playoffs: 4 weeks in January
	// Super Bowl: Early February

	if month == time.September && day < 5 {
		// Preseason/early September
		info.IsPreseason = true
		info.CurrentWeek = 0
	} else if month >= time.September && month <= time.December {
		// Regular season (September - December)
		info.IsRegular = true
		info.CurrentWeek = calculateWeek(info.Year, now)
	} else if month == time.January {
		// Playoffs or end of regular season
		weekNum := calculateWeek(info.Year, now)
		if weekNum <= 18 {
			info.IsRegular = true
			info.CurrentWeek = weekNum
		} else {
			info.IsPostseason = true
			info.CurrentWeek = weekNum
		}
	} else if month == time.February && day <= 15 {
		// Super Bowl week
		info.IsPostseason = true
		info.CurrentWeek = 22 // Super Bowl is around week 22
	} else {
		// Offseason
		info.IsOffseason = true
		info.CurrentWeek = 0
	}

	return info
}

// calculateWeek calculates the current NFL week based on the season start date
func calculateWeek(seasonYear int, now time.Time) int {
	// NFL regular season typically starts the first Thursday after Labor Day
	// Labor Day is first Monday in September
	// For simplicity, assume season starts first week of September

	// Find the first Thursday in September of the season year
	seasonStart := time.Date(seasonYear, time.September, 1, 0, 0, 0, 0, time.UTC)

	// Find first Thursday
	for seasonStart.Weekday() != time.Thursday {
		seasonStart = seasonStart.AddDate(0, 0, 1)
	}

	// Calculate days since season start
	daysSinceStart := now.Sub(seasonStart).Hours() / 24

	// Calculate week (1-indexed)
	week := int(daysSinceStart/7) + 1

	if week < 1 {
		week = 1
	}
	if week > 22 {
		week = 22
	}

	return week
}

// GetSeasonWeek returns the season and week for a given date
func GetSeasonWeek(date time.Time) (season int, week int) {
	year := date.Year()
	month := date.Month()

	// Determine season year
	if month >= time.September {
		season = year
	} else if month <= time.February {
		season = year - 1
	} else {
		season = year
		week = 0
		return
	}

	week = calculateWeek(season, date)
	return
}

// IsSeasonActive checks if the NFL season is currently active
func IsSeasonActive() bool {
	info := GetCurrentSeason()
	return !info.IsOffseason
}

// GetAllWeeksForSeason returns all week numbers for a given season (1-18 regular season)
func GetAllWeeksForSeason(season int) []int {
	weeks := make([]int, 18)
	for i := 0; i < 18; i++ {
		weeks[i] = i + 1
	}
	return weeks
}

// ShouldFetchGames determines if we should auto-fetch games for a given season/week
func ShouldFetchGames(season, week int) bool {
	currentInfo := GetCurrentSeason()

	// Always fetch current season data
	if season == currentInfo.Year {
		return true
	}

	// Fetch previous season if we're in early season
	if season == currentInfo.Year-1 && currentInfo.CurrentWeek <= 5 {
		return true
	}

	// Don't auto-fetch historical data beyond 1 year
	if season < currentInfo.Year-1 {
		return false
	}

	// Validate week range
	if week < 1 || week > 18 {
		return false
	}

	return true
}
