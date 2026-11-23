package pricing

import "testing"

func TestMimoV26RatesFollowTheVendorPage(t *testing.T) {
	cases := []struct {
		model          string
		hit, miss, out float64
		v25Twin        string
	}{
		{"mimo-v2.6-pro", 0.025, 3, 6, "mimo-v2.5-pro"},
		{"mimo-v2.6-flash", 0.02, 1, 2, "mimo-v2.5"},
		{"mimo-v2.6-pro-ultraspeed", 0.25, 30, 60, ""},
	}
	for _, tc := range cases {
		got := CurrentRate("mimo", tc.model, "CNY")
		if got == nil || got.CacheHit != tc.hit || got.Input != tc.miss || got.Output != tc.out {
			t.Fatalf("%s rate = %+v, want %v/%v/%v", tc.model, got, tc.hit, tc.miss, tc.out)
		}
		if tc.v25Twin == "" {
			continue
		}
		old, ok := LookupCatalog("mimo", tc.v25Twin, "CNY", BillingModePAYG)
		// The twin is retired, so it is no longer a current rate; its published
		// numbers stay in the catalog because usage was billed at them.
		if !ok || old.CacheHit != got.CacheHit || old.Input != got.Input || old.Output != got.Output {
			t.Fatalf("%s must keep its published rate beside %s: %+v vs %+v", tc.v25Twin, tc.model, old, got)
		}
	}
}
