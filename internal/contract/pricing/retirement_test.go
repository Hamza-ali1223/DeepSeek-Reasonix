package pricing

import (
	"testing"
	"time"
)

// Xiaomi's deprecation page states both dates: v2.5-pro and v2.5 died at
// 10:00 UTC+8 on 2026-10-21, and v2-flash at 00:00 UTC+8 on 2026-06-30.
// A retired model must not answer as a current rate — nothing bills at it —
// while the historical rows stay in the catalog so old usage still values.
func TestRetiredModelsAreNotCurrentRates(t *testing.T) {
	retired := []struct{ model, currency string }{
		{"mimo-v2.5-pro", "CNY"},
		{"mimo-v2.5", "CNY"},
		{"mimo-v2-flash", "CNY"},
	}
	for _, r := range retired {
		if got := CurrentRate("mimo", r.model, r.currency); got != nil {
			t.Errorf("CurrentRate(mimo/%s %s) = %+v, want nil for a retired model", r.model, r.currency, got)
		}
	}
}

// Superseding a model is not retiring it: the live rows must keep answering, or
// the fix for one vendor quietly empties the rest of the table.
func TestLiveModelsAreStillCurrentRates(t *testing.T) {
	live := []string{"mimo-v2.6-pro", "mimo-v2.6-flash", "mimo-v2.6-pro-ultraspeed"}
	for _, model := range live {
		if got := CurrentRate("mimo", model, "CNY"); got == nil {
			t.Errorf("CurrentRate(mimo/%s CNY) = nil, want the published rate", model)
		}
	}
	for _, p := range []struct{ provider, model, currency string }{
		{"longcat", "LongCat-2.0", "CNY"},
		{"longcat", "LongCat-2.0", "USD"},
		{"deepseek", "deepseek-flash", "CNY"},
		{"google", "gemini-3.8-flash", "USD"},
	} {
		if got := CurrentRate(p.provider, p.model, p.currency); got == nil {
			t.Errorf("CurrentRate(%s/%s %s) = nil, want the published rate", p.provider, p.model, p.currency)
		}
	}
}

// A retired row keeps its place in the catalog with a bounded lifetime, so a
// caller can tell "no longer offered" from "never priced".
func TestRetiredRowsPublishTheirRetirementDate(t *testing.T) {
	want := map[string]string{
		"mimo-v2.5-pro": "2026-10-21",
		"mimo-v2.5":     "2026-10-21",
		"mimo-v2-flash": "2026-06-30",
	}
	seen := map[string]bool{}
	for _, entry := range OfficialCatalog() {
		if entry.Provider != "mimo" {
			continue
		}
		expected, ok := want[entry.Model]
		if !ok {
			if entry.EffectiveTo != "" {
				t.Errorf("mimo/%s %s is live but publishes EffectiveTo %q", entry.Model, entry.BillingMode, entry.EffectiveTo)
			}
			continue
		}
		seen[entry.Model] = true
		if entry.EffectiveTo != expected {
			t.Errorf("mimo/%s %s EffectiveTo = %q, want %q", entry.Model, entry.BillingMode, entry.EffectiveTo, expected)
		}
		if entry.InEffect(time.Now().UTC()) {
			t.Errorf("mimo/%s %s is still in effect on its retirement date", entry.Model, entry.BillingMode)
		}
	}
	for model := range want {
		if !seen[model] {
			t.Errorf("mimo/%s is published by no catalog row; old usage would lose its official basis", model)
		}
	}
}

// Retirement withdraws a row from being offered, never from being recognised.
// Historical valuation calls RatesAt on whatever the catalog still knows, so a
// retired row that stopped matching would strip the official basis from every
// old usage record for that model.
func TestRetiredRowsStillMatchTheCatalog(t *testing.T) {
	for _, entry := range OfficialCatalog() {
		if entry.EffectiveTo == "" {
			continue
		}
		card := RateCardFromCatalog(entry)
		if _, ok := MatchesCatalog(entry.Provider, entry.Model, card); !ok {
			t.Errorf("%s/%s %s is retired but no longer matches the catalog; old usage would lose its price", entry.Provider, entry.Model, entry.BillingMode)
		}
		if _, ok := LookupCatalog(entry.Provider, entry.Model, entry.Currency, entry.BillingMode); !ok {
			t.Errorf("%s/%s %s is retired but LookupCatalog no longer finds it", entry.Provider, entry.Model, entry.BillingMode)
		}
	}
}

// Retirement bounds a row's lifetime; it does not excuse it from the freshness
// guard's other half, or a row could be retired to escape ever being re-read.
func TestRetiredRowsStillCarryACheckedOnDate(t *testing.T) {
	for _, entry := range OfficialCatalog() {
		if entry.EffectiveTo == "" {
			continue
		}
		if entry.CheckedOn == "" {
			t.Errorf("%s/%s %s is retired with no CheckedOn", entry.Provider, entry.Model, entry.Currency)
			continue
		}
		if _, err := time.Parse("2006-01-02", entry.CheckedOn); err != nil {
			t.Errorf("%s/%s %s CheckedOn %q is not YYYY-MM-DD", entry.Provider, entry.Model, entry.Currency, entry.CheckedOn)
		}
	}
}
