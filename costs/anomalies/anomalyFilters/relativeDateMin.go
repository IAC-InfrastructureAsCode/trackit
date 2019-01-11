package anomalyFilters

type (
	// relativeDateMin will hide every entry before
	// today minus the given duration.
	//
	// Format (positive integer, in seconds):
	// 3600
	relativeDateMin struct{}
)

func init() {
	registerFilter("relative_date_min", relativeDateMin{})
}

// valid verifies the validity of the data
func (f relativeDateMin) valid(data interface{}) error {
	return genericValidUnsignedInteger(f, data)
}

// apply applies the filter to the anomaly results
func (f relativeDateMin) apply() {
}
