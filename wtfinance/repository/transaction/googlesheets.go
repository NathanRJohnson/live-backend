package transaction

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/NathanRJohnson/live-backend/wtfinance/model"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsRepo struct {
	Service *sheets.Service
}

func (g *GoogleSheetsRepo) Insert(ctx context.Context, transaction model.Transaction, sheetRef string) error {

	// spreadsheetID := "13y1kKcEwJX4xsDwsPS5w_wBXh4mZch-YapETkv1POV8"
	columnRange := "Sheet1!A:A"

	resp, err := g.Service.Spreadsheets.Values.Get(sheetRef, columnRange).Do()
	if err != nil {
		log.Printf("Unable to retrieve data from column: %v", err)
		return err
	}
	nextEmptyRow := len(resp.Values) + 1

	values := [][]interface{}{
		{
			transaction.DateCreated.Format("1/2"),
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
	_, err = g.Service.Spreadsheets.Values.Update(sheetRef, writeRange, vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Printf("Unable to update data: %v", err)
		return err
	}

	fmt.Println("Spreadsheet updated successfully!")
	return err
}

func (g *GoogleSheetsRepo) FetchAll(ctx context.Context, sheetRef string) ([]model.Transaction, error) {
	readRange := "Sheet1!A:D" // Get all rows from columns A to D

	// Read the values from the specified range
	resp, err := g.Service.Spreadsheets.Values.Get(sheetRef, readRange).Do()
	if err != nil {
		log.Printf("Unable to retrieve data from range: %v", err)
		return nil, err
	}

	// Check if any values were returned
	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	}

	var transactions []model.Transaction
	for _, row := range resp.Values {
		amount64, err := strconv.ParseFloat(strings.Replace(fmt.Sprintf("%v", row[2]), "$", "", 1), 2)
		if err != nil {
			log.Printf("Bad value for field 'amount': %v: %v", row[2], err)
			continue
		}
		date, err := time.Parse("1/2", fmt.Sprintf("%v", row[0]))
		if err != nil {
			log.Printf("Bad value for field date: %v: %v", row[0], err)
			continue
		}
		transactions = append(transactions, model.Transaction{
			DateCreated: &date,
			Name:        fmt.Sprintf("%v", row[1]),
			Amount:      float32(amount64),
			Category:    fmt.Sprintf("%v", row[3]),
		})
	}

	return transactions, err
}

func (g *GoogleSheetsRepo) FetchCircleAmounts(ctx context.Context, sheetRef string) (interface{}, error) {
	readRange := "Sheet1!H43:H44"
	resp, err := g.Service.Spreadsheets.Values.Get(sheetRef, readRange).Do()
	if err != nil {
		log.Printf("Unable to retrieve data from range: %v", err)
		return nil, err
	}

	var amounts [2]float32
	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
		return nil, err
	}

	for i, row := range resp.Values {
		dollar_amount := fmt.Sprintf("%v", row[0])
		formatted_amount := strings.Replace(dollar_amount, "$", "", 1)
		amount := strings.Replace(formatted_amount, ",", "", 1)
		a, err := strconv.ParseFloat(amount, 2)
		if err != nil {
			fmt.Printf("Bad value for spent: %v", err)
			return nil, err
		}
		amounts[i] = float32(a)
	}

	total := amounts[0] + amounts[1]

	type CircleValues struct {
		Spent    float32 `json:"spent"`
		Overflow float32 `json:"overflow"`
		Total    float32 `json:"total"`
	}

	cv := CircleValues{
		Spent:    amounts[0],
		Overflow: amounts[1],
		Total:    total,
	}

	return cv, err
}
