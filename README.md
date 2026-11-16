# database-helpers-go

Query pipelines for Go services. Compose a complex relational or Elasticsearch query from small, reusable operators, and get filtering, search, ordering and pagination driven by the request's query string.

[![Go Reference](https://pkg.go.dev/badge/github.com/pixie-sh/database-helpers-go.svg)](https://pkg.go.dev/github.com/pixie-sh/database-helpers-go)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Made by [Pixie](https://pixie.sh). Project page: <https://pixie-sh.github.io/database-helpers-go/>

## Install

```sh
go get github.com/pixie-sh/database-helpers-go
```

## Why a pipeline?

A list endpoint tends to grow the same way every time: a tenant filter, then free-text search, then sorting, then pagination, each with its own `if` blocks around the query builder. Here each of those is an **operator**, and a query is the list of operators you add, in order.

- **One operator, one concern.** Operators are small and testable on their own.
- **Driven by the request.** Search, order and pagination operators read `?q=`, `?sort_by=`, `?page=` and friends from the query string and validate them against the fields you allow.
- **Two back ends, one shape.** The same pipeline builds a GORM query or an Elasticsearch `_search` body.
- **Your own operators.** Business rules such as "freshness score" become operators you reuse across endpoints.

## Quick start: a relational list endpoint

```go
import (
	"github.com/pixie-sh/database-helpers-go/database"
	"github.com/pixie-sh/database-helpers-go/pipeline"
	"github.com/pixie-sh/database-helpers-go/pipeline/operators"
	"github.com/pixie-sh/database-helpers-go/pipeline/operators/models"
	"github.com/pixie-sh/logger-go/logger"
)

// GET /users?q=ann&sort_by=-created_at&page=2&per_page=25
func listUsers(ctx context.Context, db *database.DB, tenantID uint64, queryParams map[string][]string) ([]User, error) {
	var users []User

	_, err := pipeline.NewPipeline(logger.Logger).
		AddOperator(
			// always applied, never controlled by the caller
			operators.NewHardWhereOperator("tenant_id = ?", tenantID),
			// ?q= searches the listed fields
			operators.NewGlobalSearchOperator(queryParams, "q",
				models.SearchableProperty{Field: "name", Type: "text", Comparison: "LIKE", LikeBefore: true, LikeAfter: true},
			),
			// ordering, limited to the fields you allow, with a default
			operators.NewOrderByOperator(queryParams, true, []string{"-created_at"},
				models.SearchableProperty{Field: "created_at", Type: "date"},
			),
			// pagination; runs the query and fills users
			operators.NewPaginateOperator(queryParams, &users, 10, 25, 50),
		).
		ExecWithPassable(ctx, operators.NewResult(db.Model(&User{})))

	return users, err
}
```

## Quick start: an Elasticsearch search

The same pipeline builds a search body. This one filters live deals that are either online or within 250 km, excludes one ID, and ranks by a text match.

```go
import "github.com/pixie-sh/database-helpers-go/pipeline/operators/elastic"

result, err := pipeline.NewPipeline(logger.Logger).
	AddOperator(
		elastic.NewSourceOperator("id", "title", "status"),
		elastic.NewFilterOperator(elastic.Term("status", "live")),
		elastic.NewFilterOperator(elastic.BoolShould(1,
			elastic.Term("deal_type", "online"),
			elastic.BoolFilter(
				elastic.Term("deal_type", "physical"),
				elastic.GeoDistance("locations.coords", "250km", 52.3676, 4.9041),
			),
		)),
		elastic.NewMustNotOperator(elastic.IDs("hidden_deal_id")),
		elastic.NewShouldOperator(elastic.MultiMatch("eco laundry", "title^3", "description")),
		elastic.NewMinimumShouldMatchOperator(1),
		elastic.NewFromSizePaginationOperator(0, 30),
	).
	ExecWithPassable(ctx, operators.NewResult(elastic.NewBuilder("deals")))
if err != nil {
	return err
}

request := result.GetPassable().(*elastic.Builder).SearchRequest()
```

`request` is a `POST /deals/_search` with this body:

```json
{
  "_source": ["id", "title", "status"],
  "from": 0,
  "size": 30,
  "query": {
    "bool": {
      "filter": [
        { "term": { "status": "live" } },
        { "bool": { "minimum_should_match": 1, "should": [
          { "term": { "deal_type": "online" } },
          { "bool": { "filter": [
            { "term": { "deal_type": "physical" } },
            { "geo_distance": { "distance": "250km", "locations.coords": { "lat": 52.3676, "lon": 4.9041 } } }
          ] } }
        ] } }
      ],
      "must_not": [ { "ids": { "values": ["hidden_deal_id"] } } ],
      "should": [ { "multi_match": { "fields": ["title^3", "description"], "query": "eco laundry" } } ],
      "minimum_should_match": 1
    }
  }
}
```

Send it with the bundled repository, or with any client you prefer:

```go
repo := database.NewElasticRepository(database.ElasticConfiguration{
	BaseURL: "https://search.internal:9200",
	Index:   "deals",
})
response, err := repo.Search(ctx, request)
```

## Repositories and transactions

`database.Repository[T]` gives a repository transaction helpers without hiding GORM.

```go
type UserRepository struct {
	database.Repository[UserRepository]
}

func NewUserRepository(db *database.DB) UserRepository {
	return UserRepository{database.NewRepository(db, NewUserRepository)}
}

func (r UserRepository) Rename(id uint64, name string) error {
	return r.Transaction(func(tx *database.DB) error {
		return tx.Model(&User{}).Where("id = ?", id).Update("name", name).Error
	})
}
```

`WithTx(tx)` returns a copy of the repository bound to an open transaction, so several repositories can share one.

`database.ElasticRepository` covers the Elasticsearch side: `Search`, `EnsureIndex`, `BulkIndex`, a raw `Do`, and helpers for cursors (`EncodeSearchAfterCursor`, `DecodeSearchAfterCursor`, `TrimCursorHits`) and totals.

## Overview

The module has two areas:

- `database`: thin wrappers around GORM, GORM migrations, connection factories, repository transaction helpers and the Elasticsearch repository.
- `pipeline`: the operator pipeline used to compose query builders in a predictable order.

The pipeline supports two operator families:

- GORM operators, which mutate a `*gorm.DB` query.
- Elasticsearch operators, which build a dependency-free Elasticsearch/OpenSearch search request body.

## Database Helpers

The `database` package exposes aliases and helpers around GORM:

```go
db, err := database.NewGormDb(ctx, &database.GormDbConfiguration{
    Driver: database.PsqlDriver,
    Dsn:    "postgres://user:pass@localhost:5432/app",
})
```

Supported database drivers:

- `database.MysqlDriver`
- `database.PsqlDriver`

The package also includes:

- `database.FactoryInstance` for config-driven ORM creation.
- `database.Repository[T]` for transaction helpers.
- `database.NewMigrator` for `gormigrate` migrations.

## Pipeline

The pipeline executes compatible operators in the order they are added.

```go
result, err := pipeline.NewPipeline(log).
    AddOperator(op1, op2, op3).
    ExecWithPassable(ctx, operators.NewResult(passable))
```

Each operator receives a `Result`, reads the current passable value, mutates or replaces it, and returns the result.

Operator families cannot be mixed in the same pipeline. GORM operators use `DatabaseOperatorType`; Elasticsearch operators use `ElasticOperatorType`.

## GORM Query Operators

GORM operators live in `pipeline/operators` and expect the passable value to be a `*database.DB`, which is an alias for `*gorm.DB`.

Common operators include:

- `NewHardWhereOperator`
- `NewWhereIdsInOperator`
- `NewWhereUUIDsInOperator`
- `NewWherePropertiesInOperator`
- `NewSearchInPropertiesOperator`
- `NewGlobalSearchOperator`
- `NewOrderByOperator`
- `NewPaginateOperator`
- `NewOffsetPaginateOperator`
- `NewListOperator`

Example:

```go
var users []User

result, err := pipeline.NewPipeline(log).
    AddOperator(
        operators.NewHardWhereOperator("tenant_id = ?", tenantID),
        operators.NewGlobalSearchOperator(queryParams, "q",
            models.SearchableProperty{
                Field:      "name",
                Type:       "text",
                Comparison: "LIKE",
                LikeBefore: true,
                LikeAfter:  true,
            },
        ),
        operators.NewOrderByOperator(queryParams, true, []string{"-created_at"},
            models.SearchableProperty{Field: "created_at", Type: "date"},
        ),
        operators.NewPaginateOperator(queryParams, &users, 10, 25, 50),
    ).
    ExecWithPassable(ctx, operators.NewResult(db.Model(&User{})))
```

## Elasticsearch Operators

Elasticsearch operators live in `pipeline/operators/elastic` and build an Elasticsearch/OpenSearch `_search` request body.

The operators only build the request and do not depend on a specific Elasticsearch client. Execute it with `database.ElasticRepository`, the official Elasticsearch client, an OpenSearch client, or a direct HTTP call.

The passable value must be an `*elastic.Builder`:

```go
result, err := pipeline.NewPipeline(log).
    AddOperator(
        elastic.NewSourceOperator("id", "title", "status"),
        elastic.NewFromSizePaginationOperator(0, 30),
        elastic.NewFilterOperator(elastic.Term("status", "live")),
        elastic.NewMustNotOperator(elastic.IDs("hidden_deal_id")),
        elastic.NewShouldOperator(elastic.MultiMatch(
            "laundry cleaning household eco product ugc",
            "title^3",
            "description",
            "hash_tags^2",
        )),
        elastic.NewMinimumShouldMatchOperator(1),
    ).
    ExecWithPassable(ctx, operators.NewResult(elastic.NewBuilder("deals")))

builder := result.GetPassable().(*elastic.Builder)
request := builder.SearchRequest()
```

`request.Body` can then be sent to your Elasticsearch client as the JSON search body.

### Query Helpers

The elastic package includes small helpers for common query clauses:

- `elastic.MatchAll()`
- `elastic.BoolFilter(...)`
- `elastic.BoolMust(...)`
- `elastic.BoolMustNot(...)`
- `elastic.BoolShould(minimumShouldMatch, ...)`
- `elastic.Term(field, value)`
- `elastic.Terms(field, values)`
- `elastic.IDs(values...)`
- `elastic.MultiMatch(query, fields...)`
- `elastic.GeoDistance(field, distance, lat, lon)`
- `elastic.SortField(field, order)`

Example nested filter:

```go
elastic.NewFilterOperator(elastic.BoolShould(1,
    elastic.Term("deal_type", "online"),
    elastic.BoolFilter(
        elastic.Term("deal_type", "physical"),
        elastic.GeoDistance("locations.coords", "250km", 52.3676, 4.9041),
    ),
))
```

### Script Score Composition

Large Painless scripts can be built from smaller fragments.

Scripts have four ordered sections:

- `ScriptSectionHelpers`
- `ScriptSectionSetup`
- `ScriptSectionScoring`
- `ScriptSectionReturn`

`ScriptBuilder` defaults to:

```painless
double score = 0.0;

return Math.max(score, 0.001);
```

Example:

```go
result, err := pipeline.NewPipeline(log).
    AddOperator(
        elastic.NewFilterOperator(elastic.Term("status", "live")),
        elastic.NewScriptHelperOperator(`
double getVal(String field) {
  if (doc.containsKey(field) && !doc[field].empty) {
    return doc[field].value;
  }
  return 0.0;
}
`, nil),
        elastic.NewScriptScoringOperator(`
if (params.w_freshness != 0 && doc.containsKey('created_at') && !doc['created_at'].empty) {
  long ageMillis = params.now - doc['created_at'].value.toInstant().toEpochMilli();
  double ageDays = ageMillis / 86400000.0;
  double freshness = Math.max(0, 1.0 - (ageDays / params.freshness_window_days));
  score += freshness * params.w_freshness;
}
`, map[string]interface{}{
            "now":                   nowMillis,
            "w_freshness":           25.0,
            "freshness_window_days": 14,
        }),
        elastic.NewScriptScoreOperator(nil),
    ).
    ExecWithPassable(ctx, operators.NewResult(elastic.NewBuilder("deals")))
```

Script params are merged into the final `script.params` object. Duplicate param keys return an error so accidental weight collisions are caught early.

Project-specific packages should usually define business operators that wrap these fragments, for example `NewFreshnessScoreOperator`, `NewFollowerMatchScoreOperator`, or `NewGeoDistanceScoreOperator`.

### Pagination

Elasticsearch supports multiple pagination styles.

`from` plus `size` is the simplest offset-style pagination:

```go
elastic.NewFromSizePaginationOperator(60, 30)
```

`NewOffsetPaginationOperator` reads `page` and `per_page` from query params:

```go
elastic.NewOffsetPaginationOperator(queryParams, 30, 100)
```

This is simple but not recommended for deep pagination because Elasticsearch has `index.max_result_window`, commonly `10000`.

`search_after` is better for feed-style or deep pagination:

```go
elastic.NewSearchAfterPaginationOperator(
    31,
    []elastic.Query{
        elastic.SortField("_score", "desc"),
        elastic.SortField("id.keyword", "asc"),
    },
    lastScore,
    lastID,
)
```

Use `size + 1` when you want to determine `has_more`.

Point-in-time pagination can be added with:

```go
elastic.NewPointInTimeOperator(pitID, "1m")
```

`search_after` alone has no timeout because it is stateless. PIT has a `keep_alive` timeout and should be refreshed by passing `keep_alive` on each search request.

### Debugging

`DebugOperator` logs the final search request body. Put it near the end of the pipeline.

```go
elastic.NewDebugOperator(
    log,
    elastic.DebugMaxBodyBytes(20_000),
    elastic.DebugRedactParams("seen_ids"),
)
```

It logs a request-like string:

```text
POST /deals/_search
{
  "size": 30,
  "query": {}
}
```

The debug operator only logs the request being built. It does not execute Elasticsearch calls or log responses.

## License

[MIT](LICENSE)
