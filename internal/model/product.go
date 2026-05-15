package model

// Product represents a product in the catalog.
// Testing the linter: we expect to recieve an error for this typo.
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Validate checks whether the product has valid fields.
func (p *Product) Validate() bool {
	if p.Name == "" {
		return false
	}
	if p.Price < 0 {
		return false
	}
	return true
}
