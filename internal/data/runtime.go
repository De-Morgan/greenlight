package data

import (
	"fmt"
	"strconv"
)

// Movie Runtime type
type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonVal := strconv.Quote(fmt.Sprintf("%d mins", r))
	return []byte(jsonVal), nil
}
