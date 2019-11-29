package main


import (
	"fmt"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func main() {
	f := excelize.NewFile()
	// Create a new sheet.
	index := f.NewSheet("彩云请求次数统计")
	// Set value of a cell.
	f.SetCellValue("彩云请求次数统计", "A1", "2019-10-01")
	f.SetCellValue("彩云请求次数统计", "B1", 100)
	f.SetCellValue("彩云请求次数统计", "A2", "2019-10-02-2")
	f.SetCellValue("彩云请求次数统计", "B2", 1002134)
	// Set active sheet of the workbook.
	f.SetActiveSheet(index)

	index2 := f.NewSheet("彩云字符数统计")
	f.SetCellValue("彩云字符数统计", "A1", "2019-10-01")
	f.SetCellValue("彩云字符数统计", "B1", 10000000)
	f.SetActiveSheet(index2)

	// Save xlsx file by the given path.
	err := f.SaveAs("./Book5.xlsx")
	if err != nil {
		fmt.Println(err)
	}
}
