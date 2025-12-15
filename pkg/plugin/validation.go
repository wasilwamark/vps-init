package plugin

import (
	"fmt"
	"regexp"
	"strings"
	"runtime"

	"github.com/Masterminds/semver/v3"
)

// ValidationError represents a plugin validation error
type ValidationError struct {
	Field   string
	Message string
	Code    string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s: %s (%s)", e.Field, e.Message, e.Code)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}

	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validator provides plugin validation functionality
type Validator struct {
	vpsInitVersion string
}

// NewValidator creates a new plugin validator
func NewValidator(vpsInitVersion string) *Validator {
	return &Validator{
		vpsInitVersion: vpsInitVersion,
	}
}

// ValidatePlugin performs comprehensive plugin validation
func (v *Validator) ValidatePlugin(plugin Plugin) ValidationErrors {
	var errors ValidationErrors

	// Validate basic metadata
	errors = append(errors, v.validateMetadata(plugin)...)

	// Validate version
	errors = append(errors, v.validateVersion(plugin)...)

	// Validate dependencies
	errors = append(errors, v.validateDependencies(plugin)...)

	// Validate compatibility
	errors = append(errors, v.validateCompatibility(plugin)...)

	// Validate plugin structure
	errors = append(errors, v.validateStructure(plugin)...)

	// Run plugin-specific validation
	if err := plugin.Validate(); err != nil {
		errors = append(errors, ValidationError{
			Field:   "plugin",
			Message: fmt.Sprintf("plugin validation failed: %v", err),
			Code:    "PLUGIN_VALIDATION_FAILED",
		})
	}

	return errors
}

// validateMetadata validates plugin metadata
func (v *Validator) validateMetadata(plugin Plugin) ValidationErrors {
	var errors ValidationErrors

	name := plugin.Name()
	description := plugin.Description()
	author := plugin.Author()

	// Name validation
	if name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "plugin name cannot be empty",
			Code:    "EMPTY_NAME",
		})
	} else {
		// Name should contain only lowercase letters, numbers, and hyphens
		if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(name) {
			errors = append(errors, ValidationError{
				Field:   "name",
				Message: "plugin name must contain only lowercase letters, numbers, and hyphens",
				Code:    "INVALID_NAME_FORMAT",
			})
		}

		// Name length
		if len(name) > 50 {
			errors = append(errors, ValidationError{
				Field:   "name",
				Message: "plugin name cannot exceed 50 characters",
				Code:    "NAME_TOO_LONG",
			})
		}
	}

	// Description validation
	if description == "" {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "plugin description cannot be empty",
			Code:    "EMPTY_DESCRIPTION",
		})
	} else if len(description) > 500 {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "plugin description cannot exceed 500 characters",
			Code:    "DESCRIPTION_TOO_LONG",
		})
	}

	// Author validation
	if author == "" {
		errors = append(errors, ValidationError{
			Field:   "author",
			Message: "plugin author cannot be empty",
			Code:    "EMPTY_AUTHOR",
		})
	}

	return errors
}

// validateVersion validates plugin version using semantic versioning
func (v *Validator) validateVersion(plugin Plugin) ValidationErrors {
	var errors ValidationErrors

	version := plugin.Version()
	if version == "" {
		errors = append(errors, ValidationError{
			Field:   "version",
			Message: "plugin version cannot be empty",
			Code:    "EMPTY_VERSION",
		})
		return errors
	}

	// Validate semantic version
	if _, err := semver.NewVersion(version); err != nil {
		errors = append(errors, ValidationError{
			Field:   "version",
			Message: fmt.Sprintf("invalid semantic version: %v", err),
			Code:    "INVALID_SEMVER",
		})
		return errors
	}

	return errors
}

// validateDependencies validates plugin dependencies
func (v *Validator) validateDependencies(plugin Plugin) ValidationErrors {
	var errors ValidationErrors

	dependencies := plugin.Dependencies()

	// Check for circular dependencies would be done at the registry level
	for _, dep := range dependencies {
		// Validate dependency name
		if dep.Name == "" {
			errors = append(errors, ValidationError{
				Field:   "dependencies",
				Message: "dependency name cannot be empty",
				Code:    "EMPTY_DEPENDENCY_NAME",
			})
			continue
		}

		// Validate dependency version if specified
		if dep.Version != "" {
			if _, err := semver.NewConstraint(dep.Version); err != nil {
				errors = append(errors, ValidationError{
					Field:   "dependencies",
					Message: fmt.Sprintf("invalid version constraint for dependency %s: %v", dep.Name, err),
					Code:    "INVALID_DEPENDENCY_VERSION",
				})
			}
		}
	}

	return errors
}

// validateCompatibility validates plugin compatibility requirements
func (v *Validator) validateCompatibility(plugin Plugin) ValidationErrors {
	var errors ValidationErrors

	compat := plugin.Compatibility()

	// Validate VPS-Init version compatibility
	if compat.MinVPSInitVersion != "" {
		currentVersion, err := semver.NewVersion(v.vpsInitVersion)
		if err != nil {
			// If we can't parse current version, skip this check
			return errors
		}

		minVersion, err := semver.NewConstraint(compat.MinVPSInitVersion)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:   "compatibility.min_vps_init_version",
				Message: fmt.Sprintf("invalid minimum VPS-Init version constraint: %v", err),
				Code:    "INVALID_MIN_VERSION",
			})
		} else if !minVersion.Check(currentVersion) {
			errors = append(errors, ValidationError{
				Field:   "compatibility",
				Message: fmt.Sprintf("plugin requires VPS-Init version %s, but current version is %s",
					compat.MinVPSInitVersion, v.vpsInitVersion),
				Code:    "INCOMPATIBLE_VPS_INIT_VERSION",
			})
		}
	}

	// Validate maximum VPS-Init version if specified
	if compat.MaxVPSInitVersion != "" {
		currentVersion, err := semver.NewVersion(v.vpsInitVersion)
		if err != nil {
			return errors
		}

		maxVersion, err := semver.NewConstraint("<=" + compat.MaxVPSInitVersion)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:   "compatibility.max_vps_init_version",
				Message: fmt.Sprintf("invalid maximum VPS-Init version constraint: %v", err),
				Code:    "INVALID_MAX_VERSION",
			})
		} else if !maxVersion.Check(currentVersion) {
			errors = append(errors, ValidationError{
				Field:   "compatibility",
				Message: fmt.Sprintf("plugin requires VPS-Init version <= %s, but current version is %s",
					compat.MaxVPSInitVersion, v.vpsInitVersion),
				Code:    "INCOMPATIBLE_VPS_INIT_VERSION",
			})
		}
	}

	// Validate platform compatibility
	if len(compat.Platforms) > 0 {
		currentPlatform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
		platformCompatible := false

		for _, platform := range compat.Platforms {
			if platform == currentPlatform {
				platformCompatible = true
				break
			}
		}

		if !platformCompatible {
			errors = append(errors, ValidationError{
				Field:   "compatibility.platforms",
				Message: fmt.Sprintf("plugin is not compatible with current platform %s", currentPlatform),
				Code:    "INCOMPATIBLE_PLATFORM",
			})
		}
	}

	return errors
}

// validateStructure validates plugin structure and commands
func (v *Validator) validateStructure(plugin Plugin) ValidationErrors {
	var errors ValidationErrors

	// Validate commands
	commands := plugin.GetCommands()
	if len(commands) == 0 {
		errors = append(errors, ValidationError{
			Field:   "commands",
			Message: "plugin must have at least one command",
			Code:    "NO_COMMANDS",
		})
		return errors
	}

	commandNames := make(map[string]bool)
	for i, cmd := range commands {
		// Validate command name
		if cmd.Name == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("commands[%d].name", i),
				Message: "command name cannot be empty",
				Code:    "EMPTY_COMMAND_NAME",
			})
			continue
		}

		// Check for duplicate command names
		if commandNames[cmd.Name] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("commands[%d].name", i),
				Message: fmt.Sprintf("duplicate command name: %s", cmd.Name),
				Code:    "DUPLICATE_COMMAND_NAME",
			})
		}
		commandNames[cmd.Name] = true

		// Validate command name format
		if !regexp.MustCompile(`^[a-z][a-z0-9-]*$`).MatchString(cmd.Name) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("commands[%d].name", i),
				Message: "command name must start with a letter and contain only lowercase letters, numbers, and hyphens",
				Code:    "INVALID_COMMAND_NAME",
			})
		}

		// Validate command description
		if cmd.Description == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("commands[%d].description", i),
				Message: "command description cannot be empty",
				Code:    "EMPTY_COMMAND_DESCRIPTION",
			})
		}

		// Validate command handler
		if cmd.Handler == nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("commands[%d].handler", i),
				Message: "command handler cannot be nil",
				Code:    "NIL_COMMAND_HANDLER",
			})
		}
	}

	return errors
}

// ValidatePluginMetadata validates plugin metadata struct
func (v *Validator) ValidatePluginMetadata(metadata PluginMetadata) ValidationErrors {
	var errors ValidationErrors

	// Basic field validation
	if metadata.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "metadata.name",
			Message: "name cannot be empty",
			Code:    "EMPTY_NAME",
		})
	}

	if metadata.Description == "" {
		errors = append(errors, ValidationError{
			Field:   "metadata.description",
			Message: "description cannot be empty",
			Code:    "EMPTY_DESCRIPTION",
		})
	}

	if metadata.Version == "" {
		errors = append(errors, ValidationError{
			Field:   "metadata.version",
			Message: "version cannot be empty",
			Code:    "EMPTY_VERSION",
		})
	} else if _, err := semver.NewVersion(metadata.Version); err != nil {
		errors = append(errors, ValidationError{
			Field:   "metadata.version",
			Message: fmt.Sprintf("invalid semantic version: %v", err),
			Code:    "INVALID_SEMVER",
		})
	}

	// Validate checksum format if present
	if metadata.Checksum != "" {
		if !regexp.MustCompile(`^[a-fA-F0-9]{64}$`).MatchString(metadata.Checksum) {
			errors = append(errors, ValidationError{
				Field:   "metadata.checksum",
				Message: "checksum must be a valid SHA256 hash",
				Code:    "INVALID_CHECKSUM",
			})
		}
	}

	// Validate trust level
	if metadata.TrustLevel != "" {
		validTrustLevels := []string{"official", "community", "verified", "untrusted"}
		valid := false
		for _, level := range validTrustLevels {
			if metadata.TrustLevel == level {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, ValidationError{
				Field:   "metadata.trust_level",
				Message: fmt.Sprintf("trust level must be one of: %s", strings.Join(validTrustLevels, ", ")),
				Code:    "INVALID_TRUST_LEVEL",
			})
		}
	}

	return errors
}

// GetTrustLevel determines the trust level for a plugin
func GetTrustLevel(metadata PluginMetadata) string {
	if metadata.TrustLevel != "" {
		return metadata.TrustLevel
	}

	// Auto-determine trust level based on repository
	if strings.Contains(metadata.Repository, "github.com/vps-init") ||
		strings.Contains(metadata.Repository, "github.com/wasilwamark/vps-init") {
		return "official"
	}

	if metadata.Repository != "" {
		return "community"
	}

	return "untrusted"
}