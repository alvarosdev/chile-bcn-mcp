package cgr

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"
)

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
	clientSession, err := client.Connect(s.ctx, clientTransport, nil)
	s.Require().NoError(err)
	s.T().Cleanup(func() { clientSession.Close() })
	s.session = clientSession
}

func (s *PromptsSuite) getPrompt(name string, args map[string]string) string {
	res, err := s.session.CallTool(s.ctx, &mcp.CallToolParams{Name: "none", Arguments: map[string]any{}}) // dummy to avoid unused
	_ = res
	_ = err
	// Use direct render via PromptSet for unit test
	ps, _ := LoadEmbedded()
	text, err := ps.Render(name, args, allowedPlaceholders, toolVars())
	s.Require().NoError(err)
	return text
}

func (s *PromptsSuite) TestListPrompts() {
	ps, err := LoadEmbedded()
	s.Require().NoError(err)
	s.Len(ps.Templates, 4)
	for _, name := range expectedPromptNames {
		_, ok := ps.Templates[name]
		s.True(ok, "prompt %q missing", name)
	}
}

func (s *PromptsSuite) TestSearchJurisprudenceInjectsQuery() {
	text := s.getPrompt("search_jurisprudence", map[string]string{"query": "quillota"})
	s.Contains(text, "quillota")
	s.Contains(text, "count_cgr_jurisprudencia")
	s.Contains(text, "search_cgr_dictamenes")
}

func (s *PromptsSuite) TestAnalyzeDictamenInjectsID() {
	text := s.getPrompt("analyze_dictamen", map[string]string{"dictamen_id": "E179593N25"})
	s.Contains(text, "E179593N25")
	s.Contains(text, "get_cgr_dictamen")
}

func (s *PromptsSuite) TestTemplatesReferenceOnlyRegisteredTools() {
	ps, err := LoadEmbedded()
	s.Require().NoError(err)
	allowedTools := make(map[string]bool)
	for _, t := range ToolNames() {
		allowedTools[t] = true
	}
	for name, tmpl := range ps.Templates {
		text := tmpl.Root.String()
		// Simple check: template should reference at least one allowed tool
		found := false
		for tool := range allowedTools {
			if strings.Contains(text, tool) {
				found = true
				break
			}
		}
		s.True(found, "prompt %q should reference a registered tool", name)
	}
}

func (s *PromptsSuite) TestLoadEmbeddedParsesFourPrompts() {
	ps, err := LoadEmbedded()
	s.Require().NoError(err)
	s.Len(ps.Templates, 4)
}

func (s *PromptsSuite) TestLoadEmbeddedRejectsUnknownPlaceholder() {
	yamlWithBad := "prompts:\n"
	for _, name := range expectedPromptNames {
		content := "valid content"
		if name == "search_jurisprudence" {
			content = "bad {{.unknown_placeholder}}"
		}
		yamlWithBad += "  " + name + ": |\n    " + content + "\n"
	}
	_, err := loadFromBytes([]byte(yamlWithBad))
	s.Error(err)
	s.Contains(err.Error(), "unknown placeholder")
}

func (s *PromptsSuite) TestRenderWithMissingArgsStillServes() {
	ps, err := LoadEmbedded()
	s.Require().NoError(err)
	text, err := ps.Render("analyze_dictamen", map[string]string{"dictamen_id": "E179593N25"}, allowedPlaceholders, toolVars())
	s.Require().NoError(err)
	s.Contains(text, "E179593N25")
	s.NotContains(text, "{{")
}
