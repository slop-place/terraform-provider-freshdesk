package freshdesk

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// pathFor joins a collection path and an identifier.
func pathFor(collection string, id any) string {
	return strings.TrimSuffix(collection, "/") + "/" + idString(id)
}

func idString(id any) string {
	switch v := id.(type) {
	case string:
		return url.PathEscape(v)
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		return url.PathEscape(fmt.Sprint(v))
	}
}

// getResource fetches a single record.
func getResource[T any](ctx context.Context, c *Client, collection string, id any, q url.Values) (*T, error) {
	var out T
	if err := c.Get(ctx, pathFor(collection, id), q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// createResource POSTs body to a collection and decodes the created record.
func createResource[T any](ctx context.Context, c *Client, collection string, body any) (*T, error) {
	var out T
	if err := c.Post(ctx, collection, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// updateResource PUTs body to a record and decodes the updated representation.
func updateResource[T any](ctx context.Context, c *Client, collection string, id any, body any) (*T, error) {
	var out T
	if err := c.Put(ctx, pathFor(collection, id), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// patchResource PATCHes body to a record and decodes the updated
// representation, for the endpoints that reject PUT.
func patchResource[T any](
	ctx context.Context,
	c *Client,
	collection string,
	id any,
	body any,
) (*T, error) {
	var out T
	if err := c.Patch(ctx, pathFor(collection, id), body, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// deleteResource DELETEs a record.
func deleteResource(ctx context.Context, c *Client, collection string, id any) error {
	return c.Delete(ctx, pathFor(collection, id))
}
