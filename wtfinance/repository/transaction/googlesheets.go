package transaction

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/NathanRJohnson/live-backend/wtfinance/model"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsRepo struct {
	Service *sheets.Service
}

func (g *GoogleSheetsRepo) Insert(ctx context.Context, transaction model.Transaction) error {

	// b, err := os.ReadFile("../secrets/gsheets-serviceKey.json")
	// if err != nil {
	// 	log.Fatalf("Unable to read service account key file: %v", err)
	// }

	// srv, err := sheets.NewService(ctx, option.WithCredentialsJSON(b))
	// if err != nil {
	// 	log.Fatalf("Unable to create Sheets service: %v", err)
	// }

	// spreadsheetID := "13y1kKcEwJX4xsDwsPS5w_wBXh4mZch-YapETkv1POV8"
	columnRange := "Sheet1!A:A"

	resp, err := g.Service.Spreadsheets.Values.Get(transaction.SheetRef, columnRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from column: %v", err)
	}
	nextEmptyRow := len(resp.Values) + 1

	values := [][]interface{}{
		{
			transaction.DateCreated,
			transaction.Name,
			transaction.Amount,
			transaction.Category,
		},
	}

	// Prepare the value range
	vr := &sheets.ValueRange{
		Values: values,
	}

	writeRange := "Sheet1!A" + strconv.Itoa(nextEmptyRow) // Modify this to your desired range

	// Call the Sheets API to update the range
	_, err = g.Service.Spreadsheets.Values.Update(transaction.SheetRef, writeRange, vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Unable to update data: %v", err)
	}

	fmt.Println("Spreadsheet updated successfully!")
	return err
}
