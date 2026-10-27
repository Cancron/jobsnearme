package skilltag

import (
	"slices"
	"testing"
)

// TestParse_ModernAIAndEnterpriseSkills covers Odoo enterprise ERP and modern AI/data tooling.
func TestParse_ModernAIAndEnterpriseSkills(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		want   []string
		absent []string
	}{
		{
			name: "enterprise erp and crm",
			in:   "Senior Python developer needed to build custom Odoo ERP modules, integrate PostgreSQL, and extend inventory workflows.",
			want: []string{"odoo", "python", "postgresql", "erp"},
		},
		{
			name: "olap database duckdb",
			in:   "Data engineer building data pipelines with DuckDB and Parquet files for fast local analytical processing.",
			want: []string{"duckdb", "data-pipelines"},
		},
		{
			name: "llm proxy and serving",
			in:   "Deploying open-source models with SGLang and routing inference traffic through LiteLLM across multiple backends.",
			want: []string{"sglang", "litellm"},
		},
		{
			name: "foundation models and prompt optimization",
			in:   "Research engineer fine-tuning DeepSeek models and using DSPy to optimize multi-stage agent prompt pipelines.",
			want: []string{"deepseek", "dspy"},
		},
		{
			name: "typescript agent workflows",
			in:   "Building production agentic applications using Mastra with Node.js and TypeScript services.",
			want: []string{"mastra", "nodejs", "typescript"},
		},
		{
			name: "space separated variants resolve to canonical",
			in:   "Querying duck db tables while calling deep seek endpoints via lite llm and sg lang.",
			want: []string{"duckdb", "deepseek", "litellm", "sglang"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Parse(tc.in)
			for _, w := range tc.want {
				if !slices.Contains(got, w) {
					t.Errorf("Parse(%q) = %v, want it to contain %q", tc.in, got, w)
				}
			}
			for _, a := range tc.absent {
				if slices.Contains(got, a) {
					t.Errorf("Parse(%q) = %v, want it NOT to contain %q", tc.in, got, a)
				}
			}
		})
	}
}
