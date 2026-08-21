package bcn

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"
)

// PromptsSuite validates the curated prompts against an in-memory MCP
// server. Prompts are pure templates — no LawClient involved.
type PromptsSuite struct {
	suite.Suite
	ctx     context.Context
	session *mcp.ClientSession
}

func TestPromptsSuite(t *testing.T) {
	suite.Run(t, new(PromptsSuite))
}

func (s *PromptsSuite) SetupTest() {
	s.ctx = context.Background()

	ps, err := LoadEmbedded()
	s.Require().NoError(err)

	server := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	RegisterPrompts(server, ps)

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(s.ctx, serverTransport, nil)
	s.Require().NoError(err)
	s.T().Cleanup(func() { serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	s.session, err = client.Connect(s.ctx, clientTransport, nil)
	s.Require().NoError(err)
	s.T().Cleanup(func() { s.session.Close() })
}

func (s *PromptsSuite) getPrompt(name string, args map[string]string) string {
	res, err := s.session.GetPrompt(s.ctx, &mcp.GetPromptParams{Name: name, Arguments: args})
	s.Require().NoError(err)
	s.Require().NotEmpty(res.Messages)
	return res.Messages[0].Content.(*mcp.TextContent).Text
}

func (s *PromptsSuite) TestListPrompts() {
	res, err := s.session.ListPrompts(s.ctx, nil)
	s.Require().NoError(err)

	names := make(map[string]*mcp.Prompt, len(res.Prompts))
	for _, p := range res.Prompts {
		names[p.Name] = p
	}
	expected := []string{
		"analyze_law", "search_legal_topic", "compare_law_versions",
		"trace_law_history", "check_law_validity", "explain_law_simply",
		"law_research_workflow", "answer_constitutional_question", "check_norm_constitutionality",
		"interpret_law",
	}
	for _, name := range expected {
		p, ok := names[name]
		s.Require().True(ok, "prompt %s missing from prompts/list", name)
		s.NotEmpty(p.Description, "prompt %s must have a description", name)

		required := map[string]bool{}
		for _, a := range p.Arguments {
			required[a.Name] = a.Required
		}
		switch name {
		case "search_legal_topic":
			s.True(required["topic"], "prompt %s: topic must be required", name)
		case "answer_constitutional_question":
			s.True(required["question"], "prompt %s: question must be required", name)
			s.False(required["article_hint"], "prompt %s: article_hint must be optional", name)
			s.False(required["version_date"], "prompt %s: version_date must be optional", name)
		case "check_norm_constitutionality":
			s.True(required["norm_id"], "prompt %s: norm_id must be required", name)
			s.False(required["question"], "prompt %s: question must be optional", name)
			s.False(required["version_date"], "prompt %s: version_date must be optional", name)
		default:
			s.True(required["norm_id"], "prompt %s: norm_id must be required", name)
		}
	}
	s.Len(names, len(expected), "unexpected number of prompts")
	// Optional arguments stay optional.
	analyze := names["analyze_law"]
	for _, a := range analyze.Arguments {
		if a.Name == "aspect" {
			s.False(a.Required)
		}
	}
}

func (s *PromptsSuite) TestAnalyzeLawInjectsNormID() {
	text := s.getPrompt("analyze_law", map[string]string{"norm_id": "1195666"})
	s.Contains(text, "Analyze Chilean norm 1195666")
	s.Contains(text, "get_law_summary(norm_id=1195666)")
	s.Contains(text, "get_law(norm_id=1195666")
	s.Contains(text, "NEVER invent articles")
}

func (s *PromptsSuite) TestCompareVersionsInjectsDates() {
	text := s.getPrompt("compare_law_versions", map[string]string{
		"norm_id": "141599", "from_date": "2010-01-01", "to_date": "2020-01-01",
	})
	s.Contains(text, "get_law(norm_id=141599, version_date=2010-01-01)")
	s.Contains(text, "get_law(norm_id=141599, version_date=2020-01-01)")
}

func (s *PromptsSuite) TestLawResearchWorkflowGuidesSectionFlow() {
	text := s.getPrompt("law_research_workflow", map[string]string{"norm_id": "1195666"})
	s.Contains(text, "Research Chilean norm 1195666 efficiently")
	s.Contains(text, "get_law_summary(norm_id=1195666)")
	s.Contains(text, "section_id=<section id>")
	s.Contains(text, "NEVER invent articles")

	focused := s.getPrompt("law_research_workflow", map[string]string{"norm_id": "1195666", "question": "¿Qué sanciones establece?"})
	s.Contains(focused, "Focus on answering: ¿Qué sanciones establece?")
}

func (s *PromptsSuite) TestTraceHistoryUsesIDNormaHL() {
	text := s.getPrompt("trace_law_history", map[string]string{"norm_id": "1195666"})
	s.Contains(text, "get_law_history(norm_id=1195666)")
	s.Contains(text, "id_norma_hl")
	s.Contains(text, "never the id_norma field nor the number in the history URL")
}

func (s *PromptsSuite) TestCheckValidityWithAndWithoutDate() {
	withDate := s.getPrompt("check_law_validity", map[string]string{"norm_id": "1", "date": "2010-01-01"})
	s.Contains(withDate, "as of 2010-01-01")
	s.Contains(withDate, "version_date=2010-01-01")

	withoutDate := s.getPrompt("check_law_validity", map[string]string{"norm_id": "1"})
	s.NotContains(withoutDate, "version_date")
}

func (s *PromptsSuite) TestExplainSimplyHasDisclaimer() {
	text := s.getPrompt("explain_law_simply", map[string]string{"norm_id": "1", "audience": "students"})
	s.Contains(text, "Audience: students")
	s.Contains(text, "this explanation is not legal advice")
}

func (s *PromptsSuite) TestMissingArgServesWithoutError() {
	// The SDK does not enforce required args server-side: the handler must
	// serve the template anyway (the required flag guides the client).
	text := s.getPrompt("analyze_law", nil)
	s.NotContains(text, "<norm_id>", "placeholder must not leak when the arg is absent")
}

func (s *PromptsSuite) TestTemplatesReferenceOnlyRegisteredTools() {
	// Every tool name mentioned in any template must be one of the four
	// registered tools (the constants are the single source of truth).
	all := ""
	all += s.getPrompt("analyze_law", map[string]string{"norm_id": "1"})
	all += s.getPrompt("search_legal_topic", map[string]string{"topic": "t"})
	all += s.getPrompt("compare_law_versions", map[string]string{"norm_id": "1", "from_date": "d1", "to_date": "d2"})
	all += s.getPrompt("trace_law_history", map[string]string{"norm_id": "1"})
	all += s.getPrompt("check_law_validity", map[string]string{"norm_id": "1"})
	all += s.getPrompt("explain_law_simply", map[string]string{"norm_id": "1"})
	all += s.getPrompt("law_research_workflow", map[string]string{"norm_id": "1"})
	all += s.getPrompt("answer_constitutional_question", map[string]string{"question": "q"})
	all += s.getPrompt("check_norm_constitutionality", map[string]string{"norm_id": "1"})

	for _, tool := range ToolNames() {
		s.True(strings.Contains(all, tool), "tool %s not referenced by any template", tool)
	}
	// No template mentions an unregistered tool name pattern (tool_name( ).
	for _, candidate := range []string{"get_norma", "fetch_law", "search_norms"} {
		s.False(strings.Contains(all, candidate), "template references unregistered tool %q", candidate)
	}
}

func (s *PromptsSuite) TestAnswerConstitutionalQuestionInjectsQuestion() {
	text := s.getPrompt("answer_constitutional_question", map[string]string{"question": "¿qué dice sobre el derecho de propiedad?"})
	s.Contains(text, "¿qué dice sobre el derecho de propiedad?")
	s.Contains(text, "get_law_summary(norm_id=242302)")
	s.Contains(text, "section_id=<section id>")
	s.Contains(text, "242302")
	s.Contains(text, "DISPOSICIONES TRANSITORIAS")
	s.Contains(text, "NEVER invent articles")
	s.Contains(text, "not legal advice")
	s.Contains(text, "Tribunal Constitucional")
	s.Contains(text, "podría interpretarse como")
}

func (s *PromptsSuite) TestAnswerConstitutionalQuestionWithHintAndVersion() {
	text := s.getPrompt("answer_constitutional_question", map[string]string{"question": "q", "article_hint": "19 Nº24", "version_date": "2019-01-01"})
	s.Contains(text, "19 Nº24")
	s.Contains(text, "version_date=2019-01-01")
	s.Contains(text, "Version: as of")

	withoutVersion := s.getPrompt("answer_constitutional_question", map[string]string{"question": "q"})
	s.NotContains(withoutVersion, "version_date=")
}

func (s *PromptsSuite) TestCheckNormConstitutionalityInjectsNormID() {
	text := s.getPrompt("check_norm_constitutionality", map[string]string{"norm_id": "1195666"})
	s.Contains(text, "1195666")
	s.Contains(text, "242302")
	s.Contains(text, "get_law_summary(norm_id=1195666")
	s.Contains(text, "get_law_summary(norm_id=242302")
	s.Contains(text, "section_id=<section id>")
	s.Contains(text, "get_law_history")
	s.Contains(text, "not legal advice")
	s.Contains(text, "Tribunal Constitucional")
}

func (s *PromptsSuite) TestCheckNormConstitutionalityWithQuestionAndVersion() {
	text := s.getPrompt("check_norm_constitutionality", map[string]string{"norm_id": "1195666", "question": "¿vulnera igualdad?", "version_date": "2020-06-01"})
	s.Contains(text, "1195666")
	s.Contains(text, "¿vulnera igualdad?")
	s.Contains(text, "version_date=2020-06-01")
	s.Contains(text, "Version: as of")
	s.Contains(text, "DISPOSICIONES TRANSITORIAS")

	withoutVersion := s.getPrompt("check_norm_constitutionality", map[string]string{"norm_id": "1195666"})
	s.NotContains(withoutVersion, "version_date=")
}

func (s *PromptsSuite) TestConstitutionalPromptsMentionHedgeAndDisclaimer() {
	a := s.getPrompt("answer_constitutional_question", map[string]string{"question": "q"})
	s.Contains(a, "podría interpretarse como")
	s.Contains(a, "general information, not legal advice")

	c := s.getPrompt("check_norm_constitutionality", map[string]string{"norm_id": "1"})
	s.Contains(c, "podría interpretarse como")
	s.Contains(c, "not legal advice")
	s.Contains(c, "Art. X")
}

func (s *PromptsSuite) TestLoadEmbeddedParsesTenPrompts() {
	ps, err := LoadEmbedded()
	s.Require().NoError(err)
	s.Len(ps.Templates, 10)
	for _, name := range expectedPromptNames {
		_, ok := ps.Templates[name]
		s.True(ok, "embedded prompt %q missing", name)
	}
}

func (s *PromptsSuite) TestLoadEmbeddedRejectsUnknownPlaceholder() {
	// Build a YAML with 9 valid prompts but one contains an unknown placeholder.
	yamlWithBad := "prompts:\n"
	for _, name := range expectedPromptNames {
		content := "valid content"
		if name == "analyze_law" {
			content = "bad {{.unknown_placeholder}}"
		}
		yamlWithBad += "  " + name + ": |\n    " + content + "\n"
	}
	_, err := loadFromBytes([]byte(yamlWithBad))
	s.Error(err)
	s.Contains(err.Error(), "unknown placeholder")
	s.Contains(err.Error(), "unknown_placeholder")
}

func (s *PromptsSuite) TestLoadEmbeddedRejectsWrongCount() {
	// Only one prompt — should fail count validation.
	yamlOne := "prompts:\n  analyze_law: |\n    hello\n"
	_, err := loadFromBytes([]byte(yamlOne))
	s.Error(err)
	s.Contains(err.Error(), "want 10 prompts")
}

func (s *PromptsSuite) TestRenderWithMissingArgsStillServes() {
	// Directly test PromptSet render with missing optional args — should not error and not leak placeholder syntax.
	ps, err := LoadEmbedded()
	s.Require().NoError(err)
	text, err := ps.Render("check_law_validity", map[string]string{"norm_id": "42"}, allowedPlaceholders, toolVars())
	s.Require().NoError(err)
	s.Contains(text, "42")
	s.NotContains(text, "{{")
	s.NotContains(text, "unknown")
}
