package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"dev.alvaros.chile-bcn-mcp/internal/bcn"
)

// GetLawSummaryArgs carries the arguments of the get_law_summary tool.
type GetLawSummaryArgs struct {
	NormID int64 `json:"norm_id" jsonschema:"the norm id (norm_id) from search_laws results"`
}

// RegisterGetLawSummary registers the get_law_summary tool on the MCP server.
func RegisterGetLawSummary(srv *mcp.Server, client bcn.LawClient) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "get_law_summary",
		Description: "Get a lightweight summary of a Chilean law, decree or resolution by its " +
			"norm_id (from search_laws): title, source, matters, norm categories and the " +
			"official BCN summary. Use it to understand what a norm is about before deciding " +
			"to read the full content with get_law.",
	}, makeGetLawSummary(client))
}

func makeGetLawSummary(client bcn.LawClient) mcp.ToolHandlerFor[GetLawSummaryArgs, bcn.NormaSummary] {
	return func(ctx context.Context, req *mcp.CallToolRequest, args GetLawSummaryArgs) (*mcp.CallToolResult, bcn.NormaSummary, error) {
		if args.NormID <= 0 {
			return errorResult("norm_id must be a positive number"), bcn.NormaSummary{}, nil
		}

		summary, err := client.GetNormaSummary(ctx, args.NormID)
		if err != nil {
			if errors.Is(err, bcn.ErrNormaNotFound) {
				return errorResult(fmt.Sprintf("norma not found: norm_id %d does not exist in LeyChile", args.NormID)), bcn.NormaSummary{}, nil
			}
			return errorResult(fmt.Sprintf("get law summary failed: %v", err)), bcn.NormaSummary{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: formatNormaSummary(summary)},
			},
		}, summary, nil
	}
}

// formatNormaSummary renders the summary for the LLM. The official BCN
// summary is short by nature, so it goes complete in the text view.
func formatNormaSummary(s bcn.NormaSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", s.TituloNorma)
	fmt.Fprintf(&b, "Source: %s\n", s.Fuente)
	if len(s.Materias) > 0 {
		fmt.Fprintf(&b, "Matters: %s\n", strings.Join(s.Materias, ", "))
	}
	if len(s.CategoriasNorma) > 0 {
		fmt.Fprintf(&b, "Categories: %s\n", strings.Join(s.CategoriasNorma, ", "))
	}
	for _, r := range s.Resumenes {
		fmt.Fprintf(&b, "\nSummary:\n%s\n", r)
	}
	return b.String()
}
