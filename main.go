package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

var recordPath string // 记录文件路径
var delayDays int     // 容忍的延迟天数
var tolerance int     // 容忍的周期波动范围

func init() {
	flag.StringVar(&recordPath, "r", "./menstruation_records.csv", "导入的 csv 记录文件路径, eg: -r ./menstruation_records.csv")
	flag.IntVar(&delayDays, "d", 5, "容忍的月经延迟天数, e.g., -d 5")
	flag.IntVar(&tolerance, "t", 15, "周期波动容忍范围（如超过该范围的周期将被视为异常）, e.g., -t 15")
}

func main() {
	records, err := readRecordsFromCSV(recordPath)
	if err != nil {
		fmt.Printf("读取CSV文件时出错: %v\n", err)
		return
	}

	if len(records) < 2 {
		fmt.Println("没有足够的月经记录来计算周期长度。")
		return
	}

	dates := parseAndSortDates(records)
	cycleLengths, averageCycle, minCycle, maxCycle := calculateCycleStats(dates)

	fmt.Printf("计算出的周期长度（天）：%v\n", cycleLengths)
	fmt.Printf("平均周期长度：%.2f 天\n", averageCycle)
	fmt.Printf("最短周期长度：%d 天\n", minCycle)
	fmt.Printf("最长周期长度：%d 天\n", maxCycle)

	lastDate := dates[len(dates)-1]
	fmt.Printf("最后一次月经日期：%s\n", lastDate.Format("2006-01-02"))

	// 预测下一次月经日期范围
	predictedEarliest := lastDate.AddDate(0, 0, int(averageCycle)-delayDays)
	predictedLatest := lastDate.AddDate(0, 0, int(averageCycle)+delayDays)
	fmt.Printf("预测的下一次月经日期范围：%s 至 %s\n", predictedEarliest.Format("2006-01-02"), predictedLatest.Format("2006-01-02"))
}

// 解析并排序日期
func parseAndSortDates(records []MenstruationRecord) []time.Time {
	var dates []time.Time
	for _, rec := range records {
		if validDate(rec) {
			dates = append(dates, time.Date(rec.Year, time.Month(rec.Month), rec.Day, 0, 0, 0, 0, time.UTC))
		}
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	return dates
}

// 校验日期有效性
func validDate(rec MenstruationRecord) bool {
	return rec.Month >= 1 && rec.Month <= 12 && rec.Day >= 1 && rec.Day <= 31
}

// 计算周期统计数据
func calculateCycleStats(dates []time.Time) ([]int, float64, int, int) {
	var cycleLengths []int
	total := 0
	for i := 1; i < len(dates); i++ {
		days := daysBetween(dates[i-1], dates[i])
		if days >= tolerance && days <= 35 { // 过滤掉异常值
			cycleLengths = append(cycleLengths, days)
			total += days
		}
	}

	if len(cycleLengths) == 0 {
		fmt.Println("没有足够的有效月经周期记录来计算统计数据。")
		return nil, 0, 0, 0
	}

	averageCycle := float64(total) / float64(len(cycleLengths))
	minCycle, maxCycle := minMax(cycleLengths)
	return cycleLengths, averageCycle, minCycle, maxCycle
}

// 获取最小值和最大值
func minMax(cycles []int) (min, max int) {
	min, max = cycles[0], cycles[0]
	for _, cycle := range cycles[1:] {
		if cycle < min {
			min = cycle
		}
		if cycle > max {
			max = cycle
		}
	}
	return
}

// MenstruationRecord 结构体用于存储每次月经的年份、月份和日期
type MenstruationRecord struct {
	Year  int
	Month int
	Day   int
}

// 计算两日期之间的天数差
func daysBetween(start, end time.Time) int {
	return int(end.Sub(start).Hours() / 24)
}

// 从CSV文件读取月经记录
func readRecordsFromCSV(filePath string) ([]MenstruationRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var records [][]string
	if records, err = reader.ReadAll(); err != nil {
		return nil, fmt.Errorf("failed to read records: %v", err)
	}

	var menstruationRecords []MenstruationRecord
	for i := 0; i < len(records); i++ {
		if len(records[i]) != 3 {
			continue // 跳过无效记录
		}
		year, err1 := strconv.Atoi(records[i][0])
		month, err2 := strconv.Atoi(records[i][1])
		day, err3 := strconv.Atoi(records[i][2])
		if err1 != nil || err2 != nil || err3 != nil {
			continue // 跳过转换错误的记录
		}
		menstruationRecords = append(menstruationRecords, MenstruationRecord{
			Year:  year,
			Month: month,
			Day:   day,
		})
	}

	return menstruationRecords, nil
}
