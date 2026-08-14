package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"dev.alvaros.chile-bcn-mcp/internal/bcn"
)

// GetLawArgs carries the arguments of the get_law tool. VersionDate and
// StructureOnly are optional (omitempty keeps them out of the required
// schema).
type GetLawArgs struct {
	NormID        int64  `json:"norm_id" jsonschema:"the norm id (norm_id) from search_laws results"`
	VersionDate   string `json:"version_date,omitempty" jsonschema:"version in force at this date (YYYY-MM-DD, optional — defaults to the latest version)"`
	StructureOnly bool   `json:"structure_only,omitempty" jsonschema:"return metadata and table of contents only, without the full content (default false)"`
}

// StructurePartOut is one flattened entry of the nested norm structure.
// The API tree is recursive; the output flattens it with a depth field so
// the JSON schema stays acyclic and consumers can rebuild the tree.
type StructurePartOut struct {
	Name  string `json:"name"`
	ID    int64  `json:"id"`
	Type  int    `json:"type,omitempty"`
	Depth int    `json:"depth"`
}

// GetLawOutput is the structured content of get_law. Content is omitted
// when the norm was requested with structure_only; VersionDate echoes the
// requested historical version (empty = latest).
type GetLawOutput struct {
	Metadatos   bcn.Metadatos      `json:"metadatos"`
	Estructura  []StructurePartOut `json:"estructura"`
	Proyectos   []bcn.Proyecto     `json:"proyectos"`
	Content     string             `json:"content,omitempty"`
	VersionDate string             `json:"version_date,omitempty"`
}

// RegisterGetLaw registers the get_law tool on the MCP server.
func RegisterGetLaw(srv *mcp.Server, client bcn.LawClient) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "get_law",
		Description: "Get the full content of a Chilean law, decree or resolution by its norm_id " +
			"(from search_laws). Returns metadata, the table of contents and the complete text " +
			"in Markdown. Use structure_only=true to explore a long norm without its content.",
	}, makeGetLaw(client))
}

func makeGetLaw(client bcn.LawClient) mcp.ToolHandlerFor[GetLawArgs, GetLawOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, args GetLawArgs) (*mcp.CallToolResult, GetLawOutput, error) {
		if args.NormID <= 0 {
			return errorResult("norm_id must be a positive number"), GetLawOutput{}, nil
		}
		if err := validateVersionDate(args.VersionDate); err != nil {
			return errorResult(err.Error()), GetLawOutput{}, nil
		}

		query := bcn.NormaQuery{NormID: args.NormID, VersionDate: args.VersionDate}
		norma, err := client.GetNorma(ctx, query)
		if err != nil {
			if errors.Is(err, bcn.ErrNormaNotFound) {
				return errorResult(fmt.Sprintf("norma not found: norm_id %d does not exist in LeyChile", args.NormID)), GetLawOutput{}, nil
			}
			return errorResult(fmt.Sprintf("get law failed: %v", err)), GetLawOutput{}, nil
		}

		output := buildGetLawOutput(norma, args)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: formatNorma(norma, args)},
			},
		}, output, nil
	}
}

// buildGetLawOutput projects the norm into the structured output. Content
// (the Markdown body) is omitted when structureOnly is set.
func buildGetLawOutput(norma bcn.NormaFull, args GetLawArgs) GetLawOutput {
	var estructura []StructurePartOut
	flattenStructure(norma.Estructura, 0, &estructura)
	out := GetLawOutput{
		Metadatos:   norma.Metadatos,
		Estructura:  estructura,
		Proyectos:   norma.Proyectos,
		VersionDate: args.VersionDate,
	}
	if !args.StructureOnly {
		out.Content = normaContentMarkdown(norma)
	}
	return out
}

// flattenStructure walks the nested structure tree into a depth-annotated
// flat list (schema-safe for the structured output).
func flattenStructure(parts []bcn.EstructuraPart, depth int, out *[]StructurePartOut) {
	for _, p := range parts {
		*out = append(*out, StructurePartOut{Name: p.N, ID: p.I, Type: p.T, Depth: depth})
		flattenStructure(p.H, depth+1, out)
	}
}

// normaContentMarkdown assembles the content section of a norm. Blocks nest
// (titles contain articles): each block gets a heading at its depth and its
// own text only when it has no children — a title block's body repeats the
// title, so it is skipped in favor of its children.
func normaContentMarkdown(norma bcn.NormaFull) string {
	var b strings.Builder
	renderBlocks(&b, norma.Html, 0)
	return b.String()
}

// renderBlocks walks the block tree, heading depth grows with nesting
// (max ######).
func renderBlocks(b *strings.Builder, blocks []bcn.HtmlBlock, depth int) {
	for _, block := range blocks {
		heading := block.SectionName
		if heading == "" {
			heading = fmt.Sprintf("Section %d", block.I)
		}
		prefix := "#"
		if depth < 3 {
			prefix = strings.Repeat("#", 3+depth)
		} else {
			prefix = "######"
		}
		fmt.Fprintf(b, "%s %s\n\n", prefix, heading)
		if len(block.H) == 0 {
			b.WriteString(block.Markdown)
			b.WriteString("\n\n")
		}
		renderBlocks(b, block.H, depth+1)
	}
}

// formatNorma renders a norm for the LLM: metadata header, related bills,
// table of contents, and (unless structureOnly) the full Markdown content.
// When a historical version was requested, the header states it.
func formatNorma(norma bcn.NormaFull, args GetLawArgs) string {
	var b strings.Builder
	m := norma.Metadatos

	fmt.Fprintf(&b, "# %s\n\n", m.TituloNorma)
	if args.VersionDate != "" {
		fmt.Fprintf(&b, "Version: as of %s\n", args.VersionDate)
	}
	fmt.Fprintf(&b, "Type: %s · Source: %s N° %s\n",
		formatTipos(m.TiposNumeros), m.Fuente, m.NumeroFuente)
	fmt.Fprintf(&b, "Organism: %s\n", strings.Join(m.Organismos, "; "))
	fmt.Fprintf(&b, "Published: %s · In force from: %s",
		m.FechaPublicacion, m.Vigencia.InicioVigencia)
	if m.Vigencia.FinVigencia != "" {
		fmt.Fprintf(&b, " to %s", m.Vigencia.FinVigencia)
	}
	fmt.Fprintf(&b, "\nDerogated: %t\n", m.Derogado)
	if len(m.Materias) > 0 {
		fmt.Fprintf(&b, "Subjects: %s\n", strings.Join(m.Materias, ", "))
	}
	if len(m.CategoriasNorma) > 0 {
		fmt.Fprintf(&b, "Categories: %s\n", strings.Join(m.CategoriasNorma, ", "))
	}
	if len(m.Vinculaciones) > 0 {
		links := make([]string, 0, len(m.Vinculaciones))
		for _, v := range m.Vinculaciones {
			links = append(links, v.Text)
		}
		fmt.Fprintf(&b, "Related norms: %s\n", strings.Join(links, "; "))
	}
	if len(m.Resumenes) > 0 {
		fmt.Fprintf(&b, "\nSummary: %s\n", truncate(m.Resumenes[0], 1200))
	}

	if len(norma.Proyectos) > 0 {
		b.WriteString("\n## Related bills\n")
		for _, p := range norma.Proyectos {
			for _, pl := range p.Pls {
				fmt.Fprintf(&b, "- %s — %s\n", pl.NroBoletin, pl.Informacion)
				if pl.Enlace != "" {
					fmt.Fprintf(&b, "  %s\n", pl.Enlace)
				}
			}
		}
	}

	b.WriteString("\n## Structure\n")
	renderStructure(&b, norma.Estructura, 0)

	if !args.StructureOnly {
		b.WriteString("\n## Content\n\n")
		b.WriteString(normaContentMarkdown(norma))
	}

	return b.String()
}

// renderStructure renders the nested table of contents as an indented list.
func renderStructure(b *strings.Builder, parts []bcn.EstructuraPart, depth int) {
	for _, part := range parts {
		fmt.Fprintf(b, "%s- %s\n", strings.Repeat("  ", depth), part.N)
		renderStructure(b, part.H, depth+1)
	}
}

// validateVersionDate enforces the strict YYYY-MM-DD format. The API
// silently ignores malformed dates (returns the latest version), so the
// tool fails fast instead of letting the model believe it read the
// requested version.
func validateVersionDate(versionDate string) error {
	if versionDate == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", versionDate); err != nil {
		return fmt.Errorf("version_date must be a valid date in YYYY-MM-DD format, got %q", versionDate)
	}
	return nil
}

// formatTipos renders the norm type list as "Ley 21600; Decreto 1".
func formatTipos(tipos []bcn.TipoNumero) string {
	parts := make([]string, 0, len(tipos))
	for _, t := range tipos {
		parts = append(parts, fmt.Sprintf("%s %s", t.Descripcion, t.Numero))
	}
	return strings.Join(parts, "; ")
}
