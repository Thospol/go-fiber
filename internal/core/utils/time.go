package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	languageTh string = "th"
)

// LoadLocation returns The time zone.
func LoadLocation() *time.Location {
	timeZone, _ := time.LoadLocation("Asia/Bangkok")
	return timeZone
}

// FormatDate format date
func FormatDate(dt *time.Time) (string, string) {
	datefmt := dt.In(LoadLocation()).Format("02 January 2006")

	return dateFormat(datefmt, "en"), dateFormat(datefmt, "th")
}

func dateFormat(datefmt, language string) string {
	if language == languageTh {
		switch getMonth(datefmt) {
		case "January":
			datefmt = strings.Replace(datefmt, "January", "มกราคม", 1)

		case "February":
			datefmt = strings.Replace(datefmt, "February", "กุมภาพันธ์", 1)

		case "March":
			datefmt = strings.Replace(datefmt, "March", "มีนาคม", 1)

		case "April":
			datefmt = strings.Replace(datefmt, "April", "เมษายน", 1)

		case "May":
			datefmt = strings.Replace(datefmt, "May", "พฤษภาคม", 1)

		case "June":
			datefmt = strings.Replace(datefmt, "June", "มิถุนายน", 1)

		case "July":
			datefmt = strings.Replace(datefmt, "July", "กรกฎาคม", 1)

		case "August":
			datefmt = strings.Replace(datefmt, "August", "สิงหาคม", 1)

		case "September":
			datefmt = strings.Replace(datefmt, "September", "กันยายน", 1)

		case "October":
			datefmt = strings.Replace(datefmt, "October", "ตุลาคม", 1)

		case "November":
			datefmt = strings.Replace(datefmt, "November", "พฤศจิกายน", 1)

		case "December":
			datefmt = strings.Replace(datefmt, "December", "ธันวาคม", 1)
		}
	}

	return fmt.Sprintf("%s %s %s", getDay(datefmt), getMonth(datefmt), getYear(datefmt, language))
}

func getDay(datefmt string) string {
	return strings.Split(datefmt, " ")[0]
}

func getMonth(datefmt string) string {
	return strings.Split(datefmt, " ")[1]
}

func getYear(datefmt, language string) string {
	year, _ := strconv.Atoi(strings.Split(datefmt, " ")[2])
	if language == languageTh {
		return strconv.Itoa(year + 543)
	}
	return strconv.Itoa(year)
}

// GetToday get datetime today timezone (th)
func GetToday() time.Time {
	timeNow := TimeNowLocationTH()
	return time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location())
}

// GetYesterday get datetime yesterday timezone (th)
func GetYesterday() time.Time {
	timeNow := TimeNowLocationTH()
	return time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day()-1, 0, 0, 0, 0, timeNow.Location())
}

// GetNextDateTime get datetime tomorrow timezone (th)
func GetNextDateTime(year, month, day int) time.Time {
	timeNow := TimeNowLocationTH()
	return time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location()).
		AddDate(year, month, day)
}

// GetPreviousDateTime get datetime yesterday timezone (th)
func GetPreviousDateTime(day int) time.Time {
	timeNow := TimeNowLocationTH()
	return time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day()-day, 0, 0, 0, 0, timeNow.Location())
}

// GetDateTimeByDate get datetime by date timezone (th)
func GetDateTimeByDate(day time.Time) time.Time {
	timeNow := TimeNowLocationTH()
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, timeNow.Location())
}

// DaysIn returns the number of days in a month for a given year.
func DaysIn(year int, m time.Month) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, LoadLocation()).Day()
}

// IsValidDay check if day is more than 6(Saturday time.Weekday)
func IsValidDay(value time.Weekday) bool {
	return value >= time.Sunday && value <= time.Saturday
}

// IsValidTime check if open and close time is in hours: 00-23 and minutes: 00-59
func IsValidTime(openTime, closeTime string) bool {
	ot := strings.Split(openTime, ":")
	ct := strings.Split(closeTime, ":")
	if len(ot) != 2 || len(ct) != 2 {
		return false
	}

	openHr := ot[0]
	openM := ot[1]
	closeHr := ct[0]
	closeM := ct[1]

	if !((openHr >= "0" && openHr <= "23") || (openM >= "0" && openM <= "59")) ||
		!((closeHr >= "0" && closeHr <= "23") || (closeM >= "0" && closeM <= "59")) {
		return false
	}

	return true
}

// TimeNowLocationTH get time location thai
func TimeNowLocationTH() time.Time {
	return time.Now().In(LoadLocation())
}

// GetDate get date from format string
func GetDate(date string) string {
	if date != "" {
		s := strings.Split(date, "T")
		if len(s) > 1 {
			date = s[0]
		}
	}
	return date
}

// NowWhichNonZeroMilliseconds current time which none zero milliseconds
func NowWhichNonZeroMilliseconds() time.Time {
	return addMilliseconds(time.Now())
}

func addMilliseconds(t time.Time) time.Time {
	zeroFormat := t.Format("2006-01-02T15:04:05.000Z")
	format := t.Format("2006-01-02T15:04:05.999Z")
	if format != zeroFormat {
		return t.Add(1 * time.Millisecond)
	}
	return t
}
