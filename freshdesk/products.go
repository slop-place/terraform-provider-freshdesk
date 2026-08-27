package freshdesk

import "context"

// Product is a product configured on the helpdesk. Products are read-only
// through the API; they are managed in the admin console.
type Product struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// PrimaryEmail is the support address associated with the product.
	PrimaryEmail string `json:"primary_email"`
	// Default marks the product new tickets fall back to.
	Default   bool `json:"default"`
	CreatedAt Time `json:"created_at"`
	UpdatedAt Time `json:"updated_at"`
}

// GetProduct fetches a product by ID.
func (c *Client) GetProduct(ctx context.Context, id int64) (*Product, error) {
	return getResource[Product](ctx, c, "products", id, nil)
}

// ListProducts returns every product.
func (c *Client) ListProducts(ctx context.Context, opts ListOptions) ([]Product, error) {
	return listAll[Product](ctx, c, "products", opts)
}
