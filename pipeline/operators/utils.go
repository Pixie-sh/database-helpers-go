package operators

import "strings"

const RemoveAccentFunction = `
CREATE OR REPLACE FUNCTION remove_accent(text) RETURNS text AS $$
SELECT translate($1, 
    'áàâãäåāăąÁÀÂÃÄÅĀĂĄéèêëēĕėęěÉÈÊËĒĔĖĘĚíìîïīĭįİÍÌÎÏĪĬĮİóòôõöōŏőÓÒÔÕÖŌŎŐúùûüūŭůűųÚÙÛÜŪŬŮŰŲ', 
    'aaaaaaaaaAAAAAAAAeeeeeeeeeeEEEEEEEEEiiiiiiiIIIIIIIIoooooooooOOOOOOOOOuuuuuuuuuUUUUUUUUU'
);
$$ LANGUAGE SQL IMMUTABLE STRICT;`

type QueryParamAggregatorEnum string

func (e QueryParamAggregatorEnum) String() string {
	return string(e)
}

const QueryParamAggregatorOR QueryParamAggregatorEnum = "{OR}"
const QueryParamAggregatorAND QueryParamAggregatorEnum = "{AND}"
const QueryParamAggregatorNONE QueryParamAggregatorEnum = "{-}"

// Grouping tokens
const QueryParamGroupOpen = "{P}"
const QueryParamGroupClose = "{/P}"

// Shorthand aliases (normalized at parse time)
const QueryParamAggregatorANDAlias = "{&}"
const QueryParamAggregatorORAlias = "{|}"

type AggregatorOperatorEnum string

func (enum AggregatorOperatorEnum) String() string {
	return string(enum)
}

const AggregatorConditionOR AggregatorOperatorEnum = " OR "
const AggregatorConditionAND AggregatorOperatorEnum = " AND "

type queryPart struct {
	Query      string
	Value      string
	Aggregator string
}

type queryCondition struct {
	Condition  string
	Value      interface{}
	Aggregator QueryParamAggregatorEnum
}

// AST node types for grouped query parsing
type queryNodeType int

const (
	queryNodeLeaf queryNodeType = iota
	queryNodeGroup
)

type queryNode struct {
	Type       queryNodeType
	Part       queryPart   // used when Type == queryNodeLeaf
	Children   []queryNode // used when Type == queryNodeGroup
	Aggregator string      // aggregator AFTER this node (e.g., "{AND}", "{OR}")
}

// groupedCondition mirrors queryNode but holds built SQL conditions instead of raw parts
type groupedCondition struct {
	Type       queryNodeType
	Condition  queryCondition     // for leaves
	Children   []groupedCondition // for groups
	Aggregator QueryParamAggregatorEnum
}

func parseQueryString(input string) []queryPart {
	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == '{' || r == '}'
	})

	var result []queryPart

	for i, part := range parts {
		if strings.Contains(part, ":") {
			query := strings.TrimSpace(part)
			var aggregator string
			if i+1 < len(parts) {
				aggregator = "{" + parts[i+1] + "}"
			}

			result = append(result, queryPart{Query: query, Aggregator: aggregator})
		}
	}

	return result
}

func buildComplexWhereClause(conditions []queryCondition) (string, []interface{}) {
	if len(conditions) == 0 {
		return "", nil
	}

	var fullCondition strings.Builder
	var args []interface{}

	fullCondition.WriteString("(")

	nextAggregator := QueryParamAggregatorAND
	for i, cond := range conditions {
		if i == 0 {
			nextAggregator = cond.Aggregator
		} else {
			switch nextAggregator {
			case QueryParamAggregatorOR:
				fullCondition.WriteString(AggregatorConditionOR.String())
				nextAggregator = cond.Aggregator
				break
			case QueryParamAggregatorAND:
				fullCondition.WriteString(AggregatorConditionAND.String())
				nextAggregator = cond.Aggregator
				break
			case QueryParamAggregatorNONE:
			default:
			}
		}

		fullCondition.WriteString(cond.Condition)

		// Handle both single values and slices (for IN clauses)
		if valueSlice, ok := cond.Value.([]interface{}); ok {
			args = append(args, valueSlice...)
		} else {
			args = append(args, cond.Value)
		}
	}

	fullCondition.WriteString(")")

	return fullCondition.String(), args
}

// normalizeAggregator converts shorthand aliases to canonical aggregator tokens.
func normalizeAggregator(token string) string {
	switch token {
	case QueryParamAggregatorANDAlias:
		return QueryParamAggregatorAND.String()
	case QueryParamAggregatorORAlias:
		return QueryParamAggregatorOR.String()
	default:
		return token
	}
}

// hasGroupTokens checks if the input contains any group delimiters.
func hasGroupTokens(input string) bool {
	return strings.Contains(input, QueryParamGroupOpen)
}

// parseQueryStringGrouped parses a query string into an AST of queryNode.
// When no group tokens ({P}/{/P}) are present, it returns a flat list of leaf nodes
// (functionally identical to parseQueryString). When groups are present, it builds
// a tree with group nodes containing children.
// Unmatched parentheses are handled gracefully by flattening (ignoring group tokens).
func parseQueryStringGrouped(input string) []queryNode {
	// Guard: check for unmatched parentheses
	if strings.Count(input, QueryParamGroupOpen) != strings.Count(input, QueryParamGroupClose) {
		// Fallback: parse as flat (ignore group tokens)
		return flatParseToNodes(input)
	}

	if !hasGroupTokens(input) {
		return flatParseToNodes(input)
	}

	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == '{' || r == '}'
	})

	// Stack-based parsing: root is an implicit group
	type stackFrame struct {
		children []queryNode
	}
	stack := []stackFrame{{}} // start with root frame

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		switch part {
		case "P":
			// Push new group
			stack = append(stack, stackFrame{})

		case "/P":
			if len(stack) < 2 {
				// Mismatched /P without P — shouldn't happen due to guard, but be safe
				continue
			}
			// Pop group and add to parent
			group := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			// Look ahead for aggregator after {/P}
			var aggregator string
			if i+1 < len(parts) && !strings.Contains(parts[i+1], ":") && parts[i+1] != "P" && parts[i+1] != "/P" {
				aggregator = normalizeAggregator("{" + parts[i+1] + "}")
				i++ // consume the aggregator token
			}

			node := queryNode{
				Type:       queryNodeGroup,
				Children:   group.children,
				Aggregator: aggregator,
			}
			stack[len(stack)-1].children = append(stack[len(stack)-1].children, node)

		default:
			if strings.Contains(part, ":") {
				// Query part — create leaf node
				query := strings.TrimSpace(part)

				// Look ahead for aggregator
				var aggregator string
				if i+1 < len(parts) {
					next := parts[i+1]
					if next != "P" && next != "/P" && !strings.Contains(next, ":") {
						aggregator = normalizeAggregator("{" + next + "}")
						i++ // consume the aggregator token
					}
				}

				node := queryNode{
					Type: queryNodeLeaf,
					Part: queryPart{
						Query:      query,
						Aggregator: aggregator,
					},
					Aggregator: aggregator,
				}
				stack[len(stack)-1].children = append(stack[len(stack)-1].children, node)
			}
			// Non-colon, non-group tokens that weren't consumed as aggregators are ignored
		}
	}

	// If stack has more than root, there were unclosed groups — flatten remaining
	for len(stack) > 1 {
		group := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		stack[len(stack)-1].children = append(stack[len(stack)-1].children, group.children...)
	}

	return stack[0].children
}

// flatParseToNodes converts a query string to a flat list of leaf queryNodes.
// Used as fallback when no group tokens are present.
func flatParseToNodes(input string) []queryNode {
	parts := parseQueryString(input)
	nodes := make([]queryNode, 0, len(parts))
	for _, p := range parts {
		p.Aggregator = normalizeAggregator(p.Aggregator)
		nodes = append(nodes, queryNode{
			Type:       queryNodeLeaf,
			Part:       p,
			Aggregator: p.Aggregator,
		})
	}
	return nodes
}

// buildGroupedWhereClause recursively builds a SQL WHERE clause from grouped conditions.
func buildGroupedWhereClause(conditions []groupedCondition) (string, []interface{}) {
	if len(conditions) == 0 {
		return "", nil
	}

	var sql strings.Builder
	var args []interface{}

	for i, cond := range conditions {
		if i > 0 {
			// Use the PREVIOUS condition's aggregator as the separator
			prevAgg := conditions[i-1].Aggregator
			switch prevAgg {
			case QueryParamAggregatorOR:
				sql.WriteString(AggregatorConditionOR.String())
			case QueryParamAggregatorAND:
				sql.WriteString(AggregatorConditionAND.String())
			default:
				// NONE or unknown — default to AND for safety
				sql.WriteString(AggregatorConditionAND.String())
			}
		}

		switch cond.Type {
		case queryNodeGroup:
			// Recurse into group, wrap in parentheses
			groupSQL, groupArgs := buildGroupedWhereClause(cond.Children)
			if groupSQL != "" {
				sql.WriteString("(")
				sql.WriteString(groupSQL)
				sql.WriteString(")")
				args = append(args, groupArgs...)
			}
		case queryNodeLeaf:
			sql.WriteString(cond.Condition.Condition)
			if valueSlice, ok := cond.Condition.Value.([]interface{}); ok {
				args = append(args, valueSlice...)
			} else {
				args = append(args, cond.Condition.Value)
			}
		}
	}

	return sql.String(), args
}
