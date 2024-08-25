package models

type SearchableProperty struct {
	Field      string
	Type       string // "text", "int", "date", "bool", "uuid"
	Comparison string // "like", "=", ">", "<", ">=", "<="
	Format     string // For dates: "2006-01-02", for others: leave empty
	LikeBefore bool   // Use % before the given text
	LikeAfter  bool   // Use % after the given text
}
