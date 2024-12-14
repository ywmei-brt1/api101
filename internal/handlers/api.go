package handlers

import (
	"encoding/json"
	"io/ioutil"
	_ "io/ioutil" // Import the ioutil package
	"net/http"
	"regexp"
	"sort"
	"sync"
	"time"
)

// Item represents a single entry with a timestamp and a string value.
type Item struct {
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`
}

// ItemList is a slice of Item that implements sort.Interface.
type ItemList []Item

func (p ItemList) Len() int           { return len(p) }
func (p ItemList) Less(i, j int) bool { return p[i].Timestamp.Before(p[j].Timestamp) }
func (p ItemList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

var (
	items ItemList
	mu    sync.Mutex // Add a mutex for thread safety
)

// PutHandler handles the PUT request to add a string with a timestamp.
func PutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body) // Read the request body
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	mu.Lock() // Acquire the lock before modifying the shared data
	item := Item{
		Timestamp: time.Now(),
		Value:     string(body),
	}
	items = append(items, item)

	// Keep only the latest 10 items
	if len(items) > 10 {
		sort.Sort(items)              // Sort by timestamp (oldest first)
		items = items[len(items)-10:] // Slice to keep only the last 10
	}

	mu.Unlock() // Release the lock

	w.WriteHeader(http.StatusCreated)
}

// GetHandler handles the GET request to retrieve all strings with timestamps.
func GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()        // Acquire the lock before reading the shared data
	sort.Sort(items) // Sort the items by timestamp
	data := make(ItemList, len(items))
	copy(data, items) // Create a copy to avoid data race
	mu.Unlock()       // Release the lock

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// SearchHandler handles the GET request to search for items matching a regex.
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	regexStr := r.URL.Query().Get("q") // Get the regex pattern from the "q" query parameter
	if regexStr == "" {
		http.Error(w, "Missing regex pattern", http.StatusBadRequest)
		return
	}

	regex, err := regexp.Compile(regexStr)
	if err != nil {
		http.Error(w, "Invalid regex", http.StatusBadRequest)
		return
	}

	mu.Lock()
	var results ItemList
	for _, item := range items {
		if regex.MatchString(item.Value) {
			results = append(results, item)
		}
	}
	mu.Unlock()

	sort.Sort(results) // Sort the results by timestamp

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
