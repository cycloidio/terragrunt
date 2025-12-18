// Package component provides public access to the internal component types.
package component

import (
	"github.com/gruntwork-io/terragrunt/internal/component"
)

// Stack represents a discovered Terragrunt stack configuration.
type Stack = component.Stack

// Unit represents a discovered Terragrunt unit configuration.
type Unit = component.Unit

// UnitExecution holds execution-specific fields for running a unit.
type UnitExecution = component.UnitExecution

// Component represents a discovered Terragrunt configuration.
type Component = component.Component

// Components is a list of discovered Terragrunt components.
type Components = component.Components

// Kind is the type of Terragrunt component.
type Kind = component.Kind

// Constants for component kinds.
const (
	UnitKind  = component.UnitKind
	StackKind = component.StackKind
)

// NewUnit creates a new Unit component with the given path.
func NewUnit(path string) *Unit {
	return component.NewUnit(path)
}

// NewStack creates a new Stack component with the given path.
func NewStack(path string) *Stack {
	return component.NewStack(path)
}
