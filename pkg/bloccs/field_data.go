package bloccs

import (
	"fmt"
	"strings"
)

const Bedrock = 'B'

type FieldData []uint8

func (d FieldData) MarshalJSON() ([]byte, error) {
	var result string

	if d == nil {
		result = "null"
	} else {
		result = strings.Join(strings.Fields(fmt.Sprintf("%d", d)), ",")
	}

	return []byte(result), nil
}
