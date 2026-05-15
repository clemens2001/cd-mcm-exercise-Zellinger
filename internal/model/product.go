package model

// Product represents a product in the catalog.
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
