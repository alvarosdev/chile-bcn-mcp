// This file is the canonical catalog of norm types, mapping the numeric
// codes used by the BCN API (e.g. tipos_numeros[].tipo = "1") to the
// official abbreviation and name.
//
// Source: GET /Consulta/getTiposNorma of nuevo.leychile.cl — the catalog
// the LeyChile frontend itself uses. Hardcoded by design decision (the
// catalog is stable; new types degrade gracefully via omitempty on the
// canonical fields, raw values stay intact).
package bcn

// normType is one entry of the norm-type catalog.
type normType struct {
	Abbr  string // official abbreviation (LEY, DTO, RES...)
	Valor string // official name (Ley, Decreto, Resolución...)
}

// normTypeCatalog maps cod → normType. NOTE: "tipo" codes inside
// get_historias_de_ley (1/3/4) are history-group types, NOT norm types —
// this catalog must not be applied there.
var normTypeCatalog = map[int]normType{
	0:  {Abbr: "OTR", Valor: "OTRO"},
	1:  {Abbr: "LEY", Valor: "Ley"},
	2:  {Abbr: "DTO", Valor: "Decreto"},
	3:  {Abbr: "RES", Valor: "Resolución"},
	4:  {Abbr: "AA", Valor: "Auto Acordado"},
	5:  {Abbr: "ACD", Valor: "Acuerdo"},
	6:  {Abbr: "ALC", Valor: "Alcance"},
	7:  {Abbr: "AVI", Valor: "Aviso"},
	8:  {Abbr: "CER", Valor: "Certificado"},
	9:  {Abbr: "CIR", Valor: "Circular"},
	10: {Abbr: "COD", Valor: "Código"},
	12: {Abbr: "CV", Valor: "Convenio"},
	13: {Abbr: "DFL", Valor: "Decreto con Fuerza de Ley"},
	14: {Abbr: "DIC", Valor: "Dictamen"},
	15: {Abbr: "DL", Valor: "Decreto Ley"},
	16: {Abbr: "INS", Valor: "Instrucción"},
	17: {Abbr: "LEI", Valor: "Norma Antigua"},
	18: {Abbr: "NTF", Valor: "Notificación"},
	19: {Abbr: "OF", Valor: "Oficio"},
	20: {Abbr: "ORD", Valor: "Orden"},
	21: {Abbr: "ORZ", Valor: "Ordenanza"},
	22: {Abbr: "SEN", Valor: "Sentencia"},
	23: {Abbr: "SES", Valor: "Sesión"},
	24: {Abbr: "REC", Valor: "Rectificación"},
	25: {Abbr: "RRE", Valor: "Reunión extraordinaria"},
	26: {Abbr: "RRO", Valor: "Reunión ordinaria"},
	27: {Abbr: "RRA", Valor: "RRA"},
	28: {Abbr: "S/N", Valor: "Sin número"},
	29: {Abbr: "TRA", Valor: "Tratado"},
	30: {Abbr: "CTR", Valor: "Constitución de la República"},
	31: {Abbr: "CCI", Valor: "Carta Circular"},
	32: {Abbr: "TCI", Valor: "Telegrama Circular"},
	33: {Abbr: "LEI", Valor: "Lei"},
	34: {Abbr: "NCG", Valor: "Norma de Carácter General"},
	35: {Abbr: "MSJ", Valor: "Mensaje"},
	36: {Abbr: "Bando", Valor: "Bando"},
	37: {Abbr: "SC", Valor: "Senado Consulto"},
	38: {Abbr: "RM", Valor: "Reglamento Municipal"},
	39: {Abbr: "CPR", Valor: "Constitución Política de la República"},
}

// canonicalNormType resolves a numeric norm-type code to the official
// name and abbreviation. ok is false for codes outside the catalog.
func canonicalNormType(cod int) (valor, abbr string, ok bool) {
	t, ok := normTypeCatalog[cod]
	if !ok {
		return "", "", false
	}
	return t.Valor, t.Abbr, true
}
