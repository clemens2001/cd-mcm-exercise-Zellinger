package handler

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mrckurz/CI-CD-MCM/internal/model"
	"github.com/mrckurz/CI-CD-MCM/internal/store"
)

const postgresHandlerTestDriverName = "postgres-handler-test"

var (
	registerPostgresHandlerTestDriver sync.Once
	postgresHandlerTestDBs            sync.Map
)

type postgresHandlerTestDB struct {
	pingErr            error
	getAllErr          error
	createErr          error
	createID           int
	updateErr          error
	updateRowsAffected int64
	deleteErr          error
	deleteRowsAffected int64
	products           []model.Product
}

type postgresHandlerTestDriver struct{}

func (d postgresHandlerTestDriver) Open(name string) (driver.Conn, error) {
	value, ok := postgresHandlerTestDBs.Load(name)
	if !ok {
		return nil, errors.New("test database not registered")
	}
	return &postgresHandlerTestConn{db: value.(*postgresHandlerTestDB)}, nil
}

type postgresHandlerTestConn struct {
	db *postgresHandlerTestDB
}

func (c *postgresHandlerTestConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("prepared statements are not supported in postgres handler tests")
}

func (c *postgresHandlerTestConn) Close() error {
	return nil
}

func (c *postgresHandlerTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not supported in postgres handler tests")
}

func (c *postgresHandlerTestConn) Ping(ctx context.Context) error {
	return c.db.pingErr
}

func (c *postgresHandlerTestConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(query, "ORDER BY id"):
		if c.db.getAllErr != nil {
			return nil, c.db.getAllErr
		}
		return postgresHandlerRowsForProducts(c.db.products), nil
	case strings.Contains(query, "WHERE id = $1"):
		id := postgresHandlerNamedValueAsInt(args[0])
		for _, product := range c.db.products {
			if product.ID == id {
				return postgresHandlerRowsForProducts([]model.Product{product}), nil
			}
		}
		return postgresHandlerRowsForProducts(nil), nil
	case strings.Contains(query, "INSERT INTO products"):
		if c.db.createErr != nil {
			return nil, c.db.createErr
		}
		return &postgresHandlerTestRows{
			columns: []string{"id"},
			values:  [][]driver.Value{{int64(c.db.createID)}},
		}, nil
	default:
		return nil, errors.New("unexpected query")
	}
}

func (c *postgresHandlerTestConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	switch {
	case strings.Contains(query, "UPDATE products"):
		return postgresHandlerTestResult(c.db.updateRowsAffected), c.db.updateErr
	case strings.Contains(query, "DELETE FROM products"):
		return postgresHandlerTestResult(c.db.deleteRowsAffected), c.db.deleteErr
	default:
		return nil, errors.New("unexpected exec")
	}
}

type postgresHandlerTestRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *postgresHandlerTestRows) Columns() []string {
	return r.columns
}

func (r *postgresHandlerTestRows) Close() error {
	return nil
}

func (r *postgresHandlerTestRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

type postgresHandlerTestResult int64

func (r postgresHandlerTestResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r postgresHandlerTestResult) RowsAffected() (int64, error) {
	return int64(r), nil
}

func setupPostgresRouter(t *testing.T, data *postgresHandlerTestDB) *mux.Router {
	t.Helper()

	registerPostgresHandlerTestDriver.Do(func() {
		sql.Register(postgresHandlerTestDriverName, postgresHandlerTestDriver{})
	})

	name := t.Name()
	postgresHandlerTestDBs.Store(name, data)
	t.Cleanup(func() {
		postgresHandlerTestDBs.Delete(name)
	})

	db, err := sql.Open(postgresHandlerTestDriverName, name)
	if err != nil {
		t.Fatalf("expected test database to open, got error: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	h := NewPostgresHandler(&store.PostgresStore{DB: db})
	r := mux.NewRouter()
	h.RegisterRoutes(r)
	return r
}

func postgresHandlerRowsForProducts(products []model.Product) driver.Rows {
	values := make([][]driver.Value, 0, len(products))
	for _, product := range products {
		values = append(values, []driver.Value{int64(product.ID), product.Name, product.Price})
	}
	return &postgresHandlerTestRows{
		columns: []string{"id", "name", "price"},
		values:  values,
	}
}

func postgresHandlerNamedValueAsInt(value driver.NamedValue) int {
	switch v := value.Value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}

func TestPostgresHealthEndpoint(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{})

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPostgresHealthEndpointDatabaseUnavailable(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{pingErr: errors.New("database down")})

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}

func TestPostgresGetProducts(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{
		products: []model.Product{{ID: 1, Name: "Widget", Price: 9.99}},
	})

	req := httptest.NewRequest("GET", "/products", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPostgresGetProductsError(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{getAllErr: errors.New("query failed")})

	req := httptest.NewRequest("GET", "/products", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestPostgresGetProduct(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{
		products: []model.Product{{ID: 1, Name: "Widget", Price: 9.99}},
	})

	req := httptest.NewRequest("GET", "/products/1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPostgresGetProductNotFound(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{})

	req := httptest.NewRequest("GET", "/products/999", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestPostgresCreateProduct(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{createID: 7})

	req := httptest.NewRequest("POST", "/products", strings.NewReader(`{"name":"Widget","price":9.99}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}

func TestPostgresCreateProductInvalidPayload(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{})

	req := httptest.NewRequest("POST", "/products", strings.NewReader(`{"name"`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPostgresCreateProductInvalidProduct(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{})

	req := httptest.NewRequest("POST", "/products", strings.NewReader(`{"name":"","price":9.99}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPostgresCreateProductStoreError(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{createErr: errors.New("insert failed")})

	req := httptest.NewRequest("POST", "/products", strings.NewReader(`{"name":"Widget","price":9.99}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestPostgresUpdateProduct(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{updateRowsAffected: 1})

	req := httptest.NewRequest("PUT", "/products/7", strings.NewReader(`{"name":"Updated","price":14.99}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPostgresUpdateProductInvalidPayload(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{})

	req := httptest.NewRequest("PUT", "/products/7", strings.NewReader(`{"name"`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPostgresUpdateProductNotFound(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{updateRowsAffected: 0})

	req := httptest.NewRequest("PUT", "/products/7", strings.NewReader(`{"name":"Missing","price":1}`))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestPostgresDeleteProduct(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{deleteRowsAffected: 1})

	req := httptest.NewRequest("DELETE", "/products/7", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPostgresDeleteProductNotFound(t *testing.T) {
	r := setupPostgresRouter(t, &postgresHandlerTestDB{deleteRowsAffected: 0})

	req := httptest.NewRequest("DELETE", "/products/7", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
