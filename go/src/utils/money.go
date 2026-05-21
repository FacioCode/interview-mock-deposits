package utils

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"math"
	"strconv"
	"strings"
)

type Money struct {
	Value float64
}

func (t Money) CompareTo(other Money) int {
	roundedT := math.Round(t.Value * 100)
	roundedOther := math.Round(other.Value * 100)
	return int(roundedT) - int(roundedOther)
}

func (t Money) ToString() string {
	return fmt.Sprintf("%.2f", t.Value)
}
func (t *Money) UnmarshalJSON(b []byte) (err error) {
	var money float64
	value := string(b)
	value = strings.ReplaceAll(value, "\"", "")
	if value == "null" {
		t.Value = 0.0
	} else {
		money, err = strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		t.Value = money
	}
	return
}

func (t Money) MarshalJSON() ([]byte, error) {
	value := t.ToString()
	value = fmt.Sprintf("\"%s\"", value)

	return []byte(value), nil
}

func (t Money) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	value := t.ToString()
	av.N = &value
	return nil
}

func (t *Money) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	if av.N == nil {
		t.Value = 0.0
	} else {
		floatResult, _ := strconv.ParseFloat(*av.N, 64)
		t.Value = floatResult
	}
	return nil
}
