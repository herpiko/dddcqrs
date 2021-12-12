package query

// Struct helper for elastic raw query

type MultiMatch struct {
	Fields    []string `json:"fields,omitempty"`
	Query     string   `json:"query,omitempty"`
	Fuzziness string   `json:"fuzziness,omitempty"`
}

type Query struct {
	MultiMatch MultiMatch `json:"multi_match,omitempty"`
}

type SortItem struct {
	CreatedAt SortOpt `json:"created_at,omitempty"`
}

type SortOpt struct {
	Order string `json:"order,omitempty"`
}

type ElRawQuery struct {
	Sort  []SortItem `json:"sort,omitempty"`
	From  string     `json:"from,omitempty"`
	Size  string     `json:"size,omitempty"`
	Query *Query     `json:"query,omitempty"`
}
