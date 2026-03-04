package repository

import (
	"MovingCompanyGo/config"
	"MovingCompanyGo/models"
	"context"
	"encoding/json"
	"log" // P3: needed to log the cleanup error if the rollback delete also fails
	"time"

	supabase "github.com/supabase-community/supabase-go"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *models.Booking, furnitureItems *models.FurnitureItem) error
	GetByID(ctx context.Context, bookingID string) (*models.Booking, error)
	GetFurnitureItemsByBookingID(ctx context.Context, bookingID string) (*models.FurnitureItem, error)
	Update(ctx context.Context, booking *models.Booking) error
	Delete(ctx context.Context, bookingID string) error
	List(ctx context.Context) ([]*models.Booking, error)
}

type SupabaseBookingRepository struct {
	client *supabase.Client
}

func NewSupabaseBookingRepository() *SupabaseBookingRepository {
	return &SupabaseBookingRepository{
		client: config.SupabaseClient,
	}
}

func (r *SupabaseBookingRepository) Create(ctx context.Context, booking *models.Booking, furnitureItems *models.FurnitureItem) error {
	booking.CreatedAt = time.Now()
	booking.UpdatedAt = time.Now()

	// Insert the booking record into booking_user — DB table name and column names unchanged
	_, _, err := r.client.From("booking_user").Insert(booking, false, "", "", "").Execute()
	if err != nil {
		return err
	}

	// Link the furniture record to the booking using the booking's generated ID
	furnitureItems.FurnitureID = booking.BookingID // R5: field renamed from UserID to BookingID; json:"user_id" tag unchanged

	// Insert furniture items into booking_furniture_items — DB table name and column names unchanged
	_, _, err = r.client.From("booking_furniture_items").Insert(furnitureItems, false, "", "", "").Execute()
	if err != nil {
		// P3: The booking row was inserted but the furniture insert failed, which would leave an
		// orphaned booking record with no furniture data. Delete the booking to restore consistency.
		// Supabase Go client v0.0.4 has no transaction support, so manual cleanup is the best option.
		_, _, cleanupErr := r.client.From("booking_user").Delete("", "").Eq("user_id", booking.BookingID).Execute()
		if cleanupErr != nil {
			// P3: Log the cleanup failure so the operator knows a stale record needs manual removal
			log.Printf("Create: furniture insert failed (%v) and cleanup of booking %s also failed: %v", err, booking.BookingID, cleanupErr)
		}
		return err // return the original furniture error, not the cleanup error
	}

	return nil
}

func (r *SupabaseBookingRepository) GetByID(ctx context.Context, id string) (*models.Booking, error) {
	var booking models.Booking
	// "user_id" is the DB column name — must remain unchanged
	result, _, err := r.client.From("booking_user").Select("*", "", false).Eq("user_id", id).Single().Execute()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(result, &booking); err != nil {
		return nil, err
	}

	return &booking, nil
}

func (r *SupabaseBookingRepository) GetFurnitureItemsByBookingID(ctx context.Context, bookingID string) (*models.FurnitureItem, error) {
	var furnitureItem models.FurnitureItem
	// "furniture_id" is the DB column name — must remain unchanged
	result, _, err := r.client.From("booking_furniture_items").Select("*", "", false).Eq("furniture_id", bookingID).Single().Execute()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(result, &furnitureItem); err != nil {
		return nil, err
	}

	return &furnitureItem, nil
}

func (r *SupabaseBookingRepository) Update(ctx context.Context, booking *models.Booking) error {
	booking.UpdatedAt = time.Now()
	// "user_id" is the DB column name used for the WHERE clause — must remain unchanged
	_, _, err := r.client.From("booking_user").Update(booking, "", "").Eq("user_id", booking.BookingID).Execute() // R5: field renamed from UserID to BookingID
	return err
}

func (r *SupabaseBookingRepository) Delete(ctx context.Context, id string) error {
	// "user_id" is the DB column name — must remain unchanged
	_, _, err := r.client.From("booking_user").Delete("", "").Eq("user_id", id).Execute()
	return err
}

func (r *SupabaseBookingRepository) List(ctx context.Context) ([]*models.Booking, error) {
	var bookings []*models.Booking
	result, _, err := r.client.From("booking_user").Select("*", "", false).Execute()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(result, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}
