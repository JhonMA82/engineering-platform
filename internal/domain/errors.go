package domain

import "fmt"

// Class partitions failures into typed buckets.
type Class string

const (
	ClassValidation      Class = "validation"
	ClassCatalog         Class = "catalog"
	ClassResolution      Class = "resolution"
	ClassComposition     Class = "composition"
	ClassMaterialization Class = "materialization"
	ClassFilesystem      Class = "filesystem"
	ClassExternalCommand Class = "external-command"
)

// Error is a typed domain failure.
type Error struct {
	Class   Class  `json:"class"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Class, e.Message)
}

// Validation builds a validation-class error.
func Validation(msg string) *Error {
	return &Error{Class: ClassValidation, Message: msg}
}

// Catalog builds a catalog-class error.
func Catalog(msg string) *Error {
	return &Error{Class: ClassCatalog, Message: msg}
}

// Resolution builds a resolution-class error.
func Resolution(msg string) *Error {
	return &Error{Class: ClassResolution, Message: msg}
}

// Composition builds a composition-class error for composer/planner failures.
func Composition(msg string) *Error {
	return &Error{Class: ClassComposition, Message: msg}
}

// Materialization builds a materialization-class error for materializer
// failures (bad plan, pin mismatch, collision, failed commit).
func Materialization(msg string) *Error {
	return &Error{Class: ClassMaterialization, Message: msg}
}

// Filesystem builds a filesystem-class error for unsafe or failed file
// operations (traversal, symlink escape, unreadable project).
func Filesystem(msg string) *Error {
	return &Error{Class: ClassFilesystem, Message: msg}
}

// ExternalCommand builds an external-command-class error for curated
// adapter command validation and execution failures.
func ExternalCommand(msg string) *Error {
	return &Error{Class: ClassExternalCommand, Message: msg}
}
