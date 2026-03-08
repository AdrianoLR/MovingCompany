package service

import (
	"MovingCompanyGo/models"
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

// invoiceConfig holds business details read from environment variables
type invoiceConfig struct {
	companyName string // display name printed at the top of every invoice
	abn         string // Australian Business Number — legally required on invoices
	country     string // country shown in the billing address section
}

// loadInvoiceConfig reads business details from environment variables with safe defaults.
func loadInvoiceConfig() invoiceConfig {
	cfg := invoiceConfig{
		companyName: os.Getenv("COMPANY_NAME"),    // e.g. "The Furniture Man"
		abn:         os.Getenv("COMPANY_ABN"),     // e.g. "89670942686"
		country:     os.Getenv("COMPANY_COUNTRY"), // e.g. "Australia"
	}
	if cfg.companyName == "" {
		cfg.companyName = "The Furniture Man" // fallback for local development
	}
	if cfg.abn == "" {
		cfg.abn = "89670942686" // fallback — replace by setting COMPANY_ABN in production
	}
	if cfg.country == "" {
		cfg.country = "Australia" // fallback for local development
	}
	return cfg
}

// invoiceNumber derives a short, unique reference from the booking ID.
func invoiceNumber(bookingID string) string {
	if len(bookingID) >= 8 {
		return strings.ToUpper(bookingID[:8]) // e.g. "A3F2B1C4" from a UUID
	}
	return bookingID // fallback for unexpectedly short IDs
}

// GenerateSampleInvoice generates a PDF invoice from booking data and returns its bytes.
func GenerateSampleInvoice(userData *models.Booking, furnitureItems *models.FurnitureItem, totalAmount float64, hoursUsed float64, jobDescription string) ([]byte, error) {
	cfg := loadInvoiceConfig()

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Company header
	pdf.SetFont("Arial", "B", 16)
	pdf.SetXY(15, 20)
	pdf.Cell(0, 10, cfg.companyName)
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 10)
	pdf.SetX(15)
	pdf.Cell(0, 5, "Moving Services")

	pdf.SetTextColor(0, 0, 0) // reset to black after any colour changes

	// BILL TO section
	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(15, 70)
	pdf.Cell(0, 8, "BILL TO")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.SetX(15)
	pdf.Cell(0, 5, userData.CustomerName)
	pdf.Ln(5)
	pdf.SetX(15)
	pdf.Cell(0, 5, userData.PickupAddress)
	pdf.Ln(5)
	pdf.SetX(15)
	pdf.Cell(0, 5, fmt.Sprintf("Drop-off: %s", userData.DropAddress))
	pdf.Ln(5)
	pdf.SetX(15)
	pdf.Cell(0, 5, cfg.country)
	pdf.Ln(10)

	// ABN row
	pdf.SetX(15)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(20, 5, "ABN:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, cfg.abn)

	// Invoice metadata (right column)
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(120, 80)
	pdf.Cell(35, 5, "Invoice No.:")
	pdf.Cell(0, 5, invoiceNumber(userData.BookingID))
	pdf.Ln(5)

	pdf.SetX(120)
	pdf.Cell(35, 5, "Issue date:")
	pdf.Cell(0, 5, userData.PickupDate.Format("2 Jan 2006"))
	pdf.Ln(5)

	pdf.SetX(120)
	pdf.Cell(35, 5, "Due date:")
	pdf.Cell(0, 5, userData.PickupDate.Format("2 Jan 2006"))
	pdf.Ln(10)

	// Reference line
	pdf.SetX(120)
	pdf.Cell(23, 5, "Reference:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Ln(5)
	pdf.SetX(120)
	pdf.Cell(0, 3, fmt.Sprintf("Removal %s to %s",
		getShortAddress(userData.PickupAddress),
		getShortAddress(userData.DropAddress)))

	// Items table header — blue background, white text
	pdf.SetY(130)
	pdf.SetFillColor(135, 149, 255)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(255, 255, 255)

	pdf.CellFormat(80, 8, "DESCRIPTION", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 8, "QUANTITY", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "UNIT PRICE (AUD)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "AMOUNT (AUD)", "1", 1, "C", true, 0, "")

	pdf.SetTextColor(0, 0, 0) // back to black for data rows
	pdf.SetFont("Arial", "", 10)

	subtotal := totalAmount // the manually entered total is treated as the pre-GST subtotal

	// Primary service line
	pdf.CellFormat(80, 8, fmt.Sprintf("Removals %s to %s",
		getShortAddress(userData.PickupAddress),
		getShortAddress(userData.DropAddress)), "1", 0, "L", false, 0, "")
	pdf.CellFormat(25, 8, "1", "1", 0, "C", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", totalAmount), "1", 0, "R", false, 0, "")
	pdf.CellFormat(35, 8, fmt.Sprintf("%.2f", totalAmount), "1", 1, "R", false, 0, "")

	// Service time description row
	pdf.SetFont("Arial", "", 9)
	endTime := userData.PickupDate.Add(time.Duration(hoursUsed * float64(time.Hour)))
	serviceDesc := fmt.Sprintf("Job done %s - %s to %s",
		userData.PickupDate.Format("02/01"),
		userData.PickupDate.Format("15:04"),
		endTime.Format("15:04"))
	pdf.CellFormat(180, 6, serviceDesc, "1", 1, "L", false, 0, "")

	// Hours and job description row
	hoursText := fmt.Sprintf("%.1f Hours job %s", hoursUsed, jobDescription)
	pdf.CellFormat(180, 6, hoursText, "1", 1, "L", false, 0, "")

	// Furniture inventory rows — listed for reference; pricing is included in the main service row
	if furnitureItems != nil {
		pdf.SetFont("Arial", "", 9)
		// ItemList() returns only categories with qty > 0, eliminating the inline struct rebuild
		for _, item := range furnitureItems.ItemList() {
			pdf.CellFormat(180, 5, fmt.Sprintf("- %s: %d items", item.Name, item.Quantity), "1", 1, "L", false, 0, "")
		}
	}

	// Totals section
	pdf.Ln(5)

	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(0, 6, "This invoice was paid in full already")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.SetX(120)
	pdf.Cell(40, 6, "SUBTOTAL")
	pdf.SetX(160)
	pdf.Cell(20, 6, fmt.Sprintf("$%.2f", subtotal))
	pdf.Ln(6)

	gst := subtotal * 0.10 // 10% GST as required by Australian tax law
	pdf.SetX(120)
	pdf.Cell(40, 6, fmt.Sprintf("GST 10%% from %.2f", subtotal))
	pdf.SetX(160)
	pdf.Cell(20, 6, fmt.Sprintf("$%.2f", gst))
	pdf.Ln(6)

	total := subtotal + gst
	pdf.SetX(120)
	pdf.Cell(40, 6, "TOTAL (AUD)")
	pdf.SetX(160)
	pdf.Cell(20, 6, fmt.Sprintf("$%.2f", total))
	pdf.Ln(6)

	pdf.SetDrawColor(41, 84, 144) // blue rule under the total row
	pdf.Line(120, pdf.GetY()+2, 180, pdf.GetY()+2)

	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(0, 6, "Thanks for Supporting your local area business")

	if pdf.Error() != nil {
		return nil, fmt.Errorf("PDF generation error: %v", pdf.Error())
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to output PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// getShortAddress returns the first comma-delimited segment of an address string.
// Used to keep route descriptions concise on the invoice (e.g. "123 Main St" not the full postcode line).
func getShortAddress(address string) string {
	parts := strings.Split(address, ",")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return address
}
