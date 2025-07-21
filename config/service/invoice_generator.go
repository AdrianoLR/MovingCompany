package service

import (
	"MovingCompanyGo/models"
	"bytes"
	"log"
	"os"

	generator "github.com/angelodlfrtr/go-invoice-generator"
)

// GenerateSampleInvoice generates a sample invoice PDF and returns its bytes.
func GenerateSampleInvoice(userData *models.Booking) ([]byte, error) {
	doc, _ := generator.New(generator.Invoice, &generator.Options{
		TextTypeInvoice: "INVOICE",
		AutoPrint:       true,
	})
	//Header
	doc.SetRef("Bruno Nascimento")
	doc.SetVersion("1.0")
	doc.SetDate("Dado Dinamico")
	doc.SetPaymentTerm("Dado Dinamico")

	//Table
	doc.SetDescription("Description")

	logoBytes, err := os.ReadFile("./config/service/logo.png")
	if err != nil {
		log.Fatal(err)
	}

	doc.SetCustomer(&generator.Contact{
		Name: "The Furniture Man Removals",
		Address: &generator.Address{
			Address:    "26 Hastings St",
			PostalCode: "Scarborough WA 6019",
			City:       "Australia",
			Country:    "ABN: 420748904",
		},
	})

	doc.SetCompany(&generator.Contact{
		Name: "userData.CustomerName",
		Logo: logoBytes,
		Address: &generator.Address{
			Address:    "Dado dinamico",
			Address2:   "Dado dinamico 2",
			PostalCode: "Dado dinamico",
			City:       "Australia",
			// Country:    "",
		},
	})

	// for i := 0; i < 3; i++ {
	doc.AppendItem(&generator.Item{
		Name:        "Removals CB to Duncraig",
		Description: " Job done 31/03 - 9:00 to 11:00, 2 Hours job Flat pack delivery no required assembl",
		UnitCost:    "340",
		Quantity:    "1",
		// Tax: &generator.Tax{
		// 	Percent: "20",
		// },
	})
	// }

	// doc.AppendItem(&generator.Item{
	// 	Name:     "Test",
	// 	UnitCost: "3576.89",
	// 	Quantity: "2",
	// 	Discount: &generator.Discount{
	// 		Percent: "50",
	// 	},
	// })

	doc.SetDefaultTax(&generator.Tax{
		Percent: "10",
	})
	//Footer
	doc.SetNotes("Thanks for supporting your local business area")

	// doc.SetDiscount(&generator.Discount{
	// Percent: "90",
	// })
	// doc.SetDiscount(&generator.Discount{
	// 	Amount: "1340",
	// })

	pdf, err := doc.Build()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
