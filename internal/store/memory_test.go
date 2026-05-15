package store

import (
	"reflect"
	"testing"

	"github.com/mrckurz/CI-CD-MCM/internal/model"
)

func TestCreateAndGet(t *testing.T) {
	s := NewMemoryStore()

	created := s.Create(model.Product{Name: "Widget", Price: 9.99})
	if created.ID != 1 {
		t.Fatalf("expected ID 1, got %d", created.ID)
	}

	got, err := s.GetByID(created.ID)
	if err != nil {
		t.Fatalf("expected product, got error: %v", err)
	}
	if !reflect.DeepEqual(got, created) {
		t.Fatalf("expected %+v, got %+v", created, got)
	}
}

func TestGetAllEmpty(t *testing.T) {
	s := NewMemoryStore()
	products := s.GetAll()
	if len(products) != 0 {
		t.Errorf("expected 0 products, got %d", len(products))
	}
}

func TestGetAllReturnsProducts(t *testing.T) {
	s := NewMemoryStore()
	first := s.Create(model.Product{Name: "Widget", Price: 9.99})
	second := s.Create(model.Product{Name: "Gadget", Price: 19.99})

	products := s.GetAll()
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}

	byID := map[int]model.Product{}
	for _, product := range products {
		byID[product.ID] = product
	}

	if !reflect.DeepEqual(byID[first.ID], first) {
		t.Fatalf("expected first product %+v, got %+v", first, byID[first.ID])
	}
	if !reflect.DeepEqual(byID[second.ID], second) {
		t.Fatalf("expected second product %+v, got %+v", second, byID[second.ID])
	}
}

func TestGetByIDNonExistent(t *testing.T) {
	s := NewMemoryStore()

	_, err := s.GetByID(999)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateExistingProduct(t *testing.T) {
	s := NewMemoryStore()
	created := s.Create(model.Product{Name: "Widget", Price: 9.99})

	updated, err := s.Update(created.ID, model.Product{Name: "Updated", Price: 14.99})
	if err != nil {
		t.Fatalf("expected update to succeed, got error: %v", err)
	}

	expected := model.Product{ID: created.ID, Name: "Updated", Price: 14.99}
	if !reflect.DeepEqual(updated, expected) {
		t.Fatalf("expected %+v, got %+v", expected, updated)
	}

	got, err := s.GetByID(created.ID)
	if err != nil {
		t.Fatalf("expected updated product, got error: %v", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected stored product %+v, got %+v", expected, got)
	}
}

func TestUpdateNonExistentProduct(t *testing.T) {
	s := NewMemoryStore()

	_, err := s.Update(999, model.Product{Name: "Missing", Price: 1})
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteExistingProduct(t *testing.T) {
	s := NewMemoryStore()
	created := s.Create(model.Product{Name: "Widget", Price: 9.99})

	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("expected delete to succeed, got error: %v", err)
	}

	if _, err := s.GetByID(created.ID); err != ErrNotFound {
		t.Fatalf("expected deleted product to be missing, got %v", err)
	}
}

func TestDeleteNonExistent(t *testing.T) {
	s := NewMemoryStore()
	err := s.Delete(999)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound when deleting non-existent product")
	}
}
