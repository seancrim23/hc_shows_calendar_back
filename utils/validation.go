package utils

var ValidShowQueryFilters = map[string]bool{
	"state":      true,
	"city":       true,
	"promoter":   true,
	"date_range": true,
}

// TODO maybe expand this if needed in future
var DateRangeMapping = map[string]int{
	"week":  7,
	"month": 30,
}
