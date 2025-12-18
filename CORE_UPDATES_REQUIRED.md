# VPS-Init Core Package Updates Required

This document identifies what needs to be added to `github.com/wasilwamark/vps-init-core` to fully support VPS-Init's plugin system.

## Current State

### Available in vps-init-core ✅
- `core.Plugin` interface
- `core.PluginMetadata` struct
- `core.Dependency` struct
- `core.Compatibility` struct
- `core.Command` struct
- `core.Argument` struct
- `core.Flag` struct
- `core.CommandHandler` type
- `core.ArgumentType` enum

### Missing from vps-init-core ❌

The following types and fields that VPS-Init currently uses are **NOT** available in vps-init-core:

## 1. PluginMetadata Missing Fields

VPS-Init's PluginMetadata includes these fields that are missing from core:

```go
type PluginMetadata struct {
    // Fields in core but used differently:
    core.PluginMetadata

    // Missing fields needed by VPS-Init:
    Validated         bool      `json:"validated"`
    ValidationErrors  []string  `json:"validation_errors,omitempty"`
    Signature         string    `json:"signature,omitempty"`     // GPG signature
    TrustLevel        string    `json:"trust_level,omitempty"`  // official, community, untrusted
    Checksum          string    `json:"checksum,omitempty"`
    Source            string    `json:"source,omitempty"`        // git URL, package path
    BuildInfo         BuildInfo `json:"build_info,omitempty"`
}
```

## 2. BuildInfo Struct (Missing)

```go
type BuildInfo struct {
    GoVersion    string   `json:"go_version,omitempty"`
    BuildTime    string   `json:"build_time,omitempty"`
    GitCommit    string   `json:"git_commit,omitempty"`
    GitTag       string   `json:"git_tag,omitempty"`
    BuildFlags   []string `json:"build_flags,omitempty"`
    Dependencies []string `json:"dependencies,omitempty"`
}
```

## 3. ValidationError Struct (Missing)

```go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Code    string `json:"code"`
}
```

## 4. Additional Constants (Missing)

```go
const (
    ArgumentTypeString ArgumentType = iota
    ArgumentTypeInt
    ArgumentTypeBool
    ArgumentTypeSlice
)
```

## 5. Additional Utility Functions (Missing)

```go
// GetTrustLevel determines the trust level for a plugin
func GetTrustLevel(metadata PluginMetadata) string

// ValidatePluginMetadata validates plugin metadata
func ValidatePluginMetadata(metadata PluginMetadata) []ValidationError
```

## Recommended Action Plan

### Phase 1: Core Updates (Required)

1. **Update PluginMetadata in vps-init-core**:
   - Add missing fields to the existing PluginMetadata struct
   - Maintain backward compatibility

2. **Add BuildInfo struct**:
   - Add the BuildInfo struct definition

3. **Add ValidationError struct**:
   - Add the ValidationError struct for validation

4. **Add utility functions**:
   - Add GetTrustLevel function
   - Add ValidatePluginMetadata function

### Phase 2: VPS-Init Updates (After Core Updates)

1. **Update VPS-Init to use core types**:
   - Remove local type definitions
   - Use core.Plugin directly
   - Update all plugin implementations to use core types

## Implementation Priority

**High Priority** (Required for basic functionality):
- PluginMetadata missing fields
- BuildInfo struct

**Medium Priority** (Required for plugin installation):
- ValidationError struct
- ValidatePluginMetadata function
- GetTrustLevel function

**Low Priority** (Can be deferred):
- Additional ArgumentType constants (if not already in core)

## Current Impact

Without these updates:
- ❌ Plugin installation system cannot work properly
- ❌ Plugin validation is incomplete
- ❌ Git-based plugin installation fails
- ❌ Plugin trust levels cannot be determined
- ❌ Build information cannot be stored

## Next Steps

1. Clone/update the `vps-init-core` repository
2. Add the missing types and fields listed above
3. Release a new version of vps-init-core
4. Update VPS-Init to use the updated core package
5. Test plugin installation and validation functionality

## Notes

- All changes should maintain backward compatibility
- Consider adding JSON tags for serialization
- Update documentation to reflect new fields and their purposes