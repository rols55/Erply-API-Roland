package database

import (
	"encoding/json"
	"testing"

	"github.com/erply/api-go-wrapper/pkg/api/customers"
)

func TestCreateJSONTable(t *testing.T) {
	err = CreateJSONTable()
	if err != nil {
		t.Error(err)
	}
}

func TestQueryCustomer(t *testing.T) {
	// Set up the test data
	searchParams := map[string]string{
		"fullName": "Mister White",
	}
	searchParamsJSON, _ := json.Marshal(searchParams)

	// Perform the query
	cachedCustomers, err := QueryCustomer(searchParamsJSON)

	// Verify the results
	if err == nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cachedCustomers != nil {
		t.Errorf("unexpected cached customers, expected nil")
	}
}

func TestInsertCustomers(t *testing.T) {
	// Set up the test data
	testCustomers := []customers.Customer{
		{
			FullName: "John Doe",
		},
		{
			FullName: "Jane Doe",
		},
	}
	searchParams := map[string]string{
		"lastName": "Doe",
	}
	searchParamsJSON, _ := json.Marshal(searchParams)

	// Insert the customers
	InsertCustomers(testCustomers, searchParamsJSON)

	// Query the cached customers
	cachedCustomers, err := QueryCustomer(searchParamsJSON)

	// Verify the results
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(cachedCustomers) != len(testCustomers) {
		t.Errorf("unexpected number of cached customers, got: %d, want: %d", len(cachedCustomers), len(testCustomers))
	}

	for i, c := range testCustomers {
		if c.LastName != cachedCustomers[i].LastName {
			t.Errorf("unexpected cached customer last name at index %d, got: %s, want: %s", i, cachedCustomers[i].LastName, c.LastName)
		}
	}
}

func TestCleanup(t *testing.T) {
	err = Cleanup()
	if err != nil {
		t.Error(err)
	}
}
