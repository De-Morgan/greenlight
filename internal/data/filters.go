package data

import (
	"math"
	"slices"
	"strings"
)

type Filters struct {
	Page         int      `json:"page"`
	Limit        int      `json:"limit"`
	Sort         string   `json:"sort"`
	SearchTerm   string   `json:"search_term"`
	SortSafelist []string `json:"-"`
}

func (f Filters) SortPresent() bool {
	if f.Sort == "" {
		return false
	}
	return slices.ContainsFunc(f.SortSafelist, func(e string) bool {
		return e == strings.TrimPrefix(f.Sort, "-")
	})
}
func (f Filters) SortColumn() string {
	return strings.TrimPrefix(f.Sort, "-")
}

func (f Filters) SortDirection() string {
	switch strings.HasPrefix(f.Sort, "-") {
	case true:
		return "DESC"
	default:
		return "ASC"
	}
}

func (f Filters) OffSet() int {
	if f.Page < 1 {
		return 0
	}
	return (f.Page - 1) * f.Limit
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	Limit        int `json:"limit,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_record,omitempty"`
}

// The calculateMetadata() function calculates the appropriate pagination metadata
// values given the total number of records, current page, and page size values. Note
// that the last page value is calculated using the math.Ceil() function, which rounds
// up a float to the nearest integer. So, for example, if there were 12 records in total
// and a page size of 5, the last page value would be math.Ceil(12/5) = 3.
func calculateMetadata(totalRecords, page, limit int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  page,
		Limit:        limit,
		TotalRecords: totalRecords,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(limit))),
	}
}
