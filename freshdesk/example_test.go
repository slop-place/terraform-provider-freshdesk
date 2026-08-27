package freshdesk_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// The client is usable on its own, without Terraform.
func Example() {
	client, err := freshdesk.New(freshdesk.Config{
		Domain: "acme", // or "acme.freshdesk.com"
		APIKey: os.Getenv("FRESHDESK_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// List helpers follow pagination to the end of the collection.
	groups, err := client.ListGroups(ctx, freshdesk.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, g := range groups {
		fmt.Printf("%d %s\n", g.ID, g.Name)
	}
}

// Optional request fields are pointers, so a partial update sends only what
// was set. Ptr builds them.
func ExampleClient_UpdateGroup() {
	client, err := freshdesk.New(freshdesk.Config{Domain: "acme", APIKey: "key"})
	if err != nil {
		log.Fatal(err)
	}

	// Only the description changes; every other field is left alone.
	_, err = client.UpdateGroup(context.Background(), 42, freshdesk.GroupRequest{
		Description: freshdesk.Ptr("Invoices and refunds"),
	})
	if err != nil {
		log.Fatal(err)
	}
}

// Removing a value needs an explicit instruction, because Freshdesk ignores
// the fields a request omits.
func ExampleGroupRequest() {
	// Detach every agent and clear the escalation target.
	req := freshdesk.GroupRequest{
		Name:            freshdesk.Ptr("Billing"),
		ClearAgents:     true,
		ClearEscalateTo: true,
	}

	fmt.Println(req.ClearAgents, req.ClearEscalateTo)
	// Output: true true
}

// NotFound distinguishes a deleted record from a real failure.
func ExampleNotFound() {
	client, err := freshdesk.New(freshdesk.Config{Domain: "acme", APIKey: "key"})
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.GetGroup(context.Background(), 42)

	switch {
	case err == nil:
		fmt.Println("found")
	case freshdesk.NotFound(err):
		fmt.Println("the group has been deleted")
	default:
		log.Fatal(err)
	}
}

// Do reaches an endpoint the typed methods do not model.
func ExampleClient_Do() {
	client, err := freshdesk.New(freshdesk.Config{Domain: "acme", APIKey: "key"})
	if err != nil {
		log.Fatal(err)
	}

	var out map[string]any

	_, err = client.Do(context.Background(), freshdesk.Request{
		Method: "GET",
		Path:   "some/new/endpoint",
	}, &out)
	if err != nil {
		log.Fatal(err)
	}
}
