package service

import (
	"MovingCompanyGo/models"
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

// InvoiceItem represents an item in the invoice
type InvoiceItem struct {
	Name        string
	Description string
	Quantity    int
	UnitCost    float64
	Total       float64
}

// GenerateSampleInvoice generates an invoice PDF from booking data and returns its bytes.
func GenerateSampleInvoice(userData *models.Booking, furnitureItems *models.FurnitureItem, totalAmount float64, hoursUsed float64, jobDescription string) ([]byte, error) {
	// Create new PDF document
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add company name instead of logo for now
	pdf.SetFont("Arial", "B", 16)
	pdf.SetXY(15, 20)
	pdf.Cell(0, 10, "The Furniture Man")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 10)
	pdf.SetX(15)
	pdf.Cell(0, 5, "Moving Services")

	// Note: Logo loading disabled due to file format issues
	// You can add a valid PNG logo file at ./config/service/logo.png to enable it

	// Reset text color to black
	pdf.SetTextColor(0, 0, 0)

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
	pdf.Cell(0, 5, "Australia")
	pdf.Ln(10)

	// ABN
	pdf.SetX(15)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(20, 5, "ABN:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, "89670942686")

	// Invoice details (right side)
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(120, 80)
	pdf.Cell(35, 5, "Invoice No.:")
	pdf.Cell(0, 5, "001")
	pdf.Ln(5)

	pdf.SetX(120)
	pdf.Cell(35, 5, "Issue date:")
	pdf.Cell(0, 5, userData.PickupDate.Format("2 Jan 2006"))
	pdf.Ln(5)

	pdf.SetX(120)
	pdf.Cell(35, 5, "Due date:")
	pdf.Cell(0, 5, userData.PickupDate.Format("2 Jan 2006"))
	pdf.Ln(10)

	// Reference
	pdf.SetX(120)
	pdf.Cell(23, 5, "Reference:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Ln(5)
	pdf.SetX(120)
	pdf.Cell(0, 3, fmt.Sprintf("Removal %s to %s",
		getShortAddress(userData.PickupAddress),
		getShortAddress(userData.DropAddress)))

	// Items table header with blue background
	pdf.SetY(130)
	pdf.SetFillColor(135, 149, 255) // Light blue color
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(255, 255, 255) // White text

	pdf.CellFormat(80, 8, "DESCRIPTION", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 8, "QUANTITY", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "UNIT PRICE (AUD)", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "AMOUNT (AUD)", "1", 1, "C", true, 0, "")

	// Reset text color to black
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 10)

	// Use manual total amount as subtotal
	subtotal := totalAmount

	// Main service row with manual pricing
	pdf.CellFormat(80, 8, fmt.Sprintf("Removals %s to %s",
		getShortAddress(userData.PickupAddress),
		getShortAddress(userData.DropAddress)), "1", 0, "L", false, 0, "")
	pdf.CellFormat(25, 8, "1", "1", 0, "C", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", totalAmount), "1", 0, "R", false, 0, "")
	pdf.CellFormat(35, 8, fmt.Sprintf("%.2f", totalAmount), "1", 1, "R", false, 0, "")

	// Service description row with manual hours
	pdf.SetFont("Arial", "", 9)
	endTime := userData.PickupDate.Add(time.Duration(hoursUsed * float64(time.Hour)))
	serviceDesc := fmt.Sprintf("Job done %s - %s to %s",
		userData.PickupDate.Format("02/01"),
		userData.PickupDate.Format("15:04"),
		endTime.Format("15:04"))
	pdf.CellFormat(180, 6, serviceDesc, "1", 1, "L", false, 0, "")

	// Additional description with manual job description
	hoursText := fmt.Sprintf("%.1f Hours job %s", hoursUsed, jobDescription)
	pdf.CellFormat(180, 6, hoursText, "1", 1, "L", false, 0, "")

	// Add furniture items if any (for reference only, no pricing)
	if furnitureItems != nil {
		furnitureItemsList := []struct {
			name     string
			quantity int
		}{
			{"Chairs", furnitureItems.Chairs},
			{"Table (2 seats)", furnitureItems.Table2Seats},
			{"Table (3 seats)", furnitureItems.Table3Seats},
			{"Table (4+ seats)", furnitureItems.Table4PlusSeats},
			{"Fridges", furnitureItems.Fridges},
			{"Washing Machines", furnitureItems.WashingMachines},
			{"Dryers", furnitureItems.Dryers},
			{"Dishwashers", furnitureItems.Dishwashers},
			{"Boxes", furnitureItems.Boxes},
			{"Pot Plants", furnitureItems.PotPlants},
			{"Mattresses", furnitureItems.Mattresses},
			{"Bed Frames", furnitureItems.BedFrames},
			{"Sofas", furnitureItems.Sofas},
		}

		pdf.SetFont("Arial", "", 9)
		for _, item := range furnitureItemsList {
			if item.quantity > 0 {
				// Add furniture items as description only (included in main service price)
				pdf.CellFormat(180, 5, fmt.Sprintf("- %s: %d items", item.name, item.quantity), "1", 1, "L", false, 0, "")
			}
		}
	}

	// Totals section
	pdf.Ln(5)

	// Payment status
	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(0, 6, "This invoice was paid in full already")
	pdf.Ln(10)

	// Subtotal
	pdf.SetFont("Arial", "", 10)
	pdf.SetX(120)
	pdf.Cell(40, 6, "SUBTOTAL")
	pdf.SetX(160)
	pdf.Cell(20, 6, fmt.Sprintf("$%.2f", subtotal))
	pdf.Ln(6)

	// GST
	gst := subtotal * 0.10
	pdf.SetX(120)
	pdf.Cell(40, 6, fmt.Sprintf("GST 10%% from %.2f", subtotal))
	pdf.SetX(160)
	pdf.Cell(20, 6, fmt.Sprintf("$%.2f", gst))
	pdf.Ln(6)

	// Total
	total := subtotal + gst
	pdf.SetX(120)
	pdf.Cell(40, 6, "TOTAL (AUD)")
	pdf.SetX(160)
	pdf.Cell(20, 6, fmt.Sprintf("$%.2f", total))
	pdf.Ln(6)

	// Final total with blue line
	pdf.SetDrawColor(41, 84, 144)                  // Blue color for line
	pdf.Line(120, pdf.GetY()+2, 180, pdf.GetY()+2) // Blue line

	// Footer message
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(0, 6, "Thanks for Supporting your local area business")

	// Check for PDF errors before output
	if pdf.Error() != nil {
		return nil, fmt.Errorf("PDF generation error: %v", pdf.Error())
	}

	// Use bytes buffer for output
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to output PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// Helper function to get short address (first part before comma)
func getShortAddress(address string) string {
	parts := strings.Split(address, ",")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return address
}
