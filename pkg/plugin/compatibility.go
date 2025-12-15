package plugin

import (
	"fmt"
	"runtime"

	"github.com/Masterminds/semver/v3"
)

// CompatibilityChecker handles plugin compatibility validation
type CompatibilityChecker struct {
	vpsInitVersion string
	currentOS      string
	currentArch    string
	goVersion      string
}

// NewCompatibilityChecker creates a new compatibility checker
func NewCompatibilityChecker(vpsInitVersion, goVersion string) *CompatibilityChecker {
	return &CompatibilityChecker{
		vpsInitVersion: vpsInitVersion,
		currentOS:      runtime.GOOS,
		currentArch:    runtime.GOARCH,
		goVersion:      goVersion,
	}
}

// CheckCompatibility checks if a plugin is compatible with the current environment
func (cc *CompatibilityChecker) CheckCompatibility(plugin Plugin) *CompatibilityResult {
	compat := plugin.Compatibility()
	result := &CompatibilityResult{
		Compatible: true,
		Warnings:   []string{},
	}

	// Check VPS-Init version compatibility
	if err := cc.checkVPSInitVersion(compat); err != nil {
		result.Compatible = false
		result.Errors = append(result.Errors, err.Error())
	}

	// Check platform compatibility
	if err := cc.checkPlatformCompatibility(compat); err != nil {
		result.Compatible = false
		result.Errors = append(result.Errors, err.Error())
	}

	// Check Go version compatibility
	if warning := cc.checkGoVersion(compat); warning != "" {
		result.Warnings = append(result.Warnings, warning)
	}

	return result
}

// CompatibilityResult contains the result of compatibility checking
type CompatibilityResult struct {
	Compatible bool
	Warnings   []string
	Errors     []string
}

// checkVPSInitVersion checks VPS-Init version compatibility
func (cc *CompatibilityChecker) checkVPSInitVersion(compat Compatibility) error {
	if compat.MinVPSInitVersion == "" {
		return nil // No minimum version requirement
	}

	currentVersion, err := semver.NewVersion(cc.vpsInitVersion)
	if err != nil {
		return fmt.Errorf("invalid current VPS-Init version: %w", err)
	}

	minVersion, err := semver.NewConstraint(compat.MinVPSInitVersion)
	if err != nil {
		return fmt.Errorf("invalid minimum version constraint: %w", err)
	}

	if !minVersion.Check(currentVersion) {
		return fmt.Errorf("plugin requires VPS-Init version %s, but current version is %s",
			compat.MinVPSInitVersion, cc.vpsInitVersion)
	}

	// Check maximum version if specified
	if compat.MaxVPSInitVersion != "" {
		maxVersion, err := semver.NewConstraint("<=" + compat.MaxVPSInitVersion)
		if err != nil {
			return fmt.Errorf("invalid maximum version constraint: %w", err)
		}

		if !maxVersion.Check(currentVersion) {
			return fmt.Errorf("plugin requires VPS-Init version <= %s, but current version is %s",
				compat.MaxVPSInitVersion, cc.vpsInitVersion)
		}
	}

	return nil
}

// checkPlatformCompatibility checks platform compatibility
func (cc *CompatibilityChecker) checkPlatformCompatibility(compat Compatibility) error {
	if len(compat.Platforms) == 0 {
		return nil // No platform restrictions
	}

	currentPlatform := fmt.Sprintf("%s/%s", cc.currentOS, cc.currentArch)

	for _, platform := range compat.Platforms {
		if platform == currentPlatform {
			return nil // Platform is supported
		}

		// Check for wildcard matches (e.g., "linux/*" or "*/amd64")
		if platform == fmt.Sprintf("%s/*", cc.currentOS) ||
		   platform == fmt.Sprintf("*/%s", cc.currentArch) {
			return nil // Platform wildcard matches
		}
	}

	return fmt.Errorf("plugin is not compatible with platform %s", currentPlatform)
}

// checkGoVersion checks Go version compatibility
func (cc *CompatibilityChecker) checkGoVersion(compat Compatibility) string {
	if compat.GoVersion == "" {
		return "" // No Go version requirement
	}

	// This is a simplified check - in practice, you'd want more sophisticated version parsing
	if cc.goVersion != compat.GoVersion {
		return fmt.Sprintf("plugin was built with Go %s, but current Go version is %s",
			compat.GoVersion, cc.goVersion)
	}

	return ""
}

// ResolveDependencies resolves plugin dependencies with version constraints
func (cc *CompatibilityChecker) ResolveDependencies(plugins []Plugin) (*DependencyGraph, error) {
	graph := NewDependencyGraph()

	// Add all plugins to the graph
	for _, plugin := range plugins {
		graph.AddPlugin(plugin)
	}

	// Resolve dependencies
	for _, plugin := range plugins {
		dependencies := plugin.Dependencies()
		for _, dep := range dependencies {
			if err := graph.AddDependency(plugin.Name(), dep.Name, dep.Version); err != nil {
				return nil, fmt.Errorf("failed to add dependency for plugin %s: %w", plugin.Name(), err)
			}
		}
	}

	// Check for circular dependencies
	if cycles := graph.FindCycles(); len(cycles) > 0 {
		return nil, fmt.Errorf("circular dependencies detected: %v", cycles)
	}

	return graph, nil
}

// DependencyGraph represents plugin dependencies
type DependencyGraph struct {
	plugins    map[string]*PluginNode
	dependents map[string][]string
}

// PluginNode represents a plugin in the dependency graph
type PluginNode struct {
	Plugin      Plugin
	Dependencies map[string]string // name -> version constraint
	Dependents   []string          // names of dependent plugins
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		plugins:    make(map[string]*PluginNode),
		dependents: make(map[string][]string),
	}
}

// AddPlugin adds a plugin to the graph
func (dg *DependencyGraph) AddPlugin(plugin Plugin) {
	node := &PluginNode{
		Plugin:       plugin,
		Dependencies: make(map[string]string),
	}

	dg.plugins[plugin.Name()] = node
}

// AddDependency adds a dependency relationship
func (dg *DependencyGraph) AddDependency(pluginName, depName, version string) error {
	pluginNode, exists := dg.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	pluginNode.Dependencies[depName] = version
	dg.dependents[depName] = append(dg.dependents[depName], pluginName)

	return nil
}

// FindCycles finds circular dependencies in the graph
func (dg *DependencyGraph) FindCycles() [][]string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	cycles := [][]string{}

	var dfs func(string, []string) bool
	dfs = func(node string, path []string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		pluginNode, exists := dg.plugins[node]
		if !exists {
			return false
		}

		for depName := range pluginNode.Dependencies {
			if !visited[depName] {
				if dfs(depName, path) {
					return true
				}
			} else if recStack[depName] {
				// Found a cycle
				cycleStart := -1
				for i, name := range path {
					if name == depName {
						cycleStart = i
						break
					}
				}
				if cycleStart != -1 {
					cycle := append([]string{}, path[cycleStart:]...)
					cycles = append(cycles, cycle)
				}
			}
		}

		recStack[node] = false
		return false
	}

	for pluginName := range dg.plugins {
		if !visited[pluginName] {
			dfs(pluginName, []string{})
		}
	}

	return cycles
}

// GetLoadOrder returns the order in which plugins should be loaded
func (dg *DependencyGraph) GetLoadOrder() ([]string, error) {
	if cycles := dg.FindCycles(); len(cycles) > 0 {
		return nil, fmt.Errorf("circular dependencies prevent loading")
	}

	visited := make(map[string]bool)
	order := []string{}

	var visit func(string)
	visit = func(pluginName string) {
		if visited[pluginName] {
			return
		}

		pluginNode, exists := dg.plugins[pluginName]
		if !exists {
			return
		}

		// Visit dependencies first
		for depName := range pluginNode.Dependencies {
			visit(depName)
		}

		visited[pluginName] = true
		order = append(order, pluginName)
	}

	for pluginName := range dg.plugins {
		visit(pluginName)
	}

	return order, nil
}