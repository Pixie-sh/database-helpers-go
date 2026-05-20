package elastic

import (
	"encoding/json"
	"fmt"
	"strings"

	base "github.com/pixie-sh/database-helpers-go/pipeline/operators"
	"github.com/pixie-sh/errors-go"
)

type Builder struct {
	index  string
	body   map[string]interface{}
	script *ScriptBuilder
}

func NewBuilder(index string) *Builder {
	return &Builder{
		index: index,
		body:  make(map[string]interface{}),
	}
}

func NewResult(index string) Result {
	return base.NewResult(NewBuilder(index))
}

func (b *Builder) Index() string {
	return b.index
}

func (b *Builder) Body() map[string]interface{} {
	return b.body
}

func (b *Builder) SearchRequest() SearchRequest {
	return SearchRequest{
		Method: "POST",
		Path:   "/" + strings.Trim(b.index, "/") + "/_search",
		Index:  b.index,
		Body:   b.body,
	}
}

func (b *Builder) SetSource(fields ...string) *Builder {
	if len(fields) == 0 {
		delete(b.body, "_source")
		return b
	}

	b.body["_source"] = fields
	return b
}

func (b *Builder) SetSize(size int) *Builder {
	if size > 0 {
		b.body["size"] = size
	}
	return b
}

func (b *Builder) SetFrom(from int) *Builder {
	if from >= 0 {
		b.body["from"] = from
	}
	return b
}

func (b *Builder) SetTrackTotalHits(value interface{}) *Builder {
	b.body["track_total_hits"] = value
	return b
}

func (b *Builder) SetSort(sort ...Query) *Builder {
	items := make([]interface{}, 0, len(sort))
	for _, item := range sort {
		items = append(items, item)
	}
	b.body["sort"] = items
	return b
}

func (b *Builder) SetSearchAfter(values ...interface{}) *Builder {
	if len(values) == 0 {
		delete(b.body, "search_after")
		return b
	}

	b.body["search_after"] = values
	return b
}

func (b *Builder) SetPointInTime(id string, keepAlive string) *Builder {
	if id == "" {
		delete(b.body, "pit")
		return b
	}

	pit := Query{"id": id}
	if keepAlive != "" {
		pit["keep_alive"] = keepAlive
	}
	b.body["pit"] = pit
	return b
}

func (b *Builder) SetQuery(query Query) *Builder {
	b.body["query"] = query
	return b
}

func (b *Builder) Query() Query {
	query, ok := b.body["query"].(Query)
	if ok {
		return query
	}

	queryMap, ok := b.body["query"].(map[string]interface{})
	if ok {
		return Query(queryMap)
	}

	query = Bool()
	b.body["query"] = query
	return query
}

func (b *Builder) BoolQuery() Query {
	query := b.Query()
	boolQuery, ok := query["bool"].(Query)
	if ok {
		return boolQuery
	}

	boolMap, ok := query["bool"].(map[string]interface{})
	if ok {
		return Query(boolMap)
	}

	boolQuery = Query{}
	b.body["query"] = Query{"bool": boolQuery}
	return boolQuery
}

func (b *Builder) AddFilter(query Query) *Builder {
	appendBoolClause(b.BoolQuery(), "filter", query)
	return b
}

func (b *Builder) AddMust(query Query) *Builder {
	appendBoolClause(b.BoolQuery(), "must", query)
	return b
}

func (b *Builder) AddMustNot(query Query) *Builder {
	appendBoolClause(b.BoolQuery(), "must_not", query)
	return b
}

func (b *Builder) AddShould(query Query) *Builder {
	appendBoolClause(b.BoolQuery(), "should", query)
	return b
}

func (b *Builder) SetMinimumShouldMatch(value int) *Builder {
	if value > 0 {
		b.BoolQuery()["minimum_should_match"] = value
	}
	return b
}

func (b *Builder) SetScriptBuilder(script *ScriptBuilder) *Builder {
	b.script = script
	return b
}

func (b *Builder) ScriptBuilder() *ScriptBuilder {
	if b.script == nil {
		b.script = NewScriptBuilder()
	}
	return b.script
}

func (b *Builder) ApplyScriptScore(query Query, script *ScriptBuilder) error {
	if script == nil {
		script = b.script
	}
	if script == nil {
		return errors.New("script builder is nil")
	}

	source := script.Source()
	if strings.TrimSpace(source) == "" {
		return errors.New("script source is empty")
	}

	if query == nil {
		query = b.Query()
	}

	b.body["query"] = Query{
		"script_score": Query{
			"query": query,
			"script": Query{
				"source": source,
				"params": script.Params(),
			},
		},
	}
	return nil
}

func (b *Builder) JSON(pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(b.body, "", "  ")
	}

	return json.Marshal(b.body)
}

func appendBoolClause(boolQuery Query, key string, query Query) {
	if query == nil {
		return
	}

	items, ok := boolQuery[key].([]interface{})
	if !ok {
		items = []interface{}{}
	}

	boolQuery[key] = append(items, query)
}

func (b *Builder) DebugString(pretty bool) (string, error) {
	body, err := b.JSON(pretty)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("POST /%s/_search\n%s", strings.Trim(b.index, "/"), string(body)), nil
}
