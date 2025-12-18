package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Couchbase connection details
const (
	connectionString = "couchbases://cb.6mhtjxyi5juqnmgr.cloud.couchbase.com"
	username         = "aris"
	password         = "T1ku$H1t4m"
	bucketName       = "inquiry_balance"
	scopeName        = "master"
	accountCollName  = "ddmast"
	customerCollName = "cif"
)

// Global cluster instance
var cluster *gocb.Cluster

// Request and Response structures
type InquiryRequest struct {
	Account string `json:"account"`
}

type AccountData struct {
	AccountNumber       string  `json:"account_number"`
	AccountName         string  `json:"account_name"`
	CIF                 string  `json:"cif"`
	AccountType         string  `json:"account_type"`
	Currency            string  `json:"currency"`
	AvailableBalance    float64 `json:"available_balance"`
	HoldBalance         float64 `json:"hold_balance"`
	Status              string  `json:"status"`
	BranchCode          string  `json:"branch_code"`
	OpenDate            string  `json:"open_date"`
	LastTransactionDate string  `json:"last_transaction_date"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	Province   string `json:"province"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type CustomerData struct {
	CIF                 string  `json:"cif"`
	CustomerType        string  `json:"customer_type"`
	FullName            string  `json:"full_name"`
	DateOfBirth         *string `json:"date_of_birth"`
	IDType              string  `json:"id_type"`
	IDNumber            string  `json:"id_number"`
	TaxID               string  `json:"tax_id"`
	Email               string  `json:"email"`
	Phone               string  `json:"phone"`
	Mobile              string  `json:"mobile"`
	Address             Address `json:"address"`
	Occupation          string  `json:"occupation"`
	MaritalStatus       *string `json:"marital_status"`
	Nationality         string  `json:"nationality"`
	CustomerSegment     string  `json:"customer_segment"`
	RiskRating          string  `json:"risk_rating"`
	RelationshipManager string  `json:"relationship_manager"`
	OnboardingDate      string  `json:"onboarding_date"`
	KYCStatus           string  `json:"kyc_status"`
	KYCLastUpdated      string  `json:"kyc_last_updated"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

type InquiryResponse struct {
	ResponseCode    string        `json:"response_code"`
	ResponseMessage string        `json:"response_message"`
	Timestamp       string        `json:"timestamp"`
	Account         *AccountData  `json:"account,omitempty"`
	Customer        *CustomerData `json:"customer,omitempty"`
}

type ErrorResponse struct {
	ResponseCode    string `json:"response_code"`
	ResponseMessage string `json:"response_message"`
	Timestamp       string `json:"timestamp"`
	Error           string `json:"error,omitempty"`
}

// Initialize Couchbase connection
func initCouchbase() error {
	options := gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
	}

	// Apply WAN development profile
	if err := options.ApplyProfile(gocb.ClusterConfigProfileWanDevelopment); err != nil {
		return fmt.Errorf("failed to apply profile: %w", err)
	}

	// Connect to cluster
	var err error
	cluster, err = gocb.Connect(connectionString, options)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	// Wait until cluster is ready
	err = cluster.WaitUntilReady(10*time.Second, nil)
	if err != nil {
		return fmt.Errorf("cluster not ready: %w", err)
	}

	log.Println("✓ Connected to Couchbase successfully")
	return nil
}

// Get account data from Couchbase
func getAccountData(accountNumber string) (*AccountData, error) {
	bucket := cluster.Bucket(bucketName)
	err := bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		return nil, fmt.Errorf("bucket not ready: %w", err)
	}

	scope := bucket.Scope(scopeName)
	collection := scope.Collection(accountCollName)

	// Get document by key (account number)
	result, err := collection.Get(accountNumber, nil)
	if err != nil {
		return nil, err
	}

	var account AccountData
	err = result.Content(&account)
	if err != nil {
		return nil, fmt.Errorf("failed to decode account data: %w", err)
	}

	return &account, nil
}

// Get customer data from Couchbase
func getCustomerData(cif string) (*CustomerData, error) {
	bucket := cluster.Bucket(bucketName)
	err := bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		return nil, fmt.Errorf("bucket not ready: %w", err)
	}

	scope := bucket.Scope(scopeName)
	collection := scope.Collection(customerCollName)

	// Get document by key (CIF)
	result, err := collection.Get(cif, nil)
	if err != nil {
		return nil, err
	}

	var customer CustomerData
	err = result.Content(&customer)
	if err != nil {
		return nil, fmt.Errorf("failed to decode customer data: %w", err)
	}

	return &customer, nil
}

// Handler for inquiry balance endpoint
func inquiryBalanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req InquiryRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			ResponseCode:    "400",
			ResponseMessage: "Invalid request format",
			Timestamp:       time.Now().Format(time.RFC3339),
			Error:           err.Error(),
		})
		return
	}

	// Validate account number
	if req.Account == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			ResponseCode:    "400",
			ResponseMessage: "Account number is required",
			Timestamp:       time.Now().Format(time.RFC3339),
		})
		return
	}

	// Get account data
	account, err := getAccountData(req.Account)
	if err != nil {
		// Check if it's a key not found error
		if gocb.IsKeyNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{
				ResponseCode:    "404",
				ResponseMessage: "Account not found",
				Timestamp:       time.Now().Format(time.RFC3339),
			})
			return
		}

		// Other errors
		log.Printf("Error getting account data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			ResponseCode:    "500",
			ResponseMessage: "Internal server error while retrieving account data",
			Timestamp:       time.Now().Format(time.RFC3339),
			Error:           err.Error(),
		})
		return
	}

	// Get customer data using CIF
	customer, err := getCustomerData(account.CIF)
	if err != nil {
		// Log error but still return account data
		log.Printf("Error getting customer data for CIF %s: %v", account.CIF, err)

		// Return account data without customer info
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(InquiryResponse{
			ResponseCode:    "200",
			ResponseMessage: "Account found but customer data unavailable",
			Timestamp:       time.Now().Format(time.RFC3339),
			Account:         account,
			Customer:        nil,
		})
		return
	}

	// Success response with both account and customer data
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(InquiryResponse{
		ResponseCode:    "200",
		ResponseMessage: "Success",
		Timestamp:       time.Now().Format(time.RFC3339),
		Account:         account,
		Customer:        customer,
	})
}

// Health check endpoint
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "inquiry-balance-api",
		"version":   "1.0.0",
	})
}

// CORS middleware
func setupCORS() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	})
}

func main() {
	// Initialize Couchbase connection
	log.Println("Initializing Couchbase connection...")
	if err := initCouchbase(); err != nil {
		log.Fatalf("Failed to initialize Couchbase: %v", err)
	}
	defer cluster.Close(nil)

	// Setup router
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/v1/inquiry", inquiryBalanceHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/health", healthCheckHandler).Methods("GET")

	// Setup CORS
	handler := setupCORS().Handler(router)

	// Start server
	port := ":8080"
	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📡 Endpoint: http://localhost%s/api/v1/inquiry", port)
	log.Printf("❤️  Health check: http://localhost%s/api/v1/health", port)

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
