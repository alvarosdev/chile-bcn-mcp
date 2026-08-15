package prompts

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

	server := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	RegisterPrompts(server)

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
		"law_research_workflow",
	}
	for _, name := range expected {
		p, ok := names[name]
		s.Require().True(ok, "prompt %s missing from prompts/list", name)
		s.NotEmpty(p.Description, "prompt %s must have a description", name)

		required := map[string]bool{}
		for _, a := range p.Arguments {
			required[a.Name] = a.Required
		}
		if name != "search_legal_topic" {
			s.True(required["norm_id"], "prompt %s: norm_id must be required", name)
		} else {
			s.True(required["topic"], "prompt %s: topic must be required", name)
		}
	}
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

	for _, tool := range ToolNames() {
		s.True(strings.Contains(all, tool), "tool %s not referenced by any template", tool)
	}
	// No template mentions an unregistered tool name pattern (tool_name( ).
	for _, candidate := range []string{"get_norma", "fetch_law", "search_norms"} {
		s.False(strings.Contains(all, candidate), "template references unregistered tool %q", candidate)
	}
}
