package utils

import "regexp"

// 校验每天几点几分
var dailyCron = regexp.MustCompile(`^(\d{1,2})\s+(\d{1,2})\s+\*\s+\*\s+\*$`)

func IsDailyAtHourMinute(expr string) bool {
	return dailyCron.MatchString(expr)
}

// 校验每周一几点几分
var weeklyMonCron = regexp.MustCompile(`^(\d{1,2})\s+(\d{1,2})\s+\*\s+\*\s+1$`)

func IsWeeklyMondayAtHourMinute(expr string) bool {
	return weeklyMonCron.MatchString(expr)
}

// 校验每月1号几点几分
var monthly1stCron = regexp.MustCompile(`^(\d{1,2})\s+(\d{1,2})\s+1\s+\*\s+\*$`)

func IsMonthly1stAtHourMinute(expr string) bool {
	return monthly1stCron.MatchString(expr)
}
