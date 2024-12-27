package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/jcelliott/lumber"
)

const Version = "1.0.1"

type (
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Trace(string, ...interface{})
		Debug(string, ...interface{})
	}

	Driver struct {
		mutex sync.Mutex
		mutexes map[string]*sync.Mutex
		dir string
		log Logger
	}
)

type Options struct {
	Logger 
}

func New(dir string, options *Options)(*Driver, error) {
	dir = filepath.Clean(dir)

	opts := Options{}

	if options != nil {
		opts = *options
	}

	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger(lumber.INFO)
	}

	driver := Driver {
		dir: dir,
		mutexes: make(map[string]*sync.Mutex),
		log: opts.Logger,
	}

	if _, err := os.Stat(dir); err == nil {
		opts.Logger.Debug("Using '%s' (database already exists)\n", dir)
		return &driver, nil
	}

	opts.Logger.Debug("Creating the database at '%s ...'\n", dir)

	return &driver, os.Mkdir(dir, 0755)
}

func (d *Driver) Write(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("collection cannot be empty")
	}

	if resource == ""{
		return fmt.Errorf("resource cannot be empty")
	}

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collection)

	finalPath := filepath.Join(dir, resource + ".json")

	tempPath := finalPath + ".tmp"
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t")

	if err != nil {
		return err
	}

	b = append(b, byte('\n'))

	if err := ioutil.WriteFile(tempPath, b, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, finalPath)
}

func (d *Driver) Read(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("collection cannot be empty")
	}
	
	if resource == "" {
		return fmt.Errorf("resource cannot be empty")
	}

	record := filepath.Join(d.dir, collection, resource)

	if _, err := stat(record); err != nil {
		return err
	}

	b, err := ioutil.ReadFile(record + ".json")

	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}

func (d *Driver) ReadAll(collection string)([]string, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection cannot be empty")
	}

	dir := filepath.Join(d.dir, collection)

	if _, err := stat(dir); err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(dir)

	if err != nil {
		return nil, err
	}

	records := []string{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		b, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))

		if err != nil {
			return nil, err
		}

		records = append(records, string(b))
	}

	return records, nil
}

func (d *Driver) Delete(collection, resource string) error {
	if collection == "" {
		return fmt.Errorf("collection cannot be empty")
	}

	path := filepath.Join(collection, resource)

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, path)

	switch fi, err := stat(dir);  {
		case fi == nil, err != nil:
			return fmt.Errorf("resource does not exist")
		case fi.Mode().IsDir():
			return os.RemoveAll(dir)
		case fi.Mode().IsRegular():
			return os.RemoveAll(dir + ".json")
	}

	return nil
}

func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex {
	m, ok := d.mutexes[collection]

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !ok {
		m = &sync.Mutex{} // create an empty mutex
		d.mutexes[collection] = m
	}

	return m
}

func stat(path string)(fi os.FileInfo, err error) {
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json") // dbs we create will have .json extension
	}

	return 
}


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

	if err := db.Delete("users", "May"); err != nil {
		fmt.Println("Error: ", err)
	}

	if err := db.Delete("user", ""); err != nil {
		fmt.Println("Error", err)
	}
}