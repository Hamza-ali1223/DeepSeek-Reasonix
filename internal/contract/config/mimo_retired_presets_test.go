package config

import "testing"

// intermediateMimoShape is the shape shipped between the V2.6 catalog and the
// 2026-10-21 retirement: the live V2.6 models listed beside the retired V2.5
// names, at the current window, with only the retired model marked for images.
func intermediateMimoShape(t *testing.T, id string) ProviderEntry {
	t.Helper()
	e := mimoPresetEntry(t, id)
	payg := id == "mimo-api" || id == "mimo-anthropic"
	if payg {
		e.Models = []string{"mimo-v2.6-pro", "mimo-v2.6-flash", "mimo-v2.6-pro-ultraspeed", "mimo-v2.5-pro", "mimo-v2.5"}
	} else {
		e.Models = []string{"mimo-v2.6-pro", "mimo-v2.6-flash", "mimo-v2.5-pro", "mimo-v2.5"}
	}
	e.VisionModels = []string{"mimo-v2.5"}
	e.Default = "mimo-v2.6-pro"
	e.ContextWindow = mimoContextWindow
	e.Prices = mimoDomesticPrices(e.Models)
	return e
}

// Xiaomi retired mimo-v2.5-pro and mimo-v2.5 on 2026-10-21 and listed only the
// V2.6 series as current. A preset that still offers a retired id hands the
// user a choice that answers with an API error.
func TestMimoPresetsOfferNoRetiredModels(t *testing.T) {
	retired := map[string]bool{
		"mimo-v2.5": true, "mimo-v2.5-pro": true,
		"mimo-v2-pro": true, "mimo-v2-omni": true, "mimo-v2-flash": true,
	}
	for _, id := range mimoPresetIDs {
		e := mimoPresetEntry(t, id)
		for _, model := range e.Models {
			if retired[model] {
				t.Errorf("%s offers retired model %s", id, model)
			}
		}
		if retired[e.Default] {
			t.Errorf("%s defaults to retired model %s", id, e.Default)
		}
		for _, model := range e.VisionModels {
			if retired[model] {
				t.Errorf("%s declares retired model %s image-taking", id, model)
			}
		}
	}
}

// The vendor's Image Understanding page names all three V2.6 models as
// supported, so a preset that denies them leaves a working capability switched
// off. TestIsLikelyVisionModel carries the same claim for undeclared addresses.
func TestMimoPresetsDeclareV26Vision(t *testing.T) {
	for _, id := range mimoPresetIDs {
		e := mimoPresetEntry(t, id)
		for _, model := range e.Models {
			if !e.HasVisionModel(model) {
				t.Errorf("%s lists %s but does not mark it image-taking", id, model)
			}
		}
	}
	for _, model := range []string{"mimo-v2.6-pro", "mimo-v2.6-flash", "mimo-v2.6-pro-ultraspeed"} {
		if !IsLikelyVisionModel(model) {
			t.Errorf("IsLikelyVisionModel(%q) = false; the vendor states it reads images", model)
		}
	}
}

// Declaring image support for a model the preset does not list sends images to
// something the user cannot select, which is the shape of a typo rather than a
// capability.
func TestMimoPresetVisionStaysWithinItsModelList(t *testing.T) {
	for _, id := range mimoPresetIDs {
		e := mimoPresetEntry(t, id)
		for _, model := range e.VisionModels {
			if !e.HasModel(model) {
				t.Errorf("%s marks %s image-taking but does not list it", id, model)
			}
		}
	}
}

// An install still holding the pre-V2.6 shape is ours to bring forward: exact
// catalog, window and default, so curation is what the match excludes.
func TestLegacyMimoShapeUpgradesToTheCurrentCatalog(t *testing.T) {
	for _, id := range mimoPresetIDs {
		legacy := legacyMimoShape(t, id)
		c := &Config{Providers: []ProviderEntry{legacy}}
		if !upgradeShippedPresets(c) {
			t.Errorf("%s legacy shape did not migrate", id)
			continue
		}
		canonical := mimoPresetEntry(t, id)
		got := c.Providers[0]
		if !stringSlicesEqual(got.Models, canonical.Models) {
			t.Errorf("%s models = %v, want %v", id, got.Models, canonical.Models)
		}
		if got.Default != canonical.Default {
			t.Errorf("%s default = %q, want %q", id, got.Default, canonical.Default)
		}
		for _, model := range got.Models {
			if !got.HasVisionModel(model) {
				t.Errorf("%s migrated but %s is not image-taking", id, model)
			}
		}
	}
}

// The shape shipped between V2.6 and the retirement lists the retired names
// beside the live ones. It is an exact shape too, so it must move forward — and
// because it adds no models, its vision list has to be stated by the upgrade
// rather than left waiting for `added`.
func TestIntermediateMimoShapeUpgradesToTheCurrentCatalog(t *testing.T) {
	for _, id := range mimoPresetIDs {
		c := &Config{Providers: []ProviderEntry{intermediateMimoShape(t, id)}}
		if !upgradeShippedPresets(c) {
			t.Errorf("%s intermediate shape did not migrate", id)
			continue
		}
		canonical := mimoPresetEntry(t, id)
		got := c.Providers[0]
		if !stringSlicesEqual(got.Models, canonical.Models) {
			t.Errorf("%s models = %v, want %v", id, got.Models, canonical.Models)
		}
		for _, model := range got.Models {
			if !got.HasVisionModel(model) {
				t.Errorf("%s migrated but %s is not image-taking", id, model)
			}
		}
	}
}

// Curation is what an upgrade must never overwrite. A changed model list is a
// choice, and so is a vision list the project never shipped.
func TestMimoUpgradeLeavesCuratedEntriesAlone(t *testing.T) {
	for _, id := range mimoPresetIDs {
		curatedModels := legacyMimoShape(t, id)
		curatedModels.Models = []string{"mimo-v2.5-pro"}
		curatedModels.Default = "mimo-v2.5-pro"
		curatedModels.VisionModels = []string{"mimo-v2.5-pro"}

		// nil is "never said", which the upgrade is allowed to fill; an explicit
		// empty list is the opposite claim and must survive untouched.
		unstatedVision := legacyMimoShape(t, id)
		unstatedVision.VisionModels = nil

		emptyVision := legacyMimoShape(t, id)
		emptyVision.VisionModels = []string{}

		c := &Config{Providers: []ProviderEntry{curatedModels, unstatedVision, emptyVision}}
		upgradeShippedPresets(c)

		if !stringSlicesEqual(c.Providers[0].Models, []string{"mimo-v2.5-pro"}) {
			t.Errorf("%s: an upgrade replaced a curated model list", id)
		}
		if got := c.Providers[0].VisionModels; !stringSlicesEqual(got, []string{"mimo-v2.5-pro"}) {
			t.Errorf("%s: an upgrade replaced a curated vision list, got %v", id, got)
		}
		canonical := mimoPresetEntry(t, id)
		if !stringSlicesEqual(c.Providers[1].VisionModels, canonical.VisionModels) {
			t.Errorf("%s: an undeclared vision list was not filled from the preset, got %v", id, c.Providers[1].VisionModels)
		}
		if c.Providers[2].VisionModels == nil || len(c.Providers[2].VisionModels) != 0 {
			t.Errorf("%s: an upgrade overwrote an explicit empty vision list, got %v", id, c.Providers[2].VisionModels)
		}
	}
}
