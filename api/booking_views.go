package api

import (
	"MovingCompanyGo/models"
	"html/template"
	"log"
	"net/http"
	"time"
)

// perthLocation is loaded once at startup; falls back to a fixed UTC+8 offset.
var perthLocation *time.Location

func init() {
	var err error
	perthLocation, err = time.LoadLocation("Australia/Perth")
	if err != nil {
		perthLocation = time.FixedZone("AWST", 8*60*60)
	}
}

// bookingTmpl holds all HTML fragment templates used by the admin dashboard.
var bookingTmpl = template.Must(template.New("").Parse(bookingTemplates))

const bookingTemplates = `
{{define "view-row"}}
<tr id="booking-row-{{.BookingID}}" data-pickup-date="{{.PickupDateISO}}">
    <td data-label="Customer">{{.CustomerName}}</td>
    <td data-label="Email">{{.Email}}</td>
    <td data-label="Phone">{{.Phone}}</td>
    <td data-label="Pickup Address">{{.PickupAddress}}</td>
    <td data-label="Drop Address">{{.DropAddress}}</td>
    <td data-label="Pickup Date">{{.PickupDateFormatted}}</td>
    <td data-label="Status">
        <select class="status-select" disabled>
            <option value="0"{{if eq .Status 0}} selected{{end}}>Pending</option>
            <option value="1"{{if eq .Status 1}} selected{{end}}>Confirmed</option>
            <option value="2"{{if eq .Status 2}} selected{{end}}>In Progress</option>
            <option value="3"{{if eq .Status 3}} selected{{end}}>Completed</option>
            <option value="4"{{if eq .Status 4}} selected{{end}}>Cancelled</option>
        </select>
    </td>
    <td data-label="Actions">
        <button class="edit-btn mr-2"
                hx-get="/api/bookings/{{.BookingID}}/edit"
                hx-target="closest tr"
                hx-swap="outerHTML">Edit</button>
        <button class="bg-green-600 text-white px-3 py-1 rounded text-sm hover:bg-green-700"
                onclick="generateBookingInvoice('{{.BookingID}}')">Invoice</button>
    </td>
</tr>
{{end}}

{{define "edit-row"}}
<tr id="booking-row-{{.BookingID}}">
    <td data-label="Customer">
        <input type="text" name="customer_name" class="w-full p-2 border rounded" value="{{.CustomerName}}">
    </td>
    <td data-label="Email">
        <input type="email" name="email" class="w-full p-2 border rounded" value="{{.Email}}">
    </td>
    <td data-label="Phone">
        <input type="tel" name="phone" class="w-full p-2 border rounded" value="{{.Phone}}">
    </td>
    <td data-label="Pickup Address">
        <input type="text" name="pickup_address" class="w-full p-2 border rounded" value="{{.PickupAddress}}">
    </td>
    <td data-label="Drop Address">
        <input type="text" name="drop_address" class="w-full p-2 border rounded" value="{{.DropAddress}}">
    </td>
    <td data-label="Pickup Date">
        <input type="datetime-local" name="pickup_date" class="w-full p-2 border rounded" value="{{.PickupDateLocal}}">
    </td>
    <td data-label="Status">
        <select name="status" class="status-select">
            <option value="0"{{if eq .Status 0}} selected{{end}}>Pending</option>
            <option value="1"{{if eq .Status 1}} selected{{end}}>Confirmed</option>
            <option value="2"{{if eq .Status 2}} selected{{end}}>In Progress</option>
            <option value="3"{{if eq .Status 3}} selected{{end}}>Completed</option>
            <option value="4"{{if eq .Status 4}} selected{{end}}>Cancelled</option>
        </select>
    </td>
    <td data-label="Actions">
        <button class="save-btn mr-2"
                hx-put="/api/bookings/{{.BookingID}}"
                hx-include="closest tr"
                hx-target="closest tr"
                hx-swap="outerHTML">Save</button>
        <button class="cancel-btn"
                hx-get="/api/bookings/{{.BookingID}}"
                hx-target="closest tr"
                hx-swap="outerHTML">Cancel</button>
    </td>
</tr>
{{end}}

{{define "table"}}
<table class="admin-table">
    <thead>
        <tr>
            <th>Customer</th>
            <th>Email</th>
            <th>Phone</th>
            <th>Pickup Address</th>
            <th>Drop Address</th>
            <th>Pickup Date</th>
            <th>Status</th>
            <th>Actions</th>
        </tr>
    </thead>
    <tbody>
        {{if .Rows}}
            {{range .Rows}}{{template "view-row" .}}{{end}}
        {{else}}
            <tr><td colspan="8" class="text-gray-500 p-4 text-center">No bookings found.</td></tr>
        {{end}}
    </tbody>
</table>
{{end}}

{{define "link-result"}}
<div id="link-result" class="mt-4 p-4 bg-gray-50 rounded">
    <div class="flex items-center">
        <label class="font-medium text-gray-700 mr-2">Booking Link:</label>
        <input type="text" id="booking-link" readonly
               class="flex-1 p-2 border rounded-l bg-white"
               onclick="this.select()"
               value="{{.}}">
        <button onclick="copyToClipboard('booking-link')"
                class="bg-gray-200 hover:bg-gray-300 text-gray-800 px-4 py-2 rounded-r">Copy</button>
    </div>
    <p class="mt-2 text-sm text-gray-500">This link will expire in 24 hours.</p>
</div>
{{end}}
`

// --- Data types ---------------------------------------------------------------

type viewRowData struct {
	BookingID           string
	CustomerName        string
	Email               string
	Phone               string
	PickupAddress       string
	DropAddress         string
	PickupDateFormatted string // display string in Perth time
	PickupDateISO       string // raw RFC3339 for the data attribute
	Status              int
}

type editRowData struct {
	BookingID       string
	CustomerName    string
	Email           string
	Phone           string
	PickupAddress   string
	DropAddress     string
	PickupDateLocal string // "YYYY-MM-DDTHH:mm" for datetime-local input
	Status          int
}

type tableData struct {
	Rows []viewRowData
}

// --- Converters ---------------------------------------------------------------

func bookingToViewRow(b *models.Booking) viewRowData {
	return viewRowData{
		BookingID:           b.BookingID,
		CustomerName:        b.CustomerName,
		Email:               b.Email,
		Phone:               b.Phone,
		PickupAddress:       b.PickupAddress,
		DropAddress:         b.DropAddress,
		PickupDateFormatted: b.PickupDate.In(perthLocation).Format("02/01/2006, 15:04"),
		PickupDateISO:       b.PickupDate.Format(time.RFC3339),
		Status:              b.Status,
	}
}

func bookingToEditRow(b *models.Booking) editRowData {
	return editRowData{
		BookingID:       b.BookingID,
		CustomerName:    b.CustomerName,
		Email:           b.Email,
		Phone:           b.Phone,
		PickupAddress:   b.PickupAddress,
		DropAddress:     b.DropAddress,
		PickupDateLocal: b.PickupDate.In(perthLocation).Format("2006-01-02T15:04"),
		Status:          b.Status,
	}
}

// --- Render helpers -----------------------------------------------------------

func renderViewRow(w http.ResponseWriter, b *models.Booking) {
	w.Header().Set("Content-Type", "text/html")
	if err := bookingTmpl.ExecuteTemplate(w, "view-row", bookingToViewRow(b)); err != nil {
		log.Printf("renderViewRow: template error: %v", err)
	}
}

func renderEditRow(w http.ResponseWriter, b *models.Booking) {
	w.Header().Set("Content-Type", "text/html")
	if err := bookingTmpl.ExecuteTemplate(w, "edit-row", bookingToEditRow(b)); err != nil {
		log.Printf("renderEditRow: template error: %v", err)
	}
}

func renderBookingsTable(w http.ResponseWriter, bookings []*models.Booking) {
	rows := make([]viewRowData, len(bookings))
	for i, b := range bookings {
		rows[i] = bookingToViewRow(b)
	}
	w.Header().Set("Content-Type", "text/html")
	if err := bookingTmpl.ExecuteTemplate(w, "table", tableData{Rows: rows}); err != nil {
		log.Printf("renderBookingsTable: template error: %v", err)
	}
}

func renderLinkResult(w http.ResponseWriter, url string) {
	w.Header().Set("Content-Type", "text/html")
	if err := bookingTmpl.ExecuteTemplate(w, "link-result", url); err != nil {
		log.Printf("renderLinkResult: template error: %v", err)
	}
}
