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

// RegisterPrompts registers the nine curated prompts on the MCP server.
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

	add(&mcp.Prompt{
		Name:        "law_research_workflow",
		Title:       "Research a law section by section",
		Description: "Efficient reading of a norm: summary and table of contents first, then only the relevant sections — avoiding the full text of long laws.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results", Required: true},
			{Name: "question", Title: "Question", Description: "Optional research question to focus the reading", Required: false},
		},
	}, lawResearchWorkflowTemplate)
	add(&mcp.Prompt{
		Name:        "answer_constitutional_question",
		Title:       "Answer a constitutional question",
		Description: "Answer a question about the Chilean Constitution (Decreto 100, 242302): locate the relevant chapters/articles via the table of contents and cite the actual text. Supports historical versions via version_date.",
		Arguments: []*mcp.PromptArgument{
			{Name: "question", Title: "Question", Description: "The constitutional question to answer (e.g. '¿qué dice sobre el derecho de propiedad?')", Required: true},
			{Name: "article_hint", Title: "Article hint", Description: "Optional article hint (free text, e.g. '19', '19 Nº24', '93', 'transitoria primera') to prioritize a TOC entry", Required: false},
			{Name: "version_date", Title: "Version date", Description: "Optional version in force at this date (YYYY-MM-DD) — defaults to the latest version", Required: false},
		},
	}, answerConstitutionalQuestionTemplate)

	add(&mcp.Prompt{
		Name:        "check_norm_constitutionality",
		Title:       "Check norm vs Constitution",
		Description: "Assess whether a norm is compatible with the Chilean Constitution (Decreto 100, 242302) by contrasting the relevant sections side-by-side. Supports historical versions via version_date.",
		Arguments: []*mcp.PromptArgument{
			{Name: "norm_id", Title: "Norm id", Description: "The norm id (norm_id) from search_laws results of the norm to contrast", Required: true},
			{Name: "question", Title: "Question", Description: "Optional focus question (e.g. '¿vulnera igualdad ante la ley?')", Required: false},
			{Name: "version_date", Title: "Version date", Description: "Optional version in force at this date (YYYY-MM-DD) — applied to both norms for temporal coherence", Required: false},
		},
	}, checkNormConstitutionalityTemplate)
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

// lawResearchWorkflowTemplate is the economical reading flow: the summary
// carries the size and the section ids, so the model reads only the
// sections it needs instead of the full text of long norms.
func lawResearchWorkflowTemplate(args map[string]string) string {
	question := ""
	if args["question"] != "" {
		question = fmt.Sprintf(" Focus on answering: %s.", args["question"])
	}
	return fmt.Sprintf(`Research Chilean norm %s efficiently.%s

Step 1: call %s(norm_id=%s) — it returns the summary, the size and the table of contents with the section ids.
Step 2: read the structure and call %s(norm_id=%s, section_id=<section id>) ONLY for the sections relevant to the question. NEVER call %s without section_id unless the Size line shows a short norm.

For every claim, cite the article number. NEVER invent articles or content — verify everything against the actual returned text.`,
		args["norm_id"], question, toolGetLawSummary, args["norm_id"], toolGetLaw, args["norm_id"], toolGetLaw)
}

// answerConstitutionalQuestionTemplate answers a question about the CPR (Decreto 100, 242302).
// The Constitution is ~410K chars, so the template forces the economical flow
// get_law_summary → section_id → get_law(section_id), including DISPOSICIONES TRANSITORIAS.
func answerConstitutionalQuestionTemplate(args map[string]string) string {
	question := args["question"]
	hint := args["article_hint"]
	versionDate := args["version_date"]

	hintPart := ""
	if hint != "" {
		hintPart = fmt.Sprintf(" Article hint: %s (fuzzy match against Estructura[].n, e.g. \"19\" → Artículo 19, \"transitoria primera\" → Disposición Transitoria PRIMERA).", hint)
	}
	versionPart := ""
	versionArg := ""
	versionArgLaw := ""
	if versionDate != "" {
		versionPart = fmt.Sprintf(" Version date: %s — the \"Version: as of\" line confirms which version you read.", versionDate)
		versionArg = fmt.Sprintf(", version_date=%s", versionDate)
		versionArgLaw = fmt.Sprintf(" (version_date=%s)", versionDate)
	}
	return fmt.Sprintf(`Answer the constitutional question "%s" about the Chilean Constitution (Decreto 100, norm_id=242302).%s%s

Step 1: call %s(norm_id=242302%s) — it returns the Size and the table of contents with section_ids (Chapters I-XV and DISPOSICIONES TRANSITORIAS). Read the Structure to map the question to 1-3 relevant section_ids. If article_hint is provided ("%s"), prioritize the entry matching that hint (fuzzy match against Estructura[].n); otherwise map question keywords to chapter/article names (e.g. "propiedad" → Capítulo III, "plebiscito 2022" → DISPOSICIONES TRANSITORIAS → VIGÉSIMAQUINTA).

Step 2: call %s(norm_id=242302, section_id=<section id>%s) ONLY for the sections identified. NEVER call %s without section_id — the Constitution is ~410K chars; the Size line tells you whether a section is short. When a transitory provision modifies a permanent article (e.g. transitoria mentioning Art. 142), cite both. Also consider DISPOSICIONES TRANSITORIAS as eligible sections when the question mentions reforma/plebiscito/vigencia/transitoria.

For every claim, cite the source article (e.g. Art. 19 Nº24). NEVER invent articles or content — verify everything against the actual returned text%s.

Interpret the text conditionally: use "conforme al texto retornado ... podría interpretarse como (in)compatible/constitucional/inconstitucional, en la medida que ..." This is general information, not legal advice; the binding qualification corresponds to the Tribunal Constitucional. Verify in the official text at bcn.cl and consult a qualified professional.`,
		question, hintPart, versionPart, toolGetLawSummary, versionArg, hint, toolGetLaw, versionArg, toolGetLaw, versionArgLaw)
}

// checkNormConstitutionalityTemplate contrasts a norm against the CPR (Decreto 100, 242302).
func checkNormConstitutionalityTemplate(args map[string]string) string {
	normID := args["norm_id"]
	question := args["question"]
	versionDate := args["version_date"]

	questionPart := ""
	if question != "" {
		questionPart = fmt.Sprintf(" Focus: %s.", question)
	}
	versionPart := ""
	versionArg := ""
	if versionDate != "" {
		versionPart = fmt.Sprintf(" Version date: %s (applied to both norms for temporal coherence) — the \"Version: as of\" line confirms the version.", versionDate)
		versionArg = fmt.Sprintf(", version_date=%s", versionDate)
	}
	questionArg := ""
	if question != "" {
		questionArg = fmt.Sprintf(" Question: %s.", question)
	}
	return fmt.Sprintf(`Assess whether Chilean norm %s is compatible with the Chilean Constitution (Decreto 100, norm_id=242302).%s%s%s

Step 1: call %s(norm_id=%s%s) for the target norm and %s(norm_id=242302%s) for the Constitution — both with version_date if provided, to keep temporal coherence. Each summary gives the Size and TOC with section_ids (including DISPOSICIONES TRANSITORIAS).

Step 2: from both TOCs select the 1-3 most relevant sections (using question "%s" if provided or the summaries). Include DISPOSICIONES TRANSITORIAS when the question mentions reforma/plebiscito/vigencia/transitoria.

Step 3: call %s(norm_id=%s, section_id=<section id>%s) and %s(norm_id=242302, section_id=<section id>%s) for each selected section. NEVER call %s without section_id unless the Size line shows a short norm.

Step 4: present a side-by-side textual parallelism "Art. X of norm %s says ... | Art. Z CPR says ..." citing articles of both norms, and analyze compatibility conditionally ("conforme al texto retornado de Art. X and Art. Z CPR, podría interpretarse como constitucional/inconstitucional, en la medida que ...").

For every claim, cite the article. NEVER invent articles or content — verify against the returned text. The "Version: as of" line confirms the version.

Recommended: call %s(norm_id=%s) if the target norm may have had TC review — if the modificatorias group exists, summarize it.

This analysis is general information, not legal advice; the binding qualification corresponds to the Tribunal Constitucional. Verify at bcn.cl and consult a qualified professional.`,
		normID, questionArg, questionPart, versionPart, toolGetLawSummary, normID, versionArg, toolGetLawSummary, versionArg, question, toolGetLaw, normID, versionArg, toolGetLaw, versionArg, toolGetLaw, normID, toolGetLawHistory, normID)
}
