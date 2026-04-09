package store

import (
	"testing"

	"github.com/mrckurz/CI-CD-MCM/internal/model"
)

// TestCreateAndGet - Create a product, verify GetByID returns it
func TestCreateAndGet(t *testing.T) {
	s := NewMemoryStore()

	// Create a new product
	p := model.Product{Name: "Test Product", Price: 15.99}
	created := s.Create(p)

	// Verify GetByID returns it
	fetched, err := s.GetByID(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if fetched.Name != p.Name || fetched.Price != p.Price {
		t.Errorf("expected product name %s and price %f, got %s and %f", p.Name, p.Price, fetched.Name, fetched.Price)
	}
}

// TestUpdateProduct - Create, update, verify update was applied
func TestUpdateProduct(t *testing.T) {
	s := NewMemoryStore()
	created := s.Create(model.Product{Name: "Old Name", Price: 10.0})

	updateData := model.Product{Name: "New Name", Price: 20.0}

	// Update the product
	updated, err := s.Update(created.ID, updateData)
	if err != nil {
		t.Fatalf("expected no error on update, got %v", err)
	}

	// Verify update was applied
	if updated.Name != "New Name" || updated.Price != 20.0 {
		t.Errorf("expected updated values, got %v", updated)
	}

	// Double check by fetching it again
	fetched, _ := s.GetByID(created.ID)
	if fetched.Name != "New Name" {
		t.Errorf("expected fetched product to have the new name")
	}
}

// TestDeleteProduct - Create, delete, verify GetByID returns error
func TestDeleteProduct(t *testing.T) {
	s := NewMemoryStore()
	created := s.Create(model.Product{Name: "To Delete", Price: 5.0})

	// Delete the product
	err := s.Delete(created.ID)
	if err != nil {
		t.Fatalf("expected no error on delete, got %v", err)
	}

	// Verify GetByID returns ErrNotFound
	_, err = s.GetByID(created.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}
}

// TestGetByIDNotFound - Verify non-existent ID returns ErrNotFound (Using Table-Driven Tests)
func TestGetByIDNotFound(t *testing.T) {
	s := NewMemoryStore()

	// Table-driven test cases
	tests := []struct {
		name string
		id   int
	}{
		{"Zero ID", 0},
		{"Negative ID", -1},
		{"Large Non-existent ID", 9999},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.GetByID(tc.id)
			if err != ErrNotFound {
				t.Errorf("expected ErrNotFound for id %d, got %v", tc.id, err)
			}
		})
	}
}

// Keeping the originally provided tests
func TestGetAllEmpty(t *testing.T) {
	s := NewMemoryStore()
	products := s.GetAll()
	if len(products) != 0 {
		t.Errorf("expected 0 products, got %d", len(products))
	}
}

func TestDeleteNonExistent(t *testing.T) {
	s := NewMemoryStore()
	err := s.Delete(999)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound when deleting non-existent product")
	}
}
