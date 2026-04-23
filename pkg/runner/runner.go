// Package runner is a thin re-export of the stack/unit discovery API that
// lives in internal/runner. It exists so external library consumers (terracost
// and friends) can discover a Terragrunt stack, enumerate its units, and drive
// their execution without importing packages under internal/.
//
// This is intentionally a narrow surface: only the pieces needed by downstream
// Cycloid tooling are exposed. Keep it minimal so upstream refactors inside
// internal/runner/** stay absorbable by translating here rather than rippling
// into every consumer.
package runner

import (
	"context"

	"github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/internal/component"
	"github.com/gruntwork-io/terragrunt/internal/runner"
	"github.com/gruntwork-io/terragrunt/internal/runner/common"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/pkg/log"
)

// Stack is a collection of Terragrunt units discovered under a working
// directory, together with a runner capable of executing them.
type Stack struct {
	inner common.StackRunner
}

// TerraformModule is a single Terragrunt unit inside a Stack. The name mirrors
// the pre-upstream-refactor configstack.TerraformModule type so existing
// consumers need only minor rewrites after the internal-package rename.
type TerraformModule struct {
	unit *component.Unit
}

// FindStackInSubfolders discovers every Terragrunt unit under opts.WorkingDir
// and returns a Stack that can be introspected or executed.
//
// After this call each TerraformModule has its per-unit TerragruntOptions
// populated, so callers can read m.TerragruntOptions() and m.Config() without
// having to trigger execution first.
func FindStackInSubfolders(ctx context.Context, l log.Logger, opts *options.TerragruntOptions) (*Stack, error) {
	inner, err := runner.FindStackInSubfolders(ctx, l, opts)
	if err != nil {
		return nil, err
	}

	return &Stack{inner: inner}, nil
}

// Run executes the stack using the current TerraformCommand set on opts.
func (s *Stack) Run(ctx context.Context, l log.Logger, opts *options.TerragruntOptions) error {
	return s.inner.Run(ctx, l, opts)
}

// Modules returns every unit in the stack.
//
// Units that were skipped during discovery (no .tf files, no source) are
// filtered out by the runner before this is populated, so the returned slice
// reflects only the units that would actually be invoked by Run.
func (s *Stack) Modules() []*TerraformModule {
	stack := s.inner.GetStack()
	if stack == nil {
		return nil
	}

	mods := make([]*TerraformModule, 0, len(stack.Units))
	for _, u := range stack.Units {
		if u == nil {
			continue
		}

		mods = append(mods, &TerraformModule{unit: u})
	}

	return mods
}

// Path returns the filesystem path of the unit.
func (m *TerraformModule) Path() string {
	return m.unit.Path()
}

// Config returns the parsed Terragrunt configuration for the unit.
// May be nil if discovery did not successfully parse the unit's config.
func (m *TerraformModule) Config() *config.TerragruntConfig {
	return m.unit.Config()
}

// TerragruntOptions returns the per-unit TerragruntOptions that the runner
// would use when executing this unit. May be nil if the unit was created
// outside the resolve step (e.g. in an empty stack), but for stacks returned
// from FindStackInSubfolders this is always set.
func (m *TerraformModule) TerragruntOptions() *options.TerragruntOptions {
	if m.unit.Execution == nil {
		return nil
	}

	return m.unit.Execution.TerragruntOptions
}
