package validation

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseIntParam parses a string parameter to int
func ParseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	if val, err := strconv.Atoi(param); err == nil {
		return val
	}
	return defaultValue
}

// ValidateLimit ensures pagination limit is within bounds
func ValidateLimit(limit int) int {
	if limit <= 0 {
		return 50 // default
	}
	if limit > 100 {
		return 100 // max
	}
	return limit
}

// ValidateOffset ensures offset is not negative
func ValidateOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

// ValidatePosition checks if position is valid NFL position
func ValidatePosition(position string) error {
	validPositions := []string{
		"QB", "RB", "WR", "TE", "OL", "OT", "OG", "C",
		"DL", "DE", "DT", "LB", "CB", "S", "K", "P",
	}

	position = strings.ToUpper(strings.TrimSpace(position))

	for _, valid := range validPositions {
		if position == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid position: %s", position)
}

// ValidateStatus checks if player status is valid
func ValidateStatus(status string) error {
	validStatuses := []string{"active", "injured", "inactive"}

	status = strings.ToLower(strings.TrimSpace(status))

	for _, valid := range validStatuses {
		if status == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid status: %s", status)
}