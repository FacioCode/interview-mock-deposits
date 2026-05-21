package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type FormattedDate struct {
	time.Time
}

type DateTime struct {
	time.Time
}

const dateTimeLayout = time.RFC3339
const location = "America/Sao_Paulo"
const dateLayout = "2006-01-02"
const dateLayoutWithHours = "2006-01-02T15:04:05"

func (t *FormattedDate) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	if s == "null" {
		t = nil
		return
	}
	s = strings.ReplaceAll(s, "\"", "")
	parsedTime, parserError := time.Parse(dateLayout, s)
	if parserError != nil {
		err = parserError
		return
	}
	t.Time = parsedTime
	return
}

func (t FormattedDate) MarshalJSON() ([]byte, error) {
	result := t.ToDefaultFormat()
	return []byte(fmt.Sprintf("\"%s\"", result)), nil
}

func (t FormattedDate) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	formattedDate := t.ToDefaultFormat()
	av.S = aws.String(formattedDate)
	return nil
}

func (t *FormattedDate) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	date, err := time.Parse(dateLayout, *av.S)
	if err != nil {
		return err
	}
	t.Time = date
	return nil
}

func (t FormattedDate) ToDefaultFormat() string {
	return t.Format(dateLayout)
}

func (t *DateTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" {
		t = nil
		return nil
	}
	s = strings.ReplaceAll(s, "\"", "")
	parsedTime, err := time.Parse(dateTimeLayout, s)
	if err != nil {
		return err
	}
	t.Time = parsedTime
	return nil
}

func (t DateTime) ToString() string {
	if t.IsZero() {
		return ""
	}
	return t.Format(dateTimeLayout)
}

func (t DateTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	result := t.ToString()
	return []byte(fmt.Sprintf("\"%s\"", result)), nil
}

func (t *DateTime) Parse(date string) error {
	dateResult, err := time.Parse(dateTimeLayout, date)
	if err != nil {
		return err
	}
	t.Time = dateResult
	return nil
}

func (t DateTime) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	formattedDate := t.ToString()
	av.S = aws.String(formattedDate)
	return nil
}

func (t *DateTime) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	date, err := time.Parse(dateTimeLayout, *av.S)
	if err != nil {
		return err
	}
	t.Time = date
	return nil
}

func (t *DateTime) InAmericaSaoPaulo() DateTime {
	loadLocation, _ := time.LoadLocation("America/Sao_Paulo")
	return DateTime{Time: t.Time.In(loadLocation)}
}

func (t *DateTime) ToDefaultFormatWithHours() string {
	return t.Format(dateLayoutWithHours)
}

var GetNow = func() time.Time {
	loadLocation, _ := time.LoadLocation(location)
	return time.Now().In(loadLocation)
}

func GetFormattedNow() string {
	return GetNow().Format(dateTimeLayout)
}

var GetFormattedDateNow = func() FormattedDate {
	loadLocation, _ := time.LoadLocation(location)
	now := GetNow()
	return FormattedDate{Time: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loadLocation)}
}

func FormatDate(toFormart time.Time) string {
	location, _ := time.LoadLocation(location)
	return toFormart.In(location).Format("2006-01-02")
}

func GetDate(dateStr string) (time.Time, error) {
	dateLayout := "2006-01-02"
	if dateStr == "" {
		location, _ := time.LoadLocation(location)
		return time.Now().In(location), nil
	}

	return time.Parse(dateLayout, dateStr)
}

func GetDefaultLocation() *time.Location {
	location, err := time.LoadLocation(location)
	if err != nil {
		return time.UTC
	}
	return location
}
