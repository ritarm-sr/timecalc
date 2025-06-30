package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func parseTimeInput(s string) (int, error) {
	if strings.Contains(s, ":") {
		return parseTimeToMinutes(s)
	}

	floatVal, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("無効な時間形式: %s", s)
	}
	mins := int(floatVal * 60)
	return mins, nil
}

func parseTimeToMinutes(s string) (int, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("無効な形式: %s", s)
	}
	hours, err1 := strconv.Atoi(parts[0])
	minutes, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || minutes < 0 || minutes >= 60 {
		return 0, fmt.Errorf("無効な時刻値: %s", s)
	}
	return hours*60 + minutes, nil
}

func formatMinutesWithDecimal(total int) string {
	sign := ""
	minutesAbs := total
	if total < 0 {
		sign = "-"
		minutesAbs = -total
	}
	h := minutesAbs / 60
	m := minutesAbs % 60
	decimal := float64(minutesAbs) / 60.0
	return fmt.Sprintf("%s%d:%02d (%.3f)", sign, h, m, decimal)
}

func normalizeTokens(tokens []string) []string {
	var normalized []string
	for _, t := range tokens {
		switch strings.ToLower(t) {
		case "p":
			normalized = append(normalized, "+")
		case "m":
			normalized = append(normalized, "-")
		default:
			normalized = append(normalized, t)
		}
	}
	return normalized
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("使い方: timecalc [p|m] HH:MM or 1.5 ...")
		return
	}

	tokens := normalizeTokens(args)

	i := 0
	resultMinutes := 0

	if tokens[0] == "+" || tokens[0] == "-" {
	} else {
		mins, err := parseTimeInput(tokens[0])
		if err != nil {
			fmt.Println("エラー:", err)
			return
		}
		resultMinutes = mins
		i = 1
	}

	for i < len(tokens)-1 {
		op := tokens[i]
		val := tokens[i+1]
		mins, err := parseTimeInput(val)
		if err != nil {
			fmt.Printf("無効な時間: %s\n", val)
			return
		}
		switch op {
		case "+":
			resultMinutes += mins
		case "-":
			resultMinutes -= mins
		default:
			fmt.Printf("無効な演算子: %s\n", op)
			return
		}
		i += 2
	}

	fmt.Println(formatMinutesWithDecimal(resultMinutes))
}
