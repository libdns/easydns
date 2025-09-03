package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	easydns "github.com/libdns/easydns"
	"github.com/libdns/libdns"
)

// outputRecords prints a slice of records in a readable format
func outputRecords(records []libdns.Record, label string, zone string) {
	fmt.Println(label)

	// Calculate the width needed for the index based on total records
	var indexWidth int
	if len(records) == 0 {
		indexWidth = 1
	} else {
		indexWidth = len(fmt.Sprintf("%d", len(records)-1))
	}

	for i, record := range records {
		rr := record.RR()
		switch strings.ToUpper(rr.Type) {
		case "TXT":
			if txt, ok := record.(*libdns.TXT); ok {
				fmt.Printf("  [%*d] %s (.%s): %s, %s\n", indexWidth, i, rr.Name, zone, txt.Text, rr.Type)
			} else {
				fmt.Printf("  [%*d] %s (.%s): %s, %s\n", indexWidth, i, rr.Name, zone, rr.Data, rr.Type)
			}
		case "CNAME":
			if cname, ok := record.(*libdns.CNAME); ok {
				fmt.Printf("  [%*d] %s (.%s): %s, %s\n", indexWidth, i, rr.Name, zone, cname.Target, rr.Type)
			} else {
				fmt.Printf("  [%*d] %s (.%s): %s, %s\n", indexWidth, i, rr.Name, zone, rr.Data, rr.Type)
			}
		case "A", "AAAA":
			if addr, ok := record.(*libdns.Address); ok {
				fmt.Printf("  [%*d] %s (.%s): %s, %s\n", indexWidth, i, rr.Name, zone, addr.IP.String(), rr.Type)
			} else {
				fmt.Printf("  [%*d] %s (.%s): %s, %s\n", indexWidth, i, rr.Name, zone, rr.Data, rr.Type)
			}
		default:
			fmt.Printf("  [%*d] %s (.%s): %s, %s\n", indexWidth, i, rr.Name, zone, rr.Data, rr.Type)
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("EASYDNS_TOKEN")
	if token == "" {
		fmt.Printf("EASYDNS_TOKEN not set\n")
		return
	}
	key := os.Getenv("EASYDNS_KEY")
	if key == "" {
		fmt.Printf("EASYDNS_KEY not set\n")
		return
	}
	zone := os.Getenv("EASYDNS_ZONE")
	if zone == "" {
		fmt.Printf("EASYDNS_ZONE not set\n")
		return
	}
	provider := easydns.Provider{
		APIToken: token,
		APIKey:   key,
	}

	url := os.Getenv("EASYDNS_URL")
	if url != "" {
		provider.APIUrl = url
	}

	records, err := provider.GetRecords(context.TODO(), zone)
	if err != nil {
		log.Fatalln("ERROR: ", err.Error())
	}

	testName := "_acme-challenge.home"
	hasTestName := false
	var testRecord libdns.Record = libdns.RR{}

	// Print all existing records
	outputRecords(records, "Existing records in zone", zone)

	// Find the test record if it exists
	for _, record := range records {
		rr := record.RR()
		if rr.Name == testName {
			hasTestName = true
			testRecord = record
		}
	}

	if !hasTestName {
		appendedRecords, err := provider.AppendRecords(context.TODO(), zone, []libdns.Record{
			&libdns.TXT{
				Name: testName,
				TTL:  0,
				Text: "test_add_record_value",
			},
		})

		if err != nil {
			log.Fatalln("ERROR: ", err.Error())
		}

		outputRecords(appendedRecords, "Appended records in zone:", zone)
	} else if txt, ok := testRecord.(*libdns.TXT); ok && txt.Text == "test_add_record_value" {
		txt.Text = "test_update_record_value"
		updatedRecords, err := provider.SetRecords(context.TODO(), zone, []libdns.Record{
			testRecord,
		})

		if err != nil {
			log.Fatalln("ERROR: ", err.Error())
		}

		outputRecords(updatedRecords, "Updated records in zone:", zone)
	} else {
		deleteRecords, err := provider.DeleteRecords(context.TODO(), zone, []libdns.Record{
			testRecord,
		})

		if err != nil {
			log.Fatalln("ERROR: ", err.Error())
		}

		outputRecords(deleteRecords, "Deleted records in zone:", zone)
	}
}
