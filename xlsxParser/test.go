package main

import (
	"./xlsx"
	"fmt"
)

func main() {

	xlsFile, err := xlsx.OpenFile("citys.xlsx")
	if err != nil {
		panic(err)
	}

	for sheetIdx, _ := range xlsFile.Worksheets {
		sheet, err := xlsFile.GetSheetByIdx(sheetIdx)
		if err != nil {
			panic(err)
		}
		fmt.Printf("sheet idx:%d, name:%s, row:%d\n", sheetIdx, sheet.Name, sheet.MaxRow)
		for rowIdx, row := range sheet.Rows {
			for cellIdx, cell := range row.Cells {
				fmt.Printf("cell(%d,%d):%v\n", rowIdx, cellIdx, cell.Value)
			}
		}
	}
	//btw, it can be called like this
	//sheet, err := xlsFile.GetSheetByName("城市")
	//if err != nil {
	//	panic(err)
	//}
}
