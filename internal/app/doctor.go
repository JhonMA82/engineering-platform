package app

import (
	"github.com/jhonma82/engineering-platform/internal/project"
)

// DoctorProject checks the consistency of a materialized project directory.
func DoctorProject(projectDir string) ([]project.Finding, error) {
	return project.Doctor(projectDir)
}
