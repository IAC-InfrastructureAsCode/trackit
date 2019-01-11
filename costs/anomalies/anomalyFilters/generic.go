package anomalyFilters

import (
	"fmt"
	"time"
)

type (
	// genericFilter implements valid called to validate
	// the data received by postAnomaliesFilters and apply
	// to apply the filter to anomaly results.
	// All filters have to implement genericFilter.
	genericFilter interface {
		valid(data interface{}) error
		apply()
	}
)

var (
	filters     = make(map[string]genericFilter)
	filtersName = make(map[genericFilter]string)
)

// registerFilter has to be called by every filters to register them.
func registerFilter(filterName string, filter genericFilter) {
	filters[filterName] = filter
	filtersName[filter] = filterName
}

// genericValidDate is a generic validation function to validate a date.
func genericValidDate(filter genericFilter, data interface{}) error {
	if typed, ok := data.(string); !ok {
		return fmt.Errorf("%s: not a string", filtersName[filter])
	} else if _, err := time.Parse("2006-01-02T15:04:05.000Z", typed); err != nil {
		return fmt.Errorf("%s: not a date", filtersName[filter])
	}
	return nil
}

// genericValidUnsignedInteger is a generic validation function to validate
// an unsigned integer.
func genericValidUnsignedInteger(filter genericFilter, data interface{}) error {
	if typed, ok := data.(float64); !ok {
		return fmt.Errorf("%s: not a number", filtersName[filter])
	} else if typed < 0 {
		return fmt.Errorf("%s: not a positive number", filtersName[filter])
	} else if typed != float64(int64(typed)) {
		return fmt.Errorf("%s: not an integer", filtersName[filter])
	}
	return nil
}

// genericValidUnsignedIntegerArray is a generic validation function to
// validate an array of positive integer.
func genericValidUnsignedIntegerArray(filter genericFilter, data interface{}, maxBound int) error {
	if typed, ok := data.([]interface{}); !ok {
		return fmt.Errorf("%s: not an array", filtersName[filter])
	} else if len(typed) == 0 {
		return fmt.Errorf("%s: empty array", filtersName[filter])
	} else {
		for i := range typed {
			if elemTyped, ok := typed[i].(float64); !ok {
				return fmt.Errorf("%s: not an array of number", filtersName[filter])
			} else if elemTyped < 0 || elemTyped > float64(maxBound) {
				return fmt.Errorf("%s: not an array of number between 0 and %d", filtersName[filter], maxBound)
			} else if elemTyped != float64(int64(elemTyped)) {
				return fmt.Errorf("%s: not an array of integer", filtersName[filter])
			}
		}
	}
	return nil
}

// Valid verifies the given couple filter / data.
func Valid(filterName string, data interface{}) error {
	if filter, ok := filters[filterName]; !ok {
		return fmt.Errorf("%s: rule not found", filterName)
	} else {
		return filter.valid(data)
	}
}
