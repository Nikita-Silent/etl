package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// parseFloatWithComma parses a float string that may use comma as decimal separator.
func parseFloatWithComma(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("invalid float")
	}

	commaIndex := strings.Index(s, ",")
	if commaIndex == -1 {
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}
		return val, nil
	}
	if strings.Count(s, ",") > 1 {
		return 0, fmt.Errorf("invalid float")
	}

	afterComma := s[commaIndex+1:]
	if !strings.Contains(s, ".") && len(afterComma) == 3 {
		allDigits := true
		for i := 0; i < len(afterComma); i++ {
			if afterComma[i] < '0' || afterComma[i] > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			return 0, fmt.Errorf("invalid float")
		}
	}

	s = strings.Replace(s, ",", ".", 1)
	if strings.Count(s, ".") > 1 {
		parts := strings.SplitN(s, ".", 3)
		if len(parts) >= 2 {
			s = parts[0] + "." + parts[1]
		}
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}
