package main

import (
	"encoding/json"
	"fmt"
)

type Address struct {
	City string
	State string
	Country string
	Pincode json.Number	
}

type User struct {
	Name string
	Age json.Number
	Contact string
	Company string
	Address Address
}

func main() {
	dir := "./"

	db, err := New(dir, nil)

	if err != nil {
		fmt.Println("Error", err)
	}

	employees := []User{
		{"John Doe", "30", "1234567490", "ABC Inc", Address{"New York", "NY", "USA", "10001"}},
		{"Jane Doe", "25", "1234567290", "XYZ Inc", Address{"Los Angeles", "CA", "USA", "90001"}},
		{"John Smith", "35", "1234767890", "PQR Inc", Address{"Chicago", "IL", "USA", "60007"}},
		{"Carlos May", "55", "1234067890", "GQR Inc", Address{"Mayok", "IL", "USA", "60345"}},
		{"Maria Garcia", "45", "1234567890", "MQR Inc", Address{"Posto", "IL", "USA", "60089"}},
	}

	for _, employee := range employees {
		db.Write("users", employee.Name, User {
			Name: employee.Name,
			Age: employee.Age,
			Contact: employee.Contact,
			Company: employee.Company,
			Address: Address {
				City: employee.Address.City,
				State: employee.Address.State,
				Country: employee.Address.Country,
				Pincode: employee.Address.Pincode,
			},
		})
	}

	records, err := db.ReadAll("users")

	if err != nil {
		fmt.Println("Error", err)
	}

	fmt.Println("Records", records)

	// converting json to go
	allUsers := []User{}

	for _, record := range records {
		employee := User {}
		if err := json.Unmarshal([]byte(record), &employee); err != nil {
			fmt.Println("Error", err)
		}

		allUsers = append(allUsers, employee)
	}

	fmt.Println("All Users", allUsers)
}