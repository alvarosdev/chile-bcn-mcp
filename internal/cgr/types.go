package cgr

// SearchParams carries the arguments for a Contraloría dictamen search.
type SearchParams struct {
	Query       string
	ExactSearch bool
	Order       string // "date" | "dateasc" | "score"
	Page        int    // 1-indexed for caller (0 means default 1)
}

// SearchResponse is the paginated result of a Contraloría dictamen search.
type SearchResponse struct {
	Results    []DictamenSummary `json:"results"`
	Pagination Pagination        `json:"pagination"`
}

// DictamenSummary is the lightweight projection of a dictamen — the fields
// that answer "what is this dictamen about", without the full document.
// Mirrors the minimal + carácter decision from design.
type DictamenSummary struct {
	DictamenID   string `json:"dictamen_id"`
	NDictamen    string `json:"n_dictamen"`
	NumericID    string `json:"numeric_doc_id"`
	FechaDoc     string `json:"fecha_documento"`
	Materia      string `json:"materia"`
	Descriptores string `json:"descriptores"`
	Criterio     string `json:"criterio"`
	Origen       string `json:"origen"`
	Caracter     string `json:"caracter"`
	URL          string `json:"url"`
	PDFURL       string `json:"pdf_url"`
}

// DictamenFull is the full structured content of one dictamen, including
// the sanitized document. DictamenSummary is embedded for reuse (parity with
// NormaSummary projection in bcn).
type DictamenFull struct {
	DictamenSummary `json:",inline"`
	Destinatarios   string `json:"destinatarios"`
	Abogados        string `json:"abogados"`
	FuentesLegales  string `json:"fuentes_legales"`
	Documento       string `json:"documento_completo"`
	CharCount       int    `json:"char_count"`
}

// Pagination reports the total result count so callers can page through.
// Page is 1-indexed, PageSize always 20 for CGR.
type Pagination struct {
	Total      int  `json:"total"`
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalPages int  `json:"total_pages"`
	HasMore    bool `json:"has_more"`
}

// CountResponse is the aggregated count cross-type for a query.
type CountResponse struct {
	Query   string        `json:"query"`
	Total   int           `json:"total"`
	Buckets []CountBucket `json:"buckets"`
}

// CountBucket is one entry of the count_by_type aggregation.
type CountBucket struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// Wire types for Elasticsearch responses — internal, not exposed in
// structuredContent. They decode the ES envelope and are projected to the
// public types above.

type cgrSearchResponse struct {
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []cgrHit `json:"hits"`
	} `json:"hits"`
}

type cgrHit struct {
	ID     string    `json:"_id"`
	Score  float64   `json:"_score"`
	Source cgrSource `json:"_source"`
}

type cgrSource struct {
	DocID          string `json:"doc_id"`
	NDictamen      string `json:"n_dictamen"`
	NumericDocID   string `json:"numeric_doc_id"`
	FechaDocumento string `json:"fecha_documento"`
	Carater        string `json:"carácter"`
	Materia        string `json:"materia"`
	Descriptores   string `json:"descriptores"`
	Criterio       string `json:"criterio"`
	Origen         string `json:"origen_"`
	Destinatarios  string `json:"destinatarios"`
	Abogados       string `json:"abogados"`
	FuentesLegales string `json:"fuentes_legales"`
	Documento      string `json:"documento_completo"`
}

type cgrCountResponse struct {
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
	} `json:"hits"`
	Aggregations struct {
		CountByType struct {
			Buckets []struct {
				Key      string `json:"key"`
				DocCount int    `json:"doc_count"`
			} `json:"buckets"`
		} `json:"count_by_type"`
	} `json:"aggregations"`
}
