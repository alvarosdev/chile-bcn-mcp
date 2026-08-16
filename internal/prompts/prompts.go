// Package prompts provides the curated MCP prompts of the chile-bcn-mcp
// server. Each prompt is a PURE template: it injects the received
// arguments into a message that guides the model on how to use the tools
// correctly. Serving a prompt never touches the BCN API.
//
// Templates are baked into the binary via go:embed (prompts.yaml) and
// rendered with text/template + named placeholders (e.g. {{.norm_id}}).
package prompts

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"text/template"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

// Tool names referenced by the prompt templates. Kept as constants so a
// rename is caught by the tests (templates must only reference tools that
// are actually registered).
const (
	toolSearchLaws    = "search_laws"
	toolGetLaw        = "get_law"
	toolGetLawSummary = "get_law_summary"
	toolGetLawHistory = "get_law_history"
)

// ToolNames returns the tool names the prompts reference.
func ToolNames() []string {
	return []string{toolSearchLaws, toolGetLaw, toolGetLawSummary, toolGetLawHistory}
}

//go:embed prompts.yaml
var rawPrompts []byte

// expectedPromptNames is the exact set of prompts the server must expose.
// Order matches prompts/list expectation; validation enforces exactly these 9.
var expectedPromptNames = []string{
	"analyze_law",
	"search_legal_topic",
	"compare_law_versions",
	"trace_law_history",
	"check_law_validity",
	"explain_law_simply",
	"law_research_workflow",
	"answer_constitutional_question",
	"check_norm_constitutionality",
}

// allowedPlaceholders is the closed whitelist for template vars.
// Any placeholder outside this set fails LoadEmbedded validation.
var allowedPlaceholders = map[string]bool{
	"norm_id":              true,
	"topic":                true,
	"from_date":            true,
	"to_date":              true,
	"date":                 true,
	"aspect":               true,
	"audience":             true,
	"question":             true,
	"article_hint":         true,
	"version_date":         true,
	"tool_search_laws":     true,
	"tool_get_law":         true,
	"tool_get_law_summary": true,
	"tool_get_law_history": true,
}

// placeholderRe extracts placeholder names from {{.var}} and {{if .var}}.
var placeholderRe = regexp.MustCompile(`{{\s*(?:if\s+)?\.([a-z_]+)`)

// PromptSet holds the parsed, validated prompt templates baked into the binary.
type PromptSet struct {
	templates map[string]*template.Template
}

// rawPromptsFile is the YAML shape of prompts.yaml.
type rawPromptsFile struct {
	Prompts map[string]string `yaml:"prompts"`
}

// LoadEmbedded reads and validates the embedded prompts.yaml contract.
// The file is baked into the binary via go:embed; no external file is required.
func LoadEmbedded() (*PromptSet, error) {
	return loadFromBytes(rawPrompts)
}

func loadFromBytes(data []byte) (*PromptSet, error) {
	var raw rawPromptsFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse embedded prompts: %w", err)
	}
	if len(raw.Prompts) != len(expectedPromptNames) {
		return nil, fmt.Errorf("prompts: want %d prompts, got %d", len(expectedPromptNames), len(raw.Prompts))
	}
	// Ensure exactly the expected keys, no extra/missing.
	expected := make(map[string]bool, len(expectedPromptNames))
	for _, n := range expectedPromptNames {
		expected[n] = true
	}
	for name := range raw.Prompts {
		if !expected[name] {
			return nil, fmt.Errorf("prompts: unexpected prompt %q", name)
		}
	}
	for _, name := range expectedPromptNames {
		if _, ok := raw.Prompts[name]; !ok {
			return nil, fmt.Errorf("prompts: missing prompt %q", name)
		}
	}
	templates := make(map[string]*template.Template, len(raw.Prompts))
	for name, text := range raw.Prompts {
		// Validate placeholders are within whitelist.
		for _, m := range placeholderRe.FindAllStringSubmatch(text, -1) {
			ph := m[1]
			if !allowedPlaceholders[ph] {
				return nil, fmt.Errorf("prompts: prompt %q uses unknown placeholder %q", name, ph)
			}
		}
		tmpl, err := template.New(name).Option("missingkey=error").Parse(text)
		if err != nil {
			return nil, fmt.Errorf("prompts: parse prompt %q: %w", name, err)
		}
		templates[name] = tmpl
	}
	return &PromptSet{templates: templates}, nil
}

// render executes the named prompt template with args + tool vars.
// It pre-populates all allowed placeholders with "" so missing args render as empty
// (preserving TestMissingArgServesWithoutError) while still using missingkey=error
// to catch typos via whitelist validation at load time.
func (ps *PromptSet) render(name string, args map[string]string) (string, error) {
	tmpl, ok := ps.templates[name]
	if !ok {
		return "", fmt.Errorf("prompt %q not found", name)
	}
	// Build data with all allowed keys defaulting to "".
	data := make(map[string]string, len(allowedPlaceholders))
	for k := range allowedPlaceholders {
		data[k] = ""
	}
	// Tool vars (always present).
	data["tool_search_laws"] = toolSearchLaws
	data["tool_get_law"] = toolGetLaw
	data["tool_get_law_summary"] = toolGetLawSummary
	data["tool_get_law_history"] = toolGetLawHistory
	// Overlay caller args.
	for k, v := range args {
		// Only overlay allowed keys; ignore unexpected keys (client may send extra).
		if allowedPlaceholders[k] {
			data[k] = v
		}
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RegisterPrompts registers the nine curated prompts on the MCP server.
// promptSet must be loaded via LoadEmbedded; it is injected (no filesystem access).
func RegisterPrompts(srv *mcp.Server, ps *PromptSet) {
	add := func(p *mcp.Prompt, name string) {
		srv.AddPrompt(p, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			text, err := ps.render(name, req.Params.Arguments)
			if err != nil {
				// Should not happen for validated templates; fallback to error as prompt text.
				text = fmt.Sprintf("prompt render error for %q: %v", name, err)
			}
			return &mcp.GetPromptResult{
				Messages: []*mcp.PromptMessage{{
					Role:    mcp.Role("user"),
					Content: &mcp.TextContent{Text: text},
				}},
			}, nil
		})
	}

	add(&mcp.Prompt{
		Name:        "analyze_law",
		Title:       "Analyze a Chilean law",
		Description: "Structured legal analysis of a norm: purpose, scope, obligations, sanctions and entry into force, with citations to the actual text.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results", Required: true},
			{Name: "aspect", Title: "Aspect", Description: "Optional analysis dimension to focus on (purpose, obligations, sanctions...)", Required: false},
		},
	}, "analyze_law")

	add(&mcp.Prompt{
		Name:        "search_legal_topic",
		Title:       "Find norms about a topic",
		Description: "Guided search over LeyChile: pick a good query, read summaries before opening norms, and verify with the full text.",
		Arguments: []*mcp.PromptArgument{
			{Name: "topic", Title: "Topic", Description: "The legal topic to search for", Required: true},
		},
	}, "search_legal_topic")

	add(&mcp.Prompt{
		Name:        "compare_law_versions",
		Title:       "Compare two versions of a law",
		Description: "Compare a norm between two dates using historical versions (version_date), reporting what changed and when.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results", Required: true},
			{Name: "from_date", Title: "From date", Description: "Start date (YYYY-MM-DD)", Required: true},
			{Name: "to_date", Title: "To date", Description: "End date (YYYY-MM-DD)", Required: true},
		},
	}, "compare_law_versions")

	add(&mcp.Prompt{
		Name:        "trace_law_history",
		Title:       "Trace what modified a law",
		Description: "Trace the legislative history of a norm: identify the laws that modified it and present a chronological timeline.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results", Required: true},
		},
	}, "trace_law_history")

	add(&mcp.Prompt{
		Name:        "check_law_validity",
		Title:       "Is this norm in force?",
		Description: "Check whether a norm is in force, derogated, or in force at a given date, based on its metadata and validity window.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results", Required: true},
			{Name: "date", Title: "Date", Description: "Optional date to check validity at (YYYY-MM-DD)", Required: false},
		},
	}, "check_law_validity")

	add(&mcp.Prompt{
		Name:        "explain_law_simply",
		Title:       "Explain a law in plain language",
		Description: "Explain a norm without legal jargon, citing the source article for every claim, with a no-legal-advice disclaimer.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results", Required: true},
			{Name: "audience", Title: "Audience", Description: "Optional target audience (e.g. students, small business owners)", Required: false},
		},
	}, "explain_law_simply")

	add(&mcp.Prompt{
		Name:        "law_research_workflow",
		Title:       "Research a law section by section",
		Description: "Efficient reading of a norm: summary and table of contents first, then only the relevant sections — avoiding the full text of long laws.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results", Required: true},
			{Name: "question", Title: "Question", Description: "Optional research question to focus the reading", Required: false},
		},
	}, "law_research_workflow")
	add(&mcp.Prompt{
		Name:        "answer_constitutional_question",
		Title:       "Answer a constitutional question",
		Description: "Answer a question about the Chilean Constitution (Decreto 100, 242302): locate the relevant chapters/articles via the table of contents and cite the actual text. Supports historical versions via version_date.",
		Arguments: []*mcp.PromptArgument{
			{Name: "question", Title: "Question", Description: "The constitutional question to answer (e.g. '¿qué dice sobre el derecho de propiedad?')", Required: true},
			{Name: "article_hint", Title: "Article hint", Description: "Optional article hint (free text, e.g. '19', '19 Nº24', '93', 'transitoria primera') to prioritize a TOC entry", Required: false},
			{Name: "version_date", Title: "Version date", Description: "Optional version in force at this date (YYYY-MM-DD) — defaults to the latest version", Required: false},
		},
	}, "answer_constitutional_question")

	add(&mcp.Prompt{
		Name:        "check_norm_constitutionality",
		Title:       "Check norm vs Constitution",
		Description: "Assess whether a norm is compatible with the Chilean Constitution (Decreto 100, 242302) by contrasting the relevant sections side-by-side. Supports historical versions via version_date.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results of the norm to contrast", Required: true},
			{Name: "question", Title: "Question", Description: "Optional focus question (e.g. '¿vulnera igualdad ante la ley?')", Required: false},
			{Name: "version_date", Title: "Version date", Description: "Optional version in force at this date (YYYY-MM-DD) — applied to both norms for temporal coherence", Required: false},
		},
	}, "check_norm_constitutionality")
}
