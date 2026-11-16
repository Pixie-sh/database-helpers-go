package elastic

import "github.com/pixie-sh/database-helpers-go/database"

// These aliases re-export the canonical Elasticsearch response shapes that
// live in the database package next to the generic ElasticRepository. Keeping
// the alias here means callers can read and write a single import
// (pipeline/operators/elastic) for builders, request types and response types.

type SearchResponse = database.ElasticSearchResponse
type SearchHits = database.ElasticSearchHits
type SearchHit = database.ElasticSearchHit
type BulkResponse = database.ElasticBulkResponse
type BulkItem = database.ElasticBulkItem
type BulkItemResult = database.ElasticBulkItemResult
type ErrorResponse = database.ElasticErrorResponse
type ErrorBody = database.ElasticErrorBody
