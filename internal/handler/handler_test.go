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

func TestCreateProductInvalidPayload(t *testing.T) {
	r, _ := setupRouter()

	req := httptest.NewRequest("POST", "/products", strings.NewReader(`{"name"`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestCreateProductInvalidProduct(t *testing.T) {
	r, _ := setupRouter()

	req := httptest.NewRequest("POST", "/products", strings.NewReader(`{"name":"","price":9.99}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
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
	r, h := setupRouter()
	created := h.Store.Create(model.Product{Name: "Widget", Price: 9.99})

	req := httptest.NewRequest("PUT", "/products/1", strings.NewReader(`{"name":"Updated","price":14.99}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var updated model.Product
	if err := json.NewDecoder(rr.Body).Decode(&updated); err != nil {
		t.Fatalf("expected product response, got error: %v", err)
	}
	if updated.ID != created.ID || updated.Name != "Updated" || updated.Price != 14.99 {
		t.Fatalf("unexpected updated product: %+v", updated)
	}
}

func TestUpdateProductInvalidPayload(t *testing.T) {
	r, h := setupRouter()
	h.Store.Create(model.Product{Name: "Widget", Price: 9.99})

	req := httptest.NewRequest("PUT", "/products/1", strings.NewReader(`{"name"`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpdateProductNotFound(t *testing.T) {
	r, _ := setupRouter()

	req := httptest.NewRequest("PUT", "/products/999", strings.NewReader(`{"name":"Missing","price":1}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDeleteProduct(t *testing.T) {
	r, h := setupRouter()
	created := h.Store.Create(model.Product{Name: "Widget", Price: 9.99})

	req := httptest.NewRequest("DELETE", "/products/1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if _, err := h.Store.GetByID(created.ID); err != store.ErrNotFound {
		t.Fatalf("expected deleted product to be missing, got %v", err)
	}
}

func TestDeleteProductNotFound(t *testing.T) {
	r, _ := setupRouter()

	req := httptest.NewRequest("DELETE", "/products/999", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
