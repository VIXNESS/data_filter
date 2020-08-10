package main

import (
	"encoding/base64"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	fileName := flag.String("f", "", "目标文件名")
	startTime := flag.String("s", "", "开始时间 格式 YYYY-MM-DD")
	endTime := flag.String("e", "", "结束时间 格式 YYYY-MM-DD")
	buffLen := flag.Int("b", 8192, "缓存大小, 默认 8192")
	h := flag.Bool("h", false, "帮助")
	help := flag.Bool("help", false, "帮助")
	flag.Parse()
	if *help || *h || *fileName == "" || *startTime == "" || *endTime == "" {
		fmt.Println("-f 文件名    输入的文件名", *fileName)
		fmt.Println("-s 开始时间    格式YYYY-MM-DD", *startTime)
		fmt.Println("-e 结束时间    格式YYYY-MM-DD", *endTime)
	} else {
		csvProcess(*fileName, *startTime, *endTime, *buffLen)
	}
}
func countCSVLines(inputFileName string) (uint64, error) {
	var lineCount uint64 = 0
	inputFile, inputFileErr := os.Open(inputFileName)
	if inputFileErr != nil {
		log.Fatalf("文件打开失败")
		return 0, inputFileErr
	}
	defer inputFile.Close()
	csvReader := csv.NewReader(inputFile)
	_, rowErr := csvReader.Read()
	fmt.Println("正在初始化文件")
	decoded, _ := base64.StdEncoding.DecodeString("IF8uXyAgICAgXywtJyIiYC0uXwooLC0uYC5fLCcoICAgICAgIHxcYC0vfAogICAgYC0uLScgXCApLWAoICwgbyBvKQogICAgICAgICAgYC0gICAgXGBfYCInLQ==")
	fmt.Printf("%s\n", string(decoded))
	start := time.Now()
	for rowErr != io.EOF {
		if rowErr != nil && rowErr != io.EOF {
			log.Fatal("读取错误")
			return 0, rowErr
		}
		_, rowErr = csvReader.Read()
		if lineCount%500000 == 0 {
			fmt.Print(".")
		}
		lineCount++
	}
	fmt.Printf("\n初始化完毕, 用时%s\n", time.Since(start))
	return lineCount, nil
}
func csvProcess(inputFileName string, begin string, end string, buffLength int) error {
	inputFile, inputFileErr := os.Open(inputFileName)
	if inputFileErr != nil {
		log.Fatalf("文件打开失败")
		return inputFileErr
	}
	defer inputFile.Close()

	outputFileName := strings.Split(inputFileName, ".")[0] + "." + begin + "_" + end + ".csv"
	outputFile, outputFileErr := os.Create(outputFileName)
	if outputFileErr != nil {
		log.Fatalf("文件输出出错")
	}
	defer outputFile.Close()
	outputFile.Seek(0, io.SeekEnd)
	writer := csv.NewWriter(outputFile)
	writer.Comma = ','
	writer.UseCRLF = true

	csvReader := csv.NewReader(inputFile)
	head, _ := csvReader.Read()
	// Write CSV Header
	writer.Write(head)
	writer.Flush()
	// Write CSV Main Data
	row, rowErr := csvReader.Read()
	var writerBuff [][]string
	var readerBuff [][]string
	count := 1

	var progress uint64 = 0
	csvLen, _ := countCSVLines(inputFileName)
	fmt.Println("开始查找数据")
	start := time.Now()
	beginTime, beginErr := time.Parse("2006-01-02", begin)
	if beginErr != nil {
		log.Fatalln("日期格式错误")
		return beginErr
	}
	endTime, endTimeErr := time.Parse("2006-01-02", end)
	if endTimeErr != nil {
		log.Fatalln("日期格式错误")
		return endTimeErr
	}
	doWrite := func() {
		writer.WriteAll(writerBuff)
		writer.Flush()
		writerBuff = nil
	}
	updateProgress := func() {
		progressBar := int(progress * 100 / csvLen)
		fmt.Fprintf(os.Stdout, "%d%% [%s]\r", progressBar, getS(progressBar/5, "#")+getS(20-progressBar/5, " "))
	}
	for rowErr != io.EOF {
		if rowErr != nil && rowErr != io.EOF {
			row, rowErr = csvReader.Read()
		}
		if count%buffLength != 0 {
			readerBuff = append(readerBuff, row)
			count++
		} else {
			for _, rb := range readerBuff {
				date, dateErr := time.Parse("2006-01-02", strings.Fields(rb[5])[0])
				if dateErr != nil {
					continue
				}
				if date.After(beginTime) && date.Before(endTime) {
					writerBuff = append(writerBuff, rb)
				}
			}
			count = 1
			readerBuff = nil
			if len(writerBuff) >= buffLength {
				doWrite()
			}

		}
		if progress%500000 == 0 {
			updateProgress()
		}
		progress++
		row, rowErr = csvReader.Read()
	}
	doWrite()
	updateProgress()
	decoded, _ := base64.StdEncoding.DecodeString("ICAgICAgXCAgICAvXAogICAgICAgKSAgKCAnKQogICAgICAoICAvICApCiAgICAgIFwoX18pfA==")
	fmt.Printf("\n\n\n%s\n完成\n", string(decoded))
	fmt.Println("输出文件名: ", outputFileName)
	fmt.Println("开始日期: ", begin, " 结束日期: ", end)
	fmt.Printf("用时: %s\n", time.Since(start))
	return nil
}

func getS(n int, char string) (s string) {
	for i := 1; i <= n; i++ {
		s += char
	}
	return
}
