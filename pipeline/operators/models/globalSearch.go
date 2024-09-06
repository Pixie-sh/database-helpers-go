package models

type SearchableProperty struct {
	Field      string
	Type       string // "text", "int", "date", "bool", "uuid"
	Comparison string // "like", "=", ">", "<", ">=", "<="
	Format     string // For dates: "2006-01-02", for others: leave empty
	LikeBefore bool   // Use % before the given text
	LikeAfter  bool   // Use % after the given text
	Unaccent   bool   // Use unaccent on query to avoid issues with ç/c ã/a; if false raw value is used
	Ilike      bool   //use ILIKE comparison, LikeBefore and LikeAfter is used with the same purpose; if false LIKE is used
}
