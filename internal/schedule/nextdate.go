package schedule

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	ErrBadDate  = errors.New("bad date format, want YYYYMMDD")
	ErrBadRule  = errors.New("bad repeat rule")
	ErrBadDaysN = errors.New("bad N in 'd N'")
)

const ymd = "20060102"

func NextDate(date, rule string) (string, error) {
	return NextAfter(date, rule, date)
}

func NextAfter(date, rule, now string) (string, error) {
	base, err := time.Parse(ymd, date)
	if err != nil {
		log.Printf("schedule.NextAfter: parse date=%q error: %v", date, err)
		return "", ErrBadDate
	}
	ref := base
	if strings.TrimSpace(now) != "" {
		ref, err = time.Parse(ymd, now)
		if err != nil {
			log.Printf("schedule.NextAfter: parse now=%q error: %v", now, err)
			return "", ErrBadDate
		}
	}

	rule = strings.TrimSpace(rule)
	switch {
	case rule == "y":
		m := base.Month()
		d := base.Day()

		var cand time.Time
		if base.Year() >= ref.Year() {
			// всегда следующий год от самой даты
			cand = makeYearDate(base.Year()+1, m, d)
		} else {
			// переносим в год now; если не строго после now — год +1
			cand = makeYearDate(ref.Year(), m, d)
			if !cand.After(ref) {
				cand = makeYearDate(ref.Year()+1, m, d)
			}
		}
		return cand.Format(ymd), nil

	case strings.HasPrefix(rule, "d "):
		parts := strings.Fields(rule)
		if len(parts) != 2 {
			log.Printf("schedule.NextAfter: bad rule format (no N): %q", rule)
			return "", ErrBadRule
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil || n <= 0 || n > 400 {
			log.Printf("schedule.NextAfter: bad N in 'd N' (got %q): %v", parts[1], err)
			return "", ErrBadDaysN
		}
		cand := base.AddDate(0, 0, n)
		for !cand.After(ref) {
			cand = cand.AddDate(0, 0, n)
		}
		return cand.Format(ymd), nil

	default:
		log.Printf("schedule.NextAfter: bad repeat rule: %q", rule)
		return "", fmt.Errorf("%w: %q", ErrBadRule, rule)
	}
}

// 29 февраля -> 1 марта
func makeYearDate(year int, m time.Month, d int) time.Time {
	if m == time.February && d == 29 && !isLeap(year) {
		return time.Date(year, time.March, 1, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(year, m, d, 0, 0, 0, 0, time.UTC)
}

func isLeap(year int) bool {
	if year%400 == 0 {
		return true
	}
	if year%100 == 0 {
		return false
	}
	return year%4 == 0
}
