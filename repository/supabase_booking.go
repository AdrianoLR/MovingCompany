package repository

import (
	"MovingCompanyGo/config"
	"MovingCompanyGo/models"
	"context"
	"encoding/json"
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

	// Insert booking into booking_user table
	_, _, err := r.client.From("booking_user").Insert(booking, false, "", "", "").Execute()
	if err != nil {
		return err
	}

	// Set the booking ID in furniture items
	furnitureItems.FurnitureID = booking.UserID

	// Insert furniture items into booking_furniture_items table
	_, _, err = r.client.From("booking_furniture_items").Insert(furnitureItems, false, "", "", "").Execute()
	if err != nil {
		return err
	}

	return nil
}

func (r *SupabaseBookingRepository) GetByID(ctx context.Context, id string) (*models.Booking, error) {
	var booking models.Booking
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
	_, _, err := r.client.From("booking_user").Update(booking, "", "").Eq("user_id", booking.UserID).Execute()
	return err
}

func (r *SupabaseBookingRepository) Delete(ctx context.Context, id string) error {
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
