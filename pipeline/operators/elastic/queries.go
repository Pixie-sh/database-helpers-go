package elastic

func MatchAll() Query {
	return Query{"match_all": Query{}}
}

func Bool() Query {
	return Query{"bool": Query{}}
}

func BoolFilter(queries ...Query) Query {
	boolQuery := Query{}
	for _, query := range queries {
		appendBoolClause(boolQuery, "filter", query)
	}
	return Query{"bool": boolQuery}
}

func BoolMust(queries ...Query) Query {
	boolQuery := Query{}
	for _, query := range queries {
		appendBoolClause(boolQuery, "must", query)
	}
	return Query{"bool": boolQuery}
}

func BoolMustNot(queries ...Query) Query {
	boolQuery := Query{}
	for _, query := range queries {
		appendBoolClause(boolQuery, "must_not", query)
	}
	return Query{"bool": boolQuery}
}

func BoolShould(minimumShouldMatch int, queries ...Query) Query {
	boolQuery := Query{}
	for _, query := range queries {
		appendBoolClause(boolQuery, "should", query)
	}
	if minimumShouldMatch > 0 {
		boolQuery["minimum_should_match"] = minimumShouldMatch
	}
	return Query{"bool": boolQuery}
}

func Term(field string, value interface{}) Query {
	return Query{"term": Query{field: value}}
}

func Terms(field string, values interface{}) Query {
	return Query{"terms": Query{field: values}}
}

func IDs(values ...string) Query {
	return Query{"ids": Query{"values": values}}
}

func MultiMatch(query string, fields ...string) Query {
	return Query{
		"multi_match": Query{
			"query":  query,
			"fields": fields,
		},
	}
}

func GeoDistance(field string, distance string, lat float64, lon float64) Query {
	return Query{
		"geo_distance": Query{
			"distance": distance,
			field: Query{
				"lat": lat,
				"lon": lon,
			},
		},
	}
}

func SortField(field string, order string) Query {
	if order == "" {
		order = "asc"
	}

	return Query{field: Query{"order": order}}
}

func ScriptScore(query Query, source string, params map[string]interface{}) Query {
	return Query{
		"script_score": Query{
			"query": query,
			"script": Query{
				"source": source,
				"params": params,
			},
		},
	}
}
