package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mrckurz/CI-CD-MCM/internal/model"
	"github.com/mrckurz/CI-CD-MCM/internal/store"
)

func setupRouter() (*mux.Router, *Handler) {
	s := store.NewMemoryStore()
	h := NewHandler(s)
	r := mux.NewRouter()
	h.RegisterRoutes(r)
	return r, h
}

func TestHealthEndpoint(t *testing.T) {
	r, _ := setupRouter()

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestGetProductsEmpty(t *testing.T) {
	r, _ := setupRouter()

	req := httptest.NewRequest("GET", "/products", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestCreateAndGetProduct(t *testing.T) {
	r, _ := setupRouter()

	// Create
	body := `{"name":"Widget","price":9.99}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	// Get
	req = httptest.NewRequest("GET", "/products/1", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestGetProductNotFound(t *testing.T) {
	r, _ := setupRouter()

	req := httptest.NewRequest("GET", "/products/999", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateProduct(t *testing.T) {
	r, _ := setupRouter()

	body := `{"name":"Widget","price":9.99}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	body = `{"name":"Updated Widget","price":14.99}`
	req = httptest.NewRequest("PUT", "/products/1", strings.NewReader(body))
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var updated model.Product
	if err := json.NewDecoder(rr.Body).Decode(&updated); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if updated.ID != 1 {
		t.Errorf("expected id 1, got %d", updated.ID)
	}
	if updated.Name != "Updated Widget" {
		t.Errorf("expected name Updated Widget, got %s", updated.Name)
	}
	if updated.Price != 14.99 {
		t.Errorf("expected price 14.99, got %v", updated.Price)
	}
}

func TestDeleteProduct(t *testing.T) {
	r, _ := setupRouter()

	body := `{"name":"Widget","price":9.99}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	req = httptest.NewRequest("DELETE", "/products/1", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	req = httptest.NewRequest("GET", "/products/1", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestCreateInvalidProduct(t *testing.T) {
	r, _ := setupRouter()

	body := `{"name":"","price":9.99}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
