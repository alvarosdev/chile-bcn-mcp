// Package internal provides shared prompt rendering for bcn and cgr domains.
// It is pure logic without domain knowledge — each domain passes its own
// expected names, allowed placeholders and tool vars.
package internal

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"

	"gopkg.in/yaml.v3"
)

// PromptSet holds parsed, validated prompt templates.
type PromptSet struct {
	Templates map[string]*template.Template
}

// RawPromptsFile is the YAML shape of prompts.yaml.
type RawPromptsFile struct {
	Prompts map[string]string `yaml:"prompts"`
}

// PlaceholderRe extracts placeholder names from {{.var}} and {{if .var}}.
var PlaceholderRe = regexp.MustCompile(`{{\s*(?:if\s+)?\.([a-z_]+)`)

// Load validates and parses the prompts YAML.
func Load(data []byte, expected []string, allowed map[string]bool) (*PromptSet, error) {
	var raw RawPromptsFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse embedded prompts: %w", err)
	}
	if len(raw.Prompts) != len(expected) {
		return nil, fmt.Errorf("prompts: want %d prompts, got %d", len(expected), len(raw.Prompts))
	}
	expectedMap := make(map[string]bool, len(expected))
	for _, n := range expected {
		expectedMap[n] = true
	}
	for name := range raw.Prompts {
		if !expectedMap[name] {
			return nil, fmt.Errorf("prompts: unexpected prompt %q", name)
		}
	}
	for _, name := range expected {
		if _, ok := raw.Prompts[name]; !ok {
			return nil, fmt.Errorf("prompts: missing prompt %q", name)
		}
	}
	templates := make(map[string]*template.Template, len(raw.Prompts))
	for name, text := range raw.Prompts {
		for _, m := range PlaceholderRe.FindAllStringSubmatch(text, -1) {
			ph := m[1]
			if !allowed[ph] {
				return nil, fmt.Errorf("prompts: prompt %q uses unknown placeholder %q", name, ph)
			}
		}
		tmpl, err := template.New(name).Option("missingkey=error").Parse(text)
		if err != nil {
			return nil, fmt.Errorf("prompts: parse prompt %q: %w", name, err)
		}
		templates[name] = tmpl
	}
	return &PromptSet{Templates: templates}, nil
}

// Render executes the named prompt template with args + tool vars.
func (ps *PromptSet) Render(name string, args map[string]string, allowed map[string]bool, toolVars map[string]string) (string, error) {
	tmpl, ok := ps.Templates[name]
	if !ok {
		return "", fmt.Errorf("prompt %q not found", name)
	}
	data := make(map[string]string, len(allowed))
	for k := range allowed {
		data[k] = ""
	}
	for k, v := range toolVars {
		data[k] = v
	}
	for k, v := range args {
		if allowed[k] {
			data[k] = v
		}
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
