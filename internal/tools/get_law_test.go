package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"dev.alvaros.chile-bcn-mcp/internal/bcn"
)

// GetLawSuite validates the get_law tool against a MockLawClient: the BCN
// API is never reached.
type GetLawSuite struct {
	suite.Suite
	ctx       context.Context
	lawClient *bcn.MockLawClient
	session   *mcp.ClientSession
}

func TestGetLawSuite(t *testing.T) {
	suite.Run(t, new(GetLawSuite))
}

func (s *GetLawSuite) SetupTest() {
	s.ctx = context.Background()
	s.lawClient = bcn.NewMockLawClient(s.T())
	s.session = newTestClient(s.T(), s.ctx, s.lawClient)
}

func (s *GetLawSuite) callTool(args map[string]any) (*mcp.CallToolResult, error) {
	return s.session.CallTool(s.ctx, &mcp.CallToolParams{
		Name:      "get_law",
		Arguments: args,
	})
}

func (s *GetLawSuite) sampleNorma() bcn.NormaFull {
	return bcn.NormaFull{
		Html: []bcn.HtmlBlock{
			{I: 1, T: "<div>LEY N&#xDA;M. 21.600</div>", Markdown: "LEY NÚM. 21.600", SectionName: "Encabezado"},
			{I: 2, T: "<div>Título I</div>", Markdown: "Disposiciones generales.", SectionName: "TÍTULO I"},
		},
		Estructura: []bcn.EstructuraPart{
			{N: "Encabezado", I: 1},
			{N: "TÍTULO I", I: 2},
		},
		Proyectos: []bcn.Proyecto{{
			Categoria: "Proyecto original",
			Pls: []struct {
				Enlace      string `json:"enlace"`
				Informacion string `json:"informacion"`
				NroBoletin  string `json:"nroBoletin"`
			}{{NroBoletin: "18156-04", Informacion: "Crea el servicio", Enlace: "https://senado.cl/tramitacion"}},
		}},
		Metadatos: bcn.Metadatos{
			TiposNumeros:     []bcn.TipoNumero{{Numero: "21600", Descripcion: "Ley"}},
			Organismos:       []string{"MINISTERIO DEL MEDIO AMBIENTE"},
			TituloNorma:      "CREA EL SERVICIO DE BIODIVERSIDAD",
			Fuente:           "Diario Oficial",
			NumeroFuente:     "43646",
			Materias:         []string{"Biodiversidad"},
			FechaPublicacion: "2023-09-06",
			Vigencia:         bcn.Vigencia{InicioVigencia: "2025-09-29"},
			Vinculaciones:    []bcn.Vinculacion{{Text: "MODIFICACION"}},
			Resumenes:        []string{"La presente ley crea el Servicio."},
		},
	}
}

func (s *GetLawSuite) TestGetLawFullContent() {
	s.lawClient.EXPECT().GetNorma(mock.Anything, int64(1195666)).Return(s.sampleNorma(), nil).Once()

	res, err := s.callTool(map[string]any{"norm_id": 1195666})
	s.Require().NoError(err)
	s.False(res.IsError)
	text := res.Content[0].(*mcp.TextContent).Text

	s.Contains(text, "# CREA EL SERVICIO DE BIODIVERSIDAD")
	s.Contains(text, "Type: Ley 21600 · Source: Diario Oficial N° 43646")
	s.Contains(text, "Derogated: false")
	s.Contains(text, "## Related bills")
	s.Contains(text, "18156-04")
	s.Contains(text, "## Structure")
	s.Contains(text, "- TÍTULO I")
	s.Contains(text, "## Content")
	s.Contains(text, "### Encabezado")
	s.Contains(text, "LEY NÚM. 21.600")

	// Structured content: typed, complete metadata + content included.
	sc, ok := res.StructuredContent.(map[string]any)
	s.Require().True(ok, "structuredContent expected, got %T", res.StructuredContent)
	metadatos := sc["metadatos"].(map[string]any)
	s.Equal("CREA EL SERVICIO DE BIODIVERSIDAD", metadatos["titulo_norma"])
	s.Contains(sc["content"], "### Encabezado")
}

func (s *GetLawSuite) TestGetLawStructureOnlyOmitsContent() {
	s.lawClient.EXPECT().GetNorma(mock.Anything, int64(1195666)).Return(s.sampleNorma(), nil).Once()

	res, err := s.callTool(map[string]any{"norm_id": 1195666, "structure_only": true})
	s.Require().NoError(err)
	text := res.Content[0].(*mcp.TextContent).Text

	s.Contains(text, "## Structure")
	s.NotContains(text, "## Content")

	// Structured content with structure_only: content field omitted.
	sc, ok := res.StructuredContent.(map[string]any)
	s.Require().True(ok, "structuredContent expected, got %T", res.StructuredContent)
	_, hasContent := sc["content"]
	s.False(hasContent, "content must be omitted with structure_only")
	s.NotEmpty(sc["estructura"])
}

func (s *GetLawSuite) TestGetLawInvalidNormIDFailsWithoutCallingClient() {
	// No expectations on the mock: the handler must not call GetNorma.
	res, err := s.callTool(map[string]any{"norm_id": 0})
	s.Require().NoError(err)
	s.True(res.IsError)
	s.Contains(res.Content[0].(*mcp.TextContent).Text, "norm_id must be a positive number")

	res, err = s.callTool(map[string]any{})
	s.Require().NoError(err)
	s.True(res.IsError)
}

func (s *GetLawSuite) TestGetLawNotFoundSurfacesClearMessage() {
	s.lawClient.EXPECT().GetNorma(mock.Anything, int64(999999999)).Return(bcn.NormaFull{}, bcn.ErrNormaNotFound).Once()

	res, err := s.callTool(map[string]any{"norm_id": 999999999})
	s.Require().NoError(err)
	s.True(res.IsError)
	s.Contains(res.Content[0].(*mcp.TextContent).Text, "norma not found: norm_id 999999999")
}

func (s *GetLawSuite) TestGetLawGenericErrorSurfaces() {
	s.lawClient.EXPECT().GetNorma(mock.Anything, int64(1)).Return(bcn.NormaFull{}, errors.New("circuit breaker open")).Once()

	res, err := s.callTool(map[string]any{"norm_id": 1})
	s.Require().NoError(err)
	s.True(res.IsError)
	s.Contains(res.Content[0].(*mcp.TextContent).Text, "get law failed")
}
