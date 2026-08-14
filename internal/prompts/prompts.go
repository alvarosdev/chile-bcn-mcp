// Package prompts provides the curated MCP prompts of the chile-bcn-mcp
// server. Each prompt is a PURE template: it injects the received
// arguments into a message that guides the model on how to use the tools
// correctly. Serving a prompt never touches the BCN API.
package prompts

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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

// template is a prompt builder: receives the arguments from prompts/get
// and returns the message text. Pure — no external calls.
type template func(args map[string]string) string

// RegisterPrompts registers the six curated prompts on the MCP server.
func RegisterPrompts(srv *mcp.Server) {
	add := func(p *mcp.Prompt, t template) {
		srv.AddPrompt(p, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return &mcp.GetPromptResult{
				Messages: []*mcp.PromptMessage{{
					Role:    mcp.Role("user"),
					Content: &mcp.TextContent{Text: t(req.Params.Arguments)},
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
	}, analyzeLawTemplate)

	add(&mcp.Prompt{
		Name:        "search_legal_topic",
		Title:       "Find norms about a topic",
		Description: "Guided search over LeyChile: pick a good query, read summaries before opening norms, and verify with the full text.",
		Arguments: []*mcp.PromptArgument{
			{Name: "topic", Title: "Topic", Description: "The legal topic to search for", Required: true},
		},
	}, searchLegalTopicTemplate)

	add(&mcp.Prompt{
		Name:        "compare_law_versions",
		Title:       "Compare two versions of a law",
		Description: "Compare a norm between two dates using historical versions (version_date), reporting what changed and when.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results", Required: true},
			{Name: "from_date", Title: "From date", Description: "Start date (YYYY-MM-DD)", Required: true},
			{Name: "to_date", Title: "To date", Description: "End date (YYYY-MM-DD)", Required: true},
		},
	}, compareLawVersionsTemplate)

	add(&mcp.Prompt{
		Name:        "trace_law_history",
		Title:       "Trace what modified a law",
		Description: "Trace the legislative history of a norm: identify the laws that modified it and present a chronological timeline.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results", Required: true},
		},
	}, traceLawHistoryTemplate)

	add(&mcp.Prompt{
		Name:        "check_law_validity",
		Title:       "Is this norm in force?",
		Description: "Check whether a norm is in force, derogated, or in force at a given date, based on its metadata and validity window.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results", Required: true},
			{Name: "date", Title: "Date", Description: "Optional date to check validity at (YYYY-MM-DD)", Required: false},
		},
	}, checkLawValidityTemplate)

	add(&mcp.Prompt{
		Name:        "explain_law_simply",
		Title:       "Explain a law in plain language",
		Description: "Explain a norm without legal jargon, citing the source article for every claim, with a no-legal-advice disclaimer.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results", Required: true},
			{Name: "audience", Title: "Audience", Description: "Optional target audience (e.g. students, small business owners)", Required: false},
		},
	}, explainLawSimplyTemplate)
}

// analyzeLawTemplate is the structured legal analysis flow.
func analyzeLawTemplate(args map[string]string) string {
	t := fmt.Sprintf(`Analyze Chilean norm %s.

Step 1: call %s(norm_id=%s) to understand the scope and official summary.
Step 2: call %s(norm_id=%s, structure_only=true) to read the table of contents.
Step 3: call %s(norm_id=%s) to read the full text.

Structure the analysis: purpose, scope, obligations, sanctions, entry into force.`,
		args["norm_id"], toolGetLawSummary, args["norm_id"], toolGetLaw, args["norm_id"], toolGetLaw, args["norm_id"])
	if args["aspect"] != "" {
		t += fmt.Sprintf(" Focus the analysis on: %s.", args["aspect"])
	}
	return t + `

For every claim, cite the article number. NEVER invent articles or content — verify everything against the actual returned text.`
}

// searchLegalTopicTemplate is the guided search flow.
func searchLegalTopicTemplate(args map[string]string) string {
	return fmt.Sprintf(`Find Chilean norms about %s.

Step 1: call %s(query="%s"). If results are noisy, refine the query with the norm type (e.g. "Ley %s", "Decreto %s").
Step 2: read the summaries in the results — do NOT open a norm before reading its summary.
Step 3: pick the most relevant result and confirm with %s(norm_id=...).
Step 4: verify with %s(norm_id=...) before stating anything as fact. NEVER invent norm numbers or content.`,
		args["topic"], toolSearchLaws, args["topic"], args["topic"], args["topic"], toolGetLawSummary, toolGetLaw)
}

// compareLawVersionsTemplate compares two historical versions.
func compareLawVersionsTemplate(args map[string]string) string {
	return fmt.Sprintf(`Compare norm %s between %s and %s.

Call %s(norm_id=%s, version_date=%s) and %s(norm_id=%s, version_date=%s).

Report what changed: which articles were modified, added or removed, and when each change took effect. Verify both versions against the actual returned text — the "Version: as of" line tells you which version you are reading.`,
		args["norm_id"], args["from_date"], args["to_date"],
		toolGetLaw, args["norm_id"], args["from_date"], toolGetLaw, args["norm_id"], args["to_date"])
}

// traceLawHistoryTemplate traces the legislative history. IMPORTANT: to
// read any of the norms listed in the history, use the id_norma_hl value
// (the LeyChile id of the record's norm) — never the id_norma field nor
// the number inside the history URL.
func traceLawHistoryTemplate(args map[string]string) string {
	return fmt.Sprintf(`Trace the legislative history of norm %s.

Call %s(norm_id=%s). Identify the "modificatorias" group (laws that modified this norm). For each: date, law number, summary.

To READ any of these norms, call %s with the id_norma_hl value (the LeyChile id of the record's norm) — never the id_norma field nor the number in the history URL. Present a chronological timeline.`,
		args["norm_id"], toolGetLawHistory, args["norm_id"], toolGetLaw)
}

// checkLawValidityTemplate checks validity, optionally at a date.
func checkLawValidityTemplate(args map[string]string) string {
	datePart := ""
	if args["date"] != "" {
		datePart = fmt.Sprintf(" as of %s", args["date"])
	}
	call := fmt.Sprintf("%s(norm_id=%s)", toolGetLawSummary, args["norm_id"])
	if args["date"] != "" {
		call += fmt.Sprintf(", version_date=%s", args["date"])
	}
	return fmt.Sprintf(`Check whether norm %s is in force%s.

Call %s. Report: in force / derogated / in force at the given date, based on the derogated flag and the validity window. Distinguish "vigente", "derogada" and "vigente a la fecha". Verify against the actual returned metadata.`,
		args["norm_id"], datePart, call)
}

// explainLawSimplyTemplate explains in plain language with citations.
func explainLawSimplyTemplate(args map[string]string) string {
	audience := ""
	if args["audience"] != "" {
		audience = fmt.Sprintf(" Audience: %s.", args["audience"])
	}
	return fmt.Sprintf(`Explain norm %s in plain language.%s

Call %s(norm_id=%s), then %s(norm_id=%s) to read the actual text. Explain without legal jargon.

For every claim, cite the source article. End with: this explanation is not legal advice; consult the official text at bcn.cl.`,
		args["norm_id"], audience, toolGetLawSummary, args["norm_id"], toolGetLaw, args["norm_id"])
}
