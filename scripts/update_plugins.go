//go:build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

// PluginUpdate represents the methods to add to existing plugins
const pluginMethods = `
	// Enhanced plugin interface methods
func (p *Plugin) Validate() error {
	// TODO: Add plugin-specific validation logic
	return nil
}

func (p *Plugin) Dependencies() []plugin.Dependency {
	return []plugin.Dependency{
		// TODO: Add plugin dependencies with version constraints
	}
}

func (p *Plugin) Compatibility() plugin.Compatibility {
	return plugin.Compatibility{
		MinVPSInitVersion: "1.0.0",
		GoVersion:         "1.19",
		Platforms:         []string{"linux/amd64", "linux/arm64"},
		Tags:              []string{"TODO", "add", "relevant", "tags"},
	}
}

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        p.Name(),
		Description: p.Description(),
		Version:     p.Version(),
		Author:      p.Author(),
		License:     "MIT",
		Repository:  "github.com/wasilwamark/vps-init-plugins/" + p.Name(),
		Tags:        []string{"TODO", "add", "tags"},
		Validated:   true,
		TrustLevel:  "official",
		BuildInfo: plugin.BuildInfo{
			GoVersion: "1.21",
		},
	}
}
`

func main() {
	servicesDir := "./internal/services"

	// Walk through all service directories
	entries, err := ioutil.ReadDir(servicesDir)
	if err != nil {
		fmt.Printf("Error reading services directory: %v\n", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginDir := filepath.Join(servicesDir, entry.Name())
		pluginFile := filepath.Join(pluginDir, "plugin.go")

		// Check if plugin.go exists
		if _, err := os.Stat(pluginFile); os.IsNotExist(err) {
			continue
		}

		fmt.Printf("Updating plugin: %s\n", entry.Name())
		err := updatePluginFile(pluginFile, entry.Name())
		if err != nil {
			fmt.Printf("Error updating %s: %v\n", entry.Name(), err)
		} else {
			fmt.Printf("Successfully updated %s\n", entry.Name())
		}
	}
}

func updatePluginFile(pluginFile, pluginName string) error {
	// Read the existing plugin file
	content, err := ioutil.ReadFile(pluginFile)
	if err != nil {
		return fmt.Errorf("error reading plugin file: %w", err)
	}

	contentStr := string(content)

	// Remove old Dependencies() method if it exists
	oldDepsRegex := `func \(p \*Plugin\) Dependencies\(\) \[\]string \{[\s\S]*?\n\}`
	contentStr = regexp.MustCompile(oldDepsRegex).ReplaceAllString(contentStr, "")

	// Find where to insert the new methods (after the old Dependencies method)
	// Look for the end of existing interface methods
	insertPoint := findInsertPoint(contentStr)
	if insertPoint == -1 {
		return fmt.Errorf("could not find insertion point")
	}

	// Insert the new methods
	newContent := contentStr[:insertPoint] + pluginMethods + contentStr[insertPoint:]

	// Write the updated content back to the file
	err = ioutil.WriteFile(pluginFile, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing updated plugin file: %w", err)
	}

	return nil
}

func findInsertPoint(content string) int {
	// Look for the end of the last interface method
	patterns := []string{
		`func \(p \*Plugin\) GetRootCommand\(\) \*cobra\.Command \{[^}]*\n\}`,
		`func \(p \*Plugin\) Dependencies\(\) \[\]string \{[^}]*\n\}`,
		`func \(p \*Plugin\) Author\(\) string \{[^}]*\}`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatchIndex(content, -1)
		if len(matches) > 0 {
			// Return the position after the last match
			lastMatch := matches[len(matches)-1]
			return lastMatch[1]
		}
	}

	return -1
}