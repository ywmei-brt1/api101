package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "io/ioutil" // Import the ioutil package
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/skip2/go-qrcode"
)

// For generateQR API
type GenerateRequest struct {
	Link string `json:"link"`
}

// For generateQR API
type GenerateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Image   []byte `json:"image,omitempty"` // Optional for image data
}

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
	// longPollingClients is a slice to hold channels for long polling clients.
	longPollingClients = make(map[uintptr]chan ItemList) // Use a map
	longPollingMu      sync.Mutex                        // Mutex for long polling clients
)

// LongPollHandler handles long polling requests.
func LongPollHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Create a channel for this client.
	clientChan := make(chan ItemList)
	longPollingMu.Lock()
	longPollingClients[uintptr(unsafe.Pointer(&clientChan))] = clientChan // Store channel with its address as key
	longPollingMu.Unlock()

	// Remove the client channel when the connection closes.
	defer func() {
		longPollingMu.Lock()
		delete(longPollingClients, uintptr(unsafe.Pointer(&clientChan))) // Directly delete using the key
		longPollingMu.Unlock()
	}()

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Wait for new data or a timeout.
	for { // Loop to keep the connection alive
		select {
		case <-ctx.Done(): // Context canceled (timeout)
			http.Error(w, "Timeout", http.StatusRequestTimeout)
			return
		case data := <-clientChan:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)
			// Flush the response writer to send the data immediately
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

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
	updateClientsWithData()
}

// updateClientsWithData sends the current 'items' data to all long-polling clients.
func updateClientsWithData() {
	longPollingMu.Lock()
	defer longPollingMu.Unlock()

	for _, clientChan := range longPollingClients {
		select {
		case clientChan <- items: // Send the current 'items' data
		default:
			// Handle the case where the client channel is full (optional)
		}
	}
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
	// Use json.MarshalIndent for pretty printing
	jsonData, err := json.MarshalIndent(data, "", "  ") // Indent with 2 spaces
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
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

func GenerateQR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method %s not allowed", r.Method)
		return
	}

	var req GenerateRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error parsing request body: %v", err)
		return
	}

	// Validate link format (optional, adjust as needed)
	if !isValidGoogleMeetLink(&req) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid Google Meet link format: %v", req.Link)
		return
	}

	qr, err := qrcode.New(req.Link, qrcode.Medium)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error generating QR code: %v", err)
		return
	}

	qrImage, err := qr.PNG(256) // Adjust size as needed
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error encoding QR code as PNG: %v", err)
		return
	}

	response := GenerateResponse{Success: true, Message: "QR code generated successfully"}
	response.Image = qrImage

	// Set the Content-Type header
	w.Header().Set("Content-Type", "image/png")
	// Write the image data directly to the response writer
	_, err = w.Write(qrImage)
	if err != nil {
		http.Error(w, "Error writing image to response", http.StatusInternalServerError)
		return
	}
}

func isValidGoogleMeetLink(req *GenerateRequest) bool {
	// Check for valid base URL prefix
	if !strings.HasPrefix(req.Link, "https://meet.google.com/") {
		return false
	}

	// Attempt to split and extract meeting code
	urlParts := strings.SplitN(req.Link, "?", 2)
	if len(urlParts) == 1 { // No query parameters
		return true // Valid format without query parameters
	}

	// Try to reset link with just base URL
	req.Link = urlParts[0] // Reset link to base URL

	// Validate meeting code format (optional)
	// Replace this with your logic to check the validity of the meeting code in urlParts[0][1:]
	// You can use regular expressions or other techniques for stricter validation
	return true // Assuming we don't perform additional validation here

	// Alternatively, you could return true conditionally based on further validation
	// if validateMeetingCode(urlParts[0][1:]) {
	//     return true
	// } else {
	//     return false // Invalid meeting code format even after reset
	// }
}
