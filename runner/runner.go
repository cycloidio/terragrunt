// Package runner provides public access to the internal runner functionality.
package runner

import (
	"context"

	"github.com/gruntwork-io/terragrunt/internal/runner"
	"github.com/gruntwork-io/terragrunt/internal/runner/common"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/pkg/log"
)

// StackRunner is the abstraction for running a stack of units.
type StackRunner = common.StackRunner

// Option is a configuration option for stack operations.
type Option = common.Option

// FindStackInSubfolders finds all the Terraform modules in the subfolders of the working directory
// and assembles them into a StackRunner that can be applied or destroyed in a single command.
func FindStackInSubfolders(ctx context.Context, l log.Logger, terragruntOptions *options.TerragruntOptions, opts ...Option) (StackRunner, error) {
	return runner.FindStackInSubfolders(ctx, l, terragruntOptions, opts...)
}
