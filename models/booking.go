package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Booking represents a single moving job booking.
type Booking struct {
	BookingID     string    `json:"user_id" db:"user_id"`
	CustomerName  string    `json:"customer_name" db:"customer_name"`
	Email         string    `json:"email" db:"email"`
	Phone         string    `json:"phone" db:"phone"`
	PickupAddress string    `json:"pickup_address" db:"pickup_address"`
	DropAddress   string    `json:"drop_address" db:"drop_address"`
	PickupDate    time.Time `json:"pickup_date" db:"pickup_date"`
	Status        int       `json:"status" db:"status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

func parseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil { // handles "2006-01-02T15:04:05.999Z" (JS toISOString output)
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil { // handles "2006-01-02T15:04:05Z07:00"
		return t, nil
	}
	return time.Parse("2006-01-02T15:04:05", s) // fallback for timestamps without timezone info
}

// UnmarshalJSON implements custom JSON unmarshaling for Booking.
// It handles flexible date formats and phone/status fields that may arrive as strings or numbers.
func (b *Booking) UnmarshalJSON(data []byte) error {
	type Alias Booking
	aux := &struct {
		PickupDate string      `json:"pickup_date"`
		CreatedAt  string      `json:"created_at"`
		UpdatedAt  string      `json:"updated_at"`
		Phone      interface{} `json:"phone"`
		Status     interface{} `json:"status"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.PickupDate != "" {
		t, err := parseTime(aux.PickupDate)
		if err != nil {
			return err
		}
		b.PickupDate = t
	}
	if aux.CreatedAt != "" {
		t, err := parseTime(aux.CreatedAt)
		if err != nil {
			return err
		}
		b.CreatedAt = t
	}
	if aux.UpdatedAt != "" {
		t, err := parseTime(aux.UpdatedAt)
		if err != nil {
			return err
		}
		b.UpdatedAt = t
	}

	// Handle phone that may arrive as a string, a JSON number (float64), or an int
	if aux.Phone != nil {
		switch v := aux.Phone.(type) {
		case string:
			b.Phone = v
		case float64:
			b.Phone = fmt.Sprintf("%.0f", v) // JSON numbers decode as float64; strip the decimal point
		case int:
			b.Phone = strconv.Itoa(v)
		}
	}

	// Handle status that may arrive as a named string label or an integer code
	if aux.Status != nil {
		switch v := aux.Status.(type) {
		case string:
			switch v {
			case "PENDING":
				b.Status = int(StatusPending)
			case "CONFIRMED":
				b.Status = int(StatusConfirmed)
			case "IN_PROGRESS":
				b.Status = int(StatusInProgress)
			case "COMPLETED":
				b.Status = int(StatusCompleted)
			case "CANCELLED":
				b.Status = int(StatusCancelled)
			default:
				b.Status = int(StatusPending) // unknown string values default to pending
			}
		case float64:
			b.Status = int(v) // JSON numbers always decode as float64, even when the value is an integer
		case int:
			b.Status = v
		}
	}

	return nil
}

// GetStatusString returns the human-readable label for the booking's status integer.
func (b *Booking) GetStatusString() string {
	switch b.Status {
	case int(StatusPending):
		return "PENDING"
	case int(StatusConfirmed):
		return "CONFIRMED"
	case int(StatusInProgress):
		return "IN_PROGRESS"
	case int(StatusCompleted):
		return "COMPLETED"
	case int(StatusCancelled):
		return "CANCELLED"
	default:
		return "PENDING"
	}
}

// FurnitureItem represents the furniture inventory for a single booking.
type FurnitureItem struct {
	FurnitureID     string `json:"furniture_id" db:"furniture_id"` // DB FK to booking_user.user_id
	Chairs          int    `json:"chairs" db:"chairs"`
	Table2Seats     int    `json:"table_2_seats" db:"table_2_seats"`
	Table3Seats     int    `json:"table_3_seats" db:"table_3_seats"`
	Table4PlusSeats int    `json:"table_4_plus_seats" db:"table_4_plus_seats"`
	Fridges         int    `json:"fridges" db:"fridges"`
	WashingMachines int    `json:"washing_machines" db:"washing_machines"`
	Dryers          int    `json:"dryers" db:"dryers"`
	Dishwashers     int    `json:"dishwashers" db:"dishwashers"`
	Boxes           int    `json:"boxes" db:"boxes"`
	PotPlants       int    `json:"pot_plants" db:"pot_plants"`
	Mattresses      int    `json:"mattresses" db:"mattresses"`
	BedFrames       int    `json:"bed_frames" db:"bed_frames"`
	Sofas           int    `json:"sofas" db:"sofas"`
}

// FurnitureEntry is a name-quantity pair for a single furniture category.
type FurnitureEntry struct {
	Name     string
	Quantity int
}

// ItemList returns only the furniture categories that have a non-zero quantity.
func (f *FurnitureItem) ItemList() []FurnitureEntry {
	all := []FurnitureEntry{ // define all categories in one place; order matches the original invoice output
		{"Chairs", f.Chairs},
		{"Table (2 seats)", f.Table2Seats},
		{"Table (3 seats)", f.Table3Seats},
		{"Table (4+ seats)", f.Table4PlusSeats},
		{"Fridges", f.Fridges},
		{"Washing Machines", f.WashingMachines},
		{"Dryers", f.Dryers},
		{"Dishwashers", f.Dishwashers},
		{"Boxes", f.Boxes},
		{"Pot Plants", f.PotPlants},
		{"Mattresses", f.Mattresses},
		{"Bed Frames", f.BedFrames},
		{"Sofas", f.Sofas},
	}

	var result []FurnitureEntry // fresh slice so we never modify the backing array of all
	for _, item := range all {
		if item.Quantity > 0 { // only include items that are actually being moved
			result = append(result, item)
		}
	}
	return result
}

// BookingStatus is the integer type used for booking state values.
type BookingStatus int

const (
	StatusPending    BookingStatus = 0
	StatusConfirmed  BookingStatus = 1
	StatusInProgress BookingStatus = 2
	StatusCompleted  BookingStatus = 3
	StatusCancelled  BookingStatus = 4
)

// NewBooking creates a new Booking with default values set.
func NewBooking() *Booking {
	return &Booking{
		BookingID: uuid.New().String(),
		Status:    int(StatusPending),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
