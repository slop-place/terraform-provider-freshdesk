package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/slop-place/terraform-provider-freshdesk/freshdesk"
)

// Acceptance tests run against a real Freshdesk account. They are skipped
// unless TF_ACC is set, and they create and destroy real records, so point
// them at a sandbox.
//
//	TF_ACC=1 FRESHDESK_DOMAIN=acme FRESHDESK_API_KEY=... go test ./internal/provider/ -v

// testAccProtoV6Providers wires the provider under test into the harness.
//
//nolint:gochecknoglobals // the harness requires a package-level factory
var testAccProtoV6Providers = map[string]func() (tfprotov6.ProviderServer, error){
	"freshdesk": providerserver.NewProtocol6WithError(New("acc")()),
}

// ErrStillExists reports a record that outlived its Terraform resource.
var ErrStillExists = errors.New("record still exists after destroy")

// testAccPreCheck fails fast when the environment is not configured, so a
// missing variable is reported once rather than as a wall of provider errors.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	for _, key := range []string{"FRESHDESK_API_KEY"} {
		if os.Getenv(key) == "" {
			t.Fatalf("%s must be set for acceptance tests", key)
		}
	}

	if os.Getenv("FRESHDESK_DOMAIN") == "" && os.Getenv("FRESHDESK_URL") == "" {
		t.Fatal("FRESHDESK_DOMAIN must be set for acceptance tests")
	}
}

// skipUnlessAcceptance skips a test that touches the API before
// resource.Test's own TF_ACC gate would.
func skipUnlessAcceptance(t *testing.T) {
	t.Helper()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping the acceptance test")
	}
}

// testAccClient builds a client for assertions the harness cannot make.
func testAccClient(t *testing.T) *freshdesk.Client {
	t.Helper()
	skipUnlessAcceptance(t)

	domain := os.Getenv("FRESHDESK_DOMAIN")
	if domain == "" {
		domain = os.Getenv("FRESHDESK_URL")
	}

	c, err := freshdesk.New(freshdesk.Config{
		Domain: domain,
		APIKey: os.Getenv("FRESHDESK_API_KEY"),
	})
	if err != nil {
		t.Fatalf("building the acceptance client: %v", err)
	}

	return c
}

// accRunID distinguishes one acceptance run from the next, so a record left
// behind by an interrupted run cannot collide with a fresh one on the unique
// fields Freshdesk enforces (a company name, a contact's email).
//
//nolint:gochecknoglobals // one identifier shared by the whole run
var accRunID = strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 36)

// accName builds a collision-proof name for a record created by a test. The
// tfacc- prefix makes leftovers easy to spot and remove; see
// internal/apispec/cleanup.sh.
func accName(t *testing.T, suffix string) string {
	t.Helper()

	name := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	return fmt.Sprintf("tfacc-%s-%s-%s", name, accRunID, suffix)
}

// checkDestroyed returns a CheckDestroy that asserts the record is gone,
// using getFn to look it up directly.
func checkDestroyed(
	t *testing.T,
	resourceType string,
	getFn func(context.Context, *freshdesk.Client, int64) error,
) resource.TestCheckFunc {
	t.Helper()

	return func(state *terraform.State) error {
		client := testAccClient(t)

		for name, rs := range state.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}

			id, err := parseID(rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}

			err = getFn(context.Background(), client, id)
			if err == nil {
				return fmt.Errorf("%w: %s (%s)", ErrStillExists, name, rs.Primary.ID)
			}

			if !freshdesk.NotFound(err) {
				return fmt.Errorf("%s: unexpected error checking destroy: %w", name, err)
			}
		}

		return nil
	}
}
