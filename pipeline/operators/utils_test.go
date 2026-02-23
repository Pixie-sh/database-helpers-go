package operators

import (
	"testing"
)

func TestParseQueryStringGrouped(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantNodeCount  int
		wantGroupCount int
		wantLeafCount  int
		checkFunc      func(t *testing.T, nodes []queryNode)
	}{
		{
			name:          "flat query backward compat - single term",
			input:         "field1:value1",
			wantNodeCount: 1,
			checkFunc: func(t *testing.T, nodes []queryNode) {
				if nodes[0].Type != queryNodeLeaf {
					t.Errorf("expected leaf node, got %v", nodes[0].Type)
				}
				if nodes[0].Part.Query != "field1:value1" {
					t.Errorf("expected query 'field1:value1', got '%s'", nodes[0].Part.Query)
				}
			},
		},
		{
			name:          "flat query backward compat - multiple terms with AND",
			input:         "field1:val1{AND}field2:val2",
			wantNodeCount: 2,
			checkFunc: func(t *testing.T, nodes []queryNode) {
				if nodes[0].Part.Query != "field1:val1" {
					t.Errorf("expected 'field1:val1', got '%s'", nodes[0].Part.Query)
				}
				if nodes[0].Aggregator != "{AND}" {
					t.Errorf("expected aggregator '{AND}', got '%s'", nodes[0].Aggregator)
				}
				if nodes[1].Part.Query != "field2:val2" {
					t.Errorf("expected 'field2:val2', got '%s'", nodes[1].Part.Query)
				}
			},
		},
		{
			name:          "flat query backward compat - OR aggregator",
			input:         "field1:val1{OR}field1:val2",
			wantNodeCount: 2,
			checkFunc: func(t *testing.T, nodes []queryNode) {
				if nodes[0].Aggregator != "{OR}" {
					t.Errorf("expected aggregator '{OR}', got '%s'", nodes[0].Aggregator)
				}
			},
		},
		{
			name:          "simple group with OR",
			input:         "{P}field1:val1{|}field1:val2{/P}",
			wantNodeCount: 1, // one group node
			checkFunc: func(t *testing.T, nodes []queryNode) {
				if nodes[0].Type != queryNodeGroup {
					t.Errorf("expected group node, got %v", nodes[0].Type)
				}
				children := nodes[0].Children
				if len(children) != 2 {
					t.Fatalf("expected 2 children in group, got %d", len(children))
				}
				if children[0].Part.Query != "field1:val1" {
					t.Errorf("expected 'field1:val1', got '%s'", children[0].Part.Query)
				}
				// {|} should be normalized to {OR}
				if children[0].Aggregator != "{OR}" {
					t.Errorf("expected normalized aggregator '{OR}', got '%s'", children[0].Aggregator)
				}
				if children[1].Part.Query != "field1:val2" {
					t.Errorf("expected 'field1:val2', got '%s'", children[1].Part.Query)
				}
			},
		},
		{
			name:          "group with AND between group and leaf",
			input:         "{P}field1:val1{|}field1:val2{/P}{&}field2:val3",
			wantNodeCount: 2, // one group + one leaf
			checkFunc: func(t *testing.T, nodes []queryNode) {
				if nodes[0].Type != queryNodeGroup {
					t.Errorf("expected group node, got %v", nodes[0].Type)
				}
				// group's aggregator should be {AND} (normalized from {&})
				if nodes[0].Aggregator != "{AND}" {
					t.Errorf("expected group aggregator '{AND}', got '%s'", nodes[0].Aggregator)
				}
				if nodes[1].Type != queryNodeLeaf {
					t.Errorf("expected leaf node, got %v", nodes[1].Type)
				}
				if nodes[1].Part.Query != "field2:val3" {
					t.Errorf("expected 'field2:val3', got '%s'", nodes[1].Part.Query)
				}
			},
		},
		{
			name:          "alias normalization - {&} to {AND}",
			input:         "field1:val1{&}field2:val2",
			wantNodeCount: 2,
			checkFunc: func(t *testing.T, nodes []queryNode) {
				if nodes[0].Aggregator != "{AND}" {
					t.Errorf("expected '{AND}' after normalization, got '%s'", nodes[0].Aggregator)
				}
			},
		},
		{
			name:          "alias normalization - {|} to {OR}",
			input:         "field1:val1{|}field2:val2",
			wantNodeCount: 2,
			checkFunc: func(t *testing.T, nodes []queryNode) {
				if nodes[0].Aggregator != "{OR}" {
					t.Errorf("expected '{OR}' after normalization, got '%s'", nodes[0].Aggregator)
				}
			},
		},
		{
			name:          "unmatched parens - extra open - flattens gracefully",
			input:         "{P}field1:val1{|}field1:val2",
			wantNodeCount: 2, // should flatten, no group
			checkFunc: func(t *testing.T, nodes []queryNode) {
				// All nodes should be leaves (flattened)
				for i, n := range nodes {
					if n.Type != queryNodeLeaf {
						t.Errorf("node %d: expected leaf (flattened), got group", i)
					}
				}
			},
		},
		{
			name:          "unmatched parens - extra close - flattens gracefully",
			input:         "field1:val1{|}field1:val2{/P}",
			wantNodeCount: 2,
			checkFunc: func(t *testing.T, nodes []queryNode) {
				for i, n := range nodes {
					if n.Type != queryNodeLeaf {
						t.Errorf("node %d: expected leaf (flattened), got group", i)
					}
				}
			},
		},
		{
			name:          "empty group is omitted",
			input:         "{P}{/P}{&}field1:val1",
			wantNodeCount: 2, // empty group node + leaf
			checkFunc: func(t *testing.T, nodes []queryNode) {
				// The empty group still produces a node, but buildGroupedWhereClause handles it
				foundLeaf := false
				for _, n := range nodes {
					if n.Type == queryNodeLeaf {
						foundLeaf = true
					}
				}
				if !foundLeaf {
					t.Error("expected at least one leaf node")
				}
			},
		},
		{
			name:          "mixed old and new syntax",
			input:         "field1:val1{AND}field2:val2",
			wantNodeCount: 2,
			checkFunc: func(t *testing.T, nodes []queryNode) {
				if nodes[0].Aggregator != "{AND}" {
					t.Errorf("expected '{AND}', got '%s'", nodes[0].Aggregator)
				}
			},
		},
		{
			name:          "canonical example from spec",
			input:         "{P}functional_area_01:Logística{|}functional_area_01:Putas{/P}{&}district:Beja",
			wantNodeCount: 2,
			checkFunc: func(t *testing.T, nodes []queryNode) {
				// First node: group with 2 OR'd children
				if nodes[0].Type != queryNodeGroup {
					t.Fatalf("expected group, got leaf")
				}
				if len(nodes[0].Children) != 2 {
					t.Fatalf("expected 2 children, got %d", len(nodes[0].Children))
				}
				if nodes[0].Children[0].Part.Query != "functional_area_01:Logística" {
					t.Errorf("child 0: expected 'functional_area_01:Logística', got '%s'", nodes[0].Children[0].Part.Query)
				}
				if nodes[0].Children[0].Aggregator != "{OR}" {
					t.Errorf("child 0 aggregator: expected '{OR}', got '%s'", nodes[0].Children[0].Aggregator)
				}
				if nodes[0].Children[1].Part.Query != "functional_area_01:Putas" {
					t.Errorf("child 1: expected 'functional_area_01:Putas', got '%s'", nodes[0].Children[1].Part.Query)
				}
				if nodes[0].Aggregator != "{AND}" {
					t.Errorf("group aggregator: expected '{AND}', got '%s'", nodes[0].Aggregator)
				}
				// Second node: leaf
				if nodes[1].Type != queryNodeLeaf {
					t.Fatalf("expected leaf, got group")
				}
				if nodes[1].Part.Query != "district:Beja" {
					t.Errorf("expected 'district:Beja', got '%s'", nodes[1].Part.Query)
				}
			},
		},
		{
			name:          "multiple groups",
			input:         "{P}f1:v1{|}f1:v2{/P}{&}{P}f2:v3{|}f2:v4{/P}",
			wantNodeCount: 2,
			checkFunc: func(t *testing.T, nodes []queryNode) {
				if nodes[0].Type != queryNodeGroup || nodes[1].Type != queryNodeGroup {
					t.Error("expected both nodes to be groups")
				}
				if len(nodes[0].Children) != 2 || len(nodes[1].Children) != 2 {
					t.Error("expected each group to have 2 children")
				}
				if nodes[0].Aggregator != "{AND}" {
					t.Errorf("group1 aggregator: expected '{AND}', got '%s'", nodes[0].Aggregator)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := parseQueryStringGrouped(tt.input)
			if len(nodes) != tt.wantNodeCount {
				t.Fatalf("expected %d nodes, got %d: %+v", tt.wantNodeCount, len(nodes), nodes)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, nodes)
			}
		})
	}
}

func TestBuildGroupedWhereClause(t *testing.T) {
	tests := []struct {
		name       string
		conditions []groupedCondition
		wantSQL    string
		wantArgs   int
	}{
		{
			name:       "empty conditions",
			conditions: nil,
			wantSQL:    "",
			wantArgs:   0,
		},
		{
			name: "single leaf",
			conditions: []groupedCondition{
				{
					Type: queryNodeLeaf,
					Condition: queryCondition{
						Condition: "field1 = ?",
						Value:     "val1",
					},
					Aggregator: QueryParamAggregatorNONE,
				},
			},
			wantSQL:  "field1 = ?",
			wantArgs: 1,
		},
		{
			name: "two leaves with AND",
			conditions: []groupedCondition{
				{
					Type: queryNodeLeaf,
					Condition: queryCondition{
						Condition: "field1 = ?",
						Value:     "val1",
					},
					Aggregator: QueryParamAggregatorAND,
				},
				{
					Type: queryNodeLeaf,
					Condition: queryCondition{
						Condition: "field2 = ?",
						Value:     "val2",
					},
					Aggregator: QueryParamAggregatorNONE,
				},
			},
			wantSQL:  "field1 = ? AND field2 = ?",
			wantArgs: 2,
		},
		{
			name: "single group with OR children",
			conditions: []groupedCondition{
				{
					Type: queryNodeGroup,
					Children: []groupedCondition{
						{
							Type: queryNodeLeaf,
							Condition: queryCondition{
								Condition: "field1 = ?",
								Value:     "val1",
							},
							Aggregator: QueryParamAggregatorOR,
						},
						{
							Type: queryNodeLeaf,
							Condition: queryCondition{
								Condition: "field1 = ?",
								Value:     "val2",
							},
							Aggregator: QueryParamAggregatorNONE,
						},
					},
					Aggregator: QueryParamAggregatorNONE,
				},
			},
			wantSQL:  "(field1 = ? OR field1 = ?)",
			wantArgs: 2,
		},
		{
			name: "group AND leaf - canonical example",
			conditions: []groupedCondition{
				{
					Type: queryNodeGroup,
					Children: []groupedCondition{
						{
							Type: queryNodeLeaf,
							Condition: queryCondition{
								Condition: "functional_area_01 = ?",
								Value:     "Logística",
							},
							Aggregator: QueryParamAggregatorOR,
						},
						{
							Type: queryNodeLeaf,
							Condition: queryCondition{
								Condition: "functional_area_01 = ?",
								Value:     "Putas",
							},
							Aggregator: QueryParamAggregatorNONE,
						},
					},
					Aggregator: QueryParamAggregatorAND,
				},
				{
					Type: queryNodeLeaf,
					Condition: queryCondition{
						Condition: "district = ?",
						Value:     "Beja",
					},
					Aggregator: QueryParamAggregatorNONE,
				},
			},
			wantSQL:  "(functional_area_01 = ? OR functional_area_01 = ?) AND district = ?",
			wantArgs: 3,
		},
		{
			name: "multiple groups with AND",
			conditions: []groupedCondition{
				{
					Type: queryNodeGroup,
					Children: []groupedCondition{
						{
							Type: queryNodeLeaf,
							Condition: queryCondition{
								Condition: "f1 = ?",
								Value:     "v1",
							},
							Aggregator: QueryParamAggregatorOR,
						},
						{
							Type: queryNodeLeaf,
							Condition: queryCondition{
								Condition: "f1 = ?",
								Value:     "v2",
							},
							Aggregator: QueryParamAggregatorNONE,
						},
					},
					Aggregator: QueryParamAggregatorAND,
				},
				{
					Type: queryNodeGroup,
					Children: []groupedCondition{
						{
							Type: queryNodeLeaf,
							Condition: queryCondition{
								Condition: "f2 = ?",
								Value:     "v3",
							},
							Aggregator: QueryParamAggregatorOR,
						},
						{
							Type: queryNodeLeaf,
							Condition: queryCondition{
								Condition: "f2 = ?",
								Value:     "v4",
							},
							Aggregator: QueryParamAggregatorNONE,
						},
					},
					Aggregator: QueryParamAggregatorNONE,
				},
			},
			wantSQL:  "(f1 = ? OR f1 = ?) AND (f2 = ? OR f2 = ?)",
			wantArgs: 4,
		},
		{
			name: "empty group is omitted",
			conditions: []groupedCondition{
				{
					Type:       queryNodeGroup,
					Children:   nil,
					Aggregator: QueryParamAggregatorAND,
				},
				{
					Type: queryNodeLeaf,
					Condition: queryCondition{
						Condition: "field1 = ?",
						Value:     "val1",
					},
					Aggregator: QueryParamAggregatorNONE,
				},
			},
			wantSQL:  " AND field1 = ?",
			wantArgs: 1,
		},
		{
			name: "IN clause inside group",
			conditions: []groupedCondition{
				{
					Type: queryNodeGroup,
					Children: []groupedCondition{
						{
							Type: queryNodeLeaf,
							Condition: queryCondition{
								Condition: "field1 IN (?, ?)",
								Value:     []interface{}{"v1", "v2"},
							},
							Aggregator: QueryParamAggregatorNONE,
						},
					},
					Aggregator: QueryParamAggregatorAND,
				},
				{
					Type: queryNodeLeaf,
					Condition: queryCondition{
						Condition: "field2 = ?",
						Value:     "v3",
					},
					Aggregator: QueryParamAggregatorNONE,
				},
			},
			wantSQL:  "(field1 IN (?, ?)) AND field2 = ?",
			wantArgs: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := buildGroupedWhereClause(tt.conditions)
			if sql != tt.wantSQL {
				t.Errorf("SQL mismatch:\n  want: %q\n  got:  %q", tt.wantSQL, sql)
			}
			if len(args) != tt.wantArgs {
				t.Errorf("args count: want %d, got %d", tt.wantArgs, len(args))
			}
		})
	}
}

func TestNormalizeAggregator(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"{&}", "{AND}"},
		{"{|}", "{OR}"},
		{"{AND}", "{AND}"},
		{"{OR}", "{OR}"},
		{"{-}", "{-}"},
		{"", ""},
		{"{unknown}", "{unknown}"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeAggregator(tt.input)
			if got != tt.want {
				t.Errorf("normalizeAggregator(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHasGroupTokens(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"{P}field:val{/P}", true},
		{"field:val{AND}field2:val2", false},
		{"{P}field:val{|}field:val2{/P}{&}field3:val3", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := hasGroupTokens(tt.input); got != tt.want {
				t.Errorf("hasGroupTokens(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseQueryStringGroupedDeterminism(t *testing.T) {
	// Stress test: parse the same grouped query 100 times and verify consistent output
	input := "{P}functional_area_01:Logística{|}functional_area_01:Putas{/P}{&}district:Beja"

	var referenceNodes []queryNode
	for i := 0; i < 100; i++ {
		nodes := parseQueryStringGrouped(input)
		if i == 0 {
			referenceNodes = nodes
			continue
		}

		if len(nodes) != len(referenceNodes) {
			t.Fatalf("iteration %d: node count changed: %d vs %d", i, len(nodes), len(referenceNodes))
		}

		for j, node := range nodes {
			ref := referenceNodes[j]
			if node.Type != ref.Type {
				t.Fatalf("iteration %d, node %d: type changed", i, j)
			}
			if node.Aggregator != ref.Aggregator {
				t.Fatalf("iteration %d, node %d: aggregator changed: %q vs %q", i, j, node.Aggregator, ref.Aggregator)
			}
			if node.Type == queryNodeLeaf {
				if node.Part.Query != ref.Part.Query {
					t.Fatalf("iteration %d, node %d: query changed: %q vs %q", i, j, node.Part.Query, ref.Part.Query)
				}
			}
			if node.Type == queryNodeGroup {
				if len(node.Children) != len(ref.Children) {
					t.Fatalf("iteration %d, node %d: children count changed", i, j)
				}
			}
		}
	}
}
