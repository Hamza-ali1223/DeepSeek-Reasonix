package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"reasonix/internal/safety/sandbox"
)

// A granted write root inside the workspace is ordinary workspace cleanup, not a
// protected target: `rm -rf build` has to run whether build is there or not, and
// `allow_write = ["build"]` must not make the folder undeletable.
func TestAGrantedRootInsideTheWorkspaceIsCleanupNotProtection(t *testing.T) {
	r := newDeleteScopeRig(t)
	granted := filepath.Join(r.root, "later")
	for name, roots := range map[string][]string{
		"workspace first": {r.root, granted},
		"grant first":     {granted, r.root},
	} {
		for _, state := range []string{"missing", "existing"} {
			if state == "existing" {
				if err := os.MkdirAll(granted, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			b := r.b
			b.sb = sandbox.Spec{WriteRoots: roots}
			if _, class := b.classifyDeleteTarget("later", r.root); class != deleteInside {
				t.Errorf("%s, %s: class = %d, want deleteInside", name, state, class)
			}
		}
	}
}

// A granted root outside the workspace is still a target nobody deletes whole.
func TestAGrantedRootOutsideTheWorkspaceStaysProtected(t *testing.T) {
	r := newDeleteScopeRig(t)
	b := r.b
	b.sb = sandbox.Spec{WriteRoots: []string{r.root, r.outside}}
	if _, class := b.classifyDeleteTarget(filepath.ToSlash(r.outside), r.root); class != deleteProtected {
		t.Errorf("class = %d, want deleteProtected", class)
	}
}
