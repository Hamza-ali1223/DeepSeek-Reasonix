package pricing

import "testing"

// Two vendors' deprecation tables chain: Xiaomi retired mimo-v2-pro and
// mimo-v2-omni on 2026-06-30 in favour of V2.5 names, then retired those on
// 2026-10-21. An alias that stops at the first hop resolves to a row nothing
// bills at, so a stored config quotes no rate at all.
func TestMimoAliasesResolveToModelsBillingToday(t *testing.T) {
	cases := map[string]string{
		"mimo-v2-pro":  "mimo-v2.6-pro",
		"mimo-v2-omni": "mimo-v2.6-flash",
	}
	for alias, want := range cases {
		if got := RateModelFor("mimo", alias); got != want {
			t.Errorf("RateModelFor(mimo/%s) = %q, want %q", alias, got, want)
		}
		if rate := CurrentRate("mimo", alias, "CNY"); rate == nil {
			t.Errorf("CurrentRate(mimo/%s CNY) = nil; the alias resolves to nothing billable", alias)
		}
	}
}

// The aliases answer billing, not identity: a name nothing retired must still
// resolve to itself, or every lookup becomes one through the alias table.
func TestUnretiredMimoNamesResolveToThemselves(t *testing.T) {
	for _, model := range []string{"mimo-v2.6-pro", "mimo-v2.6-flash", "mimo-v2.6-pro-ultraspeed"} {
		if got := RateModelFor("mimo", model); got != model {
			t.Errorf("RateModelFor(mimo/%s) = %q, want itself", model, got)
		}
	}
}

// A vendor's alias table says nothing about another vendor's names. DeepSeek
// keeps its own chain, and sharing one would price one vendor's model at
// another's rate.
func TestAliasesStayScopedToTheirVendor(t *testing.T) {
	for _, retired := range RetiredDeepSeekFlashModels() {
		if got := RateModelFor("deepseek", retired); got != "deepseek-flash" {
			t.Errorf("RateModelFor(deepseek/%s) = %q, want deepseek-flash", retired, got)
		}
		if got := RateModelFor("mimo", retired); got != retired {
			t.Errorf("RateModelFor(mimo/%s) = %q; a DeepSeek alias leaked into MiMo", retired, got)
		}
	}
}
