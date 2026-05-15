package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/mrckurz/CI-CD-MCM/internal/model"
)

const postgresStoreTestDriverName = "postgres-store-test"

var (
	registerPostgresStoreTestDriver sync.Once
	postgresStoreTestDBs            sync.Map
)

type postgresStoreTestDB struct {
	pingErr            error
	ensureTableErr     error
	getAllErr          error
	getAllScanErr      bool
	getByIDErr         error
	createErr          error
	createID           int
	updateErr          error
	updateRowsAffected int64
	deleteErr          error
	deleteRowsAffected int64
	products           []model.Product
}

type postgresStoreTestDriver struct{}

func (d postgresStoreTestDriver) Open(name string) (driver.Conn, error) {
	value, ok := postgresStoreTestDBs.Load(name)
	if !ok {
		return nil, errors.New("test database not registered")
	}
	return &postgresStoreTestConn{db: value.(*postgresStoreTestDB)}, nil
}

type postgresStoreTestConn struct {
	db *postgresStoreTestDB
}

func (c *postgresStoreTestConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("prepared statements are not supported in postgres store tests")
}

func (c *postgresStoreTestConn) Close() error {
	return nil
}

func (c *postgresStoreTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not supported in postgres store tests")
}

func (c *postgresStoreTestConn) Ping(ctx context.Context) error {
	return c.db.pingErr
}

func (c *postgresStoreTestConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(query, "ORDER BY id"):
		if c.db.getAllErr != nil {
			return nil, c.db.getAllErr
		}
		if c.db.getAllScanErr {
			return &postgresStoreTestRows{
				columns: []string{"id", "name", "price"},
				values:  [][]driver.Value{{"invalid-id", "Broken", 1.0}},
			}, nil
		}
		return rowsForProducts(c.db.products), nil
	case strings.Contains(query, "WHERE id = $1"):
		if c.db.getByIDErr != nil {
			return nil, c.db.getByIDErr
		}
		id := namedValueAsInt(args[0])
		for _, product := range c.db.products {
			if product.ID == id {
				return rowsForProducts([]model.Product{product}), nil
			}
		}
		return rowsForProducts(nil), nil
	case strings.Contains(query, "INSERT INTO products"):
		if c.db.createErr != nil {
			return nil, c.db.createErr
		}
		return &postgresStoreTestRows{
			columns: []string{"id"},
			values:  [][]driver.Value{{int64(c.db.createID)}},
		}, nil
	default:
		return nil, errors.New("unexpected query")
	}
}

func (c *postgresStoreTestConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	switch {
	case strings.Contains(query, "CREATE TABLE"):
		return postgresStoreTestResult(0), c.db.ensureTableErr
	case strings.Contains(query, "UPDATE products"):
		return postgresStoreTestResult(c.db.updateRowsAffected), c.db.updateErr
	case strings.Contains(query, "DELETE FROM products"):
		return postgresStoreTestResult(c.db.deleteRowsAffected), c.db.deleteErr
	default:
		return nil, errors.New("unexpected exec")
	}
}

type postgresStoreTestRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *postgresStoreTestRows) Columns() []string {
	return r.columns
}

func (r *postgresStoreTestRows) Close() error {
	return nil
}

func (r *postgresStoreTestRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

type postgresStoreTestResult int64

func (r postgresStoreTestResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r postgresStoreTestResult) RowsAffected() (int64, error) {
	return int64(r), nil
}

func openPostgresStoreTestDB(t *testing.T, data *postgresStoreTestDB) *sql.DB {
	t.Helper()

	registerPostgresStoreTestDriver.Do(func() {
		sql.Register(postgresStoreTestDriverName, postgresStoreTestDriver{})
	})

	name := t.Name()
	postgresStoreTestDBs.Store(name, data)
	t.Cleanup(func() {
		postgresStoreTestDBs.Delete(name)
	})

	db, err := sql.Open(postgresStoreTestDriverName, name)
	if err != nil {
		t.Fatalf("expected test database to open, got error: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func rowsForProducts(products []model.Product) driver.Rows {
	values := make([][]driver.Value, 0, len(products))
	for _, product := range products {
		values = append(values, []driver.Value{int64(product.ID), product.Name, product.Price})
	}
	return &postgresStoreTestRows{
		columns: []string{"id", "name", "price"},
		values:  values,
	}
}

func namedValueAsInt(value driver.NamedValue) int {
	switch v := value.Value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}

func TestPostgresStoreEnsureTable(t *testing.T) {
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{})}

	if err := s.EnsureTable(); err != nil {
		t.Fatalf("expected table creation to succeed, got error: %v", err)
	}
}

func TestPostgresStoreEnsureTableError(t *testing.T) {
	expectedErr := errors.New("create failed")
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{ensureTableErr: expectedErr})}

	if err := s.EnsureTable(); err != expectedErr {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestPostgresStoreGetAll(t *testing.T) {
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{
		products: []model.Product{
			{ID: 1, Name: "Widget", Price: 9.99},
			{ID: 2, Name: "Gadget", Price: 19.99},
		},
	})}

	products, err := s.GetAll()
	if err != nil {
		t.Fatalf("expected products, got error: %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
}

func TestPostgresStoreGetAllEmptyReturnsSlice(t *testing.T) {
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{})}

	products, err := s.GetAll()
	if err != nil {
		t.Fatalf("expected empty products, got error: %v", err)
	}
	if products == nil {
		t.Fatal("expected empty slice, got nil")
	}
}

func TestPostgresStoreGetAllQueryError(t *testing.T) {
	expectedErr := errors.New("query failed")
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{getAllErr: expectedErr})}

	if _, err := s.GetAll(); err != expectedErr {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestPostgresStoreGetAllScanError(t *testing.T) {
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{getAllScanErr: true})}

	if _, err := s.GetAll(); err == nil {
		t.Fatal("expected scan error, got nil")
	}
}

func TestPostgresStoreGetByID(t *testing.T) {
	expected := model.Product{ID: 1, Name: "Widget", Price: 9.99}
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{products: []model.Product{expected}})}

	product, err := s.GetByID(1)
	if err != nil {
		t.Fatalf("expected product, got error: %v", err)
	}
	if product != expected {
		t.Fatalf("expected %+v, got %+v", expected, product)
	}
}

func TestPostgresStoreGetByIDNotFound(t *testing.T) {
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{})}

	if _, err := s.GetByID(999); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresStoreGetByIDError(t *testing.T) {
	expectedErr := errors.New("query failed")
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{getByIDErr: expectedErr})}

	if _, err := s.GetByID(1); err != expectedErr {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestPostgresStoreCreate(t *testing.T) {
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{createID: 42})}

	product, err := s.Create(model.Product{Name: "Widget", Price: 9.99})
	if err != nil {
		t.Fatalf("expected create to succeed, got error: %v", err)
	}
	if product.ID != 42 {
		t.Fatalf("expected ID 42, got %d", product.ID)
	}
}

func TestPostgresStoreCreateError(t *testing.T) {
	expectedErr := errors.New("insert failed")
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{createErr: expectedErr})}

	if _, err := s.Create(model.Product{Name: "Widget", Price: 9.99}); err != expectedErr {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestPostgresStoreUpdate(t *testing.T) {
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{updateRowsAffected: 1})}

	product, err := s.Update(7, model.Product{Name: "Updated", Price: 14.99})
	if err != nil {
		t.Fatalf("expected update to succeed, got error: %v", err)
	}
	if product.ID != 7 {
		t.Fatalf("expected ID 7, got %d", product.ID)
	}
}

func TestPostgresStoreUpdateError(t *testing.T) {
	expectedErr := errors.New("update failed")
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{updateErr: expectedErr})}

	if _, err := s.Update(7, model.Product{Name: "Updated", Price: 14.99}); err != expectedErr {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestPostgresStoreUpdateNotFound(t *testing.T) {
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{updateRowsAffected: 0})}

	if _, err := s.Update(7, model.Product{Name: "Updated", Price: 14.99}); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresStoreDelete(t *testing.T) {
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{deleteRowsAffected: 1})}

	if err := s.Delete(7); err != nil {
		t.Fatalf("expected delete to succeed, got error: %v", err)
	}
}

func TestPostgresStoreDeleteError(t *testing.T) {
	expectedErr := errors.New("delete failed")
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{deleteErr: expectedErr})}

	if err := s.Delete(7); err != expectedErr {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestPostgresStoreDeleteNotFound(t *testing.T) {
	s := &PostgresStore{DB: openPostgresStoreTestDB(t, &postgresStoreTestDB{deleteRowsAffected: 0})}

	if err := s.Delete(7); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
