package freshdesk

import "context"

// BusinessHours is a working-hours calendar. Calendars are read-only through
// the API; they are managed in the admin console.
//
// Note that this endpoint encodes its ID as a quoted string rather than a JSON
// number, which is why the field is an ID rather than an int64.
type BusinessHours struct {
	ID          ID     `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TimeZone    string `json:"time_zone"`
	IsDefault   bool   `json:"is_default"`
	// BusinessHours maps a lowercase weekday name to its working window. Days
	// that are not worked are absent.
	BusinessHours map[string]BusinessHoursWindow `json:"business_hours"`
	// SourceBusinessHours is the newer per-channel representation, which allows
	// several time slots per day.
	SourceBusinessHours []SourceBusinessHours `json:"source_business_hours"`
	// Holidays maps a holiday name to its date, where the account defines any.
	Holidays  map[string]string `json:"holidays"`
	CreatedAt Time              `json:"created_at"`
	UpdatedAt Time              `json:"updated_at"`
}

// BusinessHoursWindow is one day's working window, e.g. "08:00:00 am".
type BusinessHoursWindow struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// SourceBusinessHours is a channel-specific working-hours definition.
type SourceBusinessHours struct {
	// BusinessHoursType is "custom" or "24x7".
	BusinessHoursType string             `json:"business_hours_type"`
	BusinessHours     []BusinessHoursDay `json:"business_hours"`
	IsDefault         bool               `json:"is_default"`
	// Sources lists the channel IDs this definition applies to.
	Sources []int `json:"sources"`
}

// BusinessHoursDay is one day's set of working slots.
type BusinessHoursDay struct {
	Day       string             `json:"day"`
	TimeSlots []BusinessTimeSlot `json:"time_slots"`
}

// BusinessTimeSlot is a single working interval in 24-hour "HH:MM" form.
type BusinessTimeSlot struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// GetBusinessHours fetches a business-hours calendar by ID.
func (c *Client) GetBusinessHours(ctx context.Context, id int64) (*BusinessHours, error) {
	return getResource[BusinessHours](ctx, c, "business_hours", id, nil)
}

// ListBusinessHours returns every business-hours calendar.
func (c *Client) ListBusinessHours(ctx context.Context, opts ListOptions) ([]BusinessHours, error) {
	return listAll[BusinessHours](ctx, c, "business_hours", opts)
}
