package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gurch101.github.io/go-web/pkg/dbutils"
	"gurch101.github.io/go-web/pkg/testutils"
)

func TestCreateTenant(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Define the input JSON for the request
	createTenantRequest := map[string]interface{}{
		"tenantName":   "TestTenant",
		"contactEmail": "test@example.com",
		"plan":         "free",
	}

	// Create a new HTTP request
	req := testutils.CreatePostRequest(t, "/tenants", createTenantRequest)

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201 Created, got %d", rr.Code)
	}

	// Check the response body
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["id"] == nil {
		t.Error("Expected non-nil ID, got nil")
	}

	var tenantID int64
	err = db.QueryRow("SELECT id FROM tenants WHERE tenant_name = ?", "TestTenant").Scan(&tenantID)
	if err != nil {
		t.Fatalf("Failed to query tenant: %v", err)
	}
}

func TestCreateTenantInvalidPlan(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Define the input JSON for the request
	createTenantRequest := map[string]interface{}{
		"tenantName":   "TestTenant",
		"contactEmail": "acme@acme.com",
		"plan":         "invalid",
	}

	// Create a new HTTP request
	req := testutils.CreatePostRequest(t, "/tenants", createTenantRequest)

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Created, got %d", rr.Code)
	}

	// Check the response body
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	testutils.AssertError(t, response, "plan", "Invalid plan")
}

func TestCreateTenant_DuplicateTenant(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Define the input JSON for the request
	createTenantRequest := map[string]interface{}{
		"tenantName":   "TestTenant",
		"contactEmail": "acme@acme.com",
		"plan":         "free",
	}

	// Create a new HTTP request
	req := testutils.CreatePostRequest(t, "/tenants", createTenantRequest)

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201 Created, got %d", rr.Code)
	}

	req = testutils.CreatePostRequest(t, "/tenants", createTenantRequest)
	rr = httptest.NewRecorder()
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got %d", rr.Code)
	}

	// Check the response body
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	testutils.AssertError(t, response, "tenantName", "This tenant is already registered")
}

func TestGetTenantHandler(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Create a new HTTP request
	req := testutils.CreateGetRequest(t, "/tenants/1")

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}

	// Check the response body
	var response GetTenantResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("Expected tenant ID 1, got %d", response.ID)
	}
	if response.TenantName != "Acme" {
		t.Errorf("Expected tenant name 'Acme', got '%s'", response.TenantName)
	}
	if response.ContactEmail != "admin@acme.com" {
		t.Errorf("Expected contact email 'admin@acme.com', got '%s'", response.ContactEmail)
	}
	if response.Plan != Free {
		t.Errorf("Expected plan 'free', got '%s'", response.Plan)
	}
	if !response.IsActive {
		t.Errorf("Expected tenant to be active")
	}
}

func TestGetTenantHandler_InvalidID(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Create a new HTTP request with an invalid ID
	req := testutils.CreateGetRequest(t, "/tenants/invalid")

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found, got %d", rr.Code)
	}
}

func TestGetTenantHandler_NotFound(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Create a new HTTP request for a non-existent tenant
	req := testutils.CreateGetRequest(t, "/tenants/9999")

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found, got %d", rr.Code)
	}
}

func TestDeleteTenantHandler(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Create a new HTTP request
	req := testutils.CreateDeleteRequest(t, "/tenants/1")

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Verify that the tenant has been deleted
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM tenants WHERE id = 1").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query tenant: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected tenant to be deleted, but it still exists")
	}
}

func TestDeleteTenantHandler_InvalidID(t *testing.T) {
	db := dbutils.SetupTestDB(t)
	tenantController := NewTenantController(db)
	req := testutils.CreateDeleteRequest(t, "/tenants/invalid")
	rr := httptest.NewRecorder()
	tenantController.GetMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestDeleteTenantHandler_NotFound(t *testing.T) {
	db := dbutils.SetupTestDB(t)
	tenantController := NewTenantController(db)
	req := testutils.CreateDeleteRequest(t, "/tenants/9999")
	rr := httptest.NewRecorder()
	tenantController.GetMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestUpdateTenantHandler(t *testing.T) {
	db := dbutils.SetupTestDB(t)

	// Create the TenantController instance with the test database
	tenantController := NewTenantController(db)

	// Define the input JSON for the update request
	updateTenantRequest := map[string]interface{}{
		"tenantName":   "UpdatedTenant",
		"contactEmail": "updated@example.com",
		"plan":         "paid",
		"isActive":     false,
	}

	// Create a new HTTP request for the update
	req := testutils.CreatePatchRequest(t, "/tenants/1", updateTenantRequest)

	// Record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request through the controller
	tenantController.GetMux().ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}

	// Check the response body
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["tenantName"] != updateTenantRequest["tenantName"] {
		t.Errorf("Expected tenant name '%s', got '%s'", updateTenantRequest["tenantName"], response["tenantName"])
	}

	// Verify database update
	var updatedTenant struct {
		TenantName   string
		ContactEmail string
		Plan         string
		IsActive     bool
	}
	err = db.QueryRow(`SELECT tenant_name, contact_email, plan, is_active FROM tenants WHERE id = 1`).
		Scan(&updatedTenant.TenantName, &updatedTenant.ContactEmail, &updatedTenant.Plan, &updatedTenant.IsActive)
	if err != nil {
		t.Fatalf("Failed to query updated tenant: %v", err)
	}

	if updatedTenant.TenantName != updateTenantRequest["tenantName"] {
		t.Errorf("Expected tenant name '%s', got '%s'", updateTenantRequest["tenantName"], updatedTenant.TenantName)
	}
	if updatedTenant.ContactEmail != updateTenantRequest["contactEmail"] {
		t.Errorf("Expected contact email '%s', got '%s'", updateTenantRequest["contactEmail"], updatedTenant.ContactEmail)
	}
	if updatedTenant.Plan != updateTenantRequest["plan"] {
		t.Errorf("Expected plan '%s', got '%s'", updateTenantRequest["plan"], updatedTenant.Plan)
	}
	if updatedTenant.IsActive != updateTenantRequest["isActive"] {
		t.Errorf("Expected isActive '%v', got '%v'", updateTenantRequest["isActive"], updatedTenant.IsActive)
	}
}

func TestUpdateTenantHandler_InvalidID(t *testing.T) {
	db := dbutils.SetupTestDB(t)
	tenantController := NewTenantController(db)
	req := testutils.CreatePatchRequest(t, "/tenants/invalid", map[string]interface{}{})
	rr := httptest.NewRecorder()
	tenantController.GetMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestUpdateTenantHandler_NotFound(t *testing.T) {
	db := dbutils.SetupTestDB(t)
	tenantController := NewTenantController(db)
	req := testutils.CreatePatchRequest(t, "/tenants/9999", map[string]interface{}{})
	rr := httptest.NewRecorder()
	tenantController.GetMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestUpdateTenantHandler_InvalidRequest(t *testing.T) {
	db := dbutils.SetupTestDB(t)
	tenantController := NewTenantController(db)
	req := testutils.CreatePatchRequest(t, "/tenants/1", map[string]interface{}{
		"tenantName":   "UpdatedTenant",
		"contactEmail": "updated@example.com",
		"plan":         "invalid_plan",
		"isActive":     true,
	})
	rr := httptest.NewRecorder()
	tenantController.GetMux().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}
