package main

import (
	"./xlsx"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type nodeItem struct {
	XMLName    xml.Name `xml:"item"`
	Type       string   `xml:"type,attr"`
	Name       string   `xml:"name,attr"`
	Code       string   `xml:"code,attr"`
	Replace    string   `xml:"replace,attr"`
	Gap        string   `xml:"gap,attr"`
	CellIdxs   []int
	Names      []string
	Marked     bool
	ReplaceArr []string
	ReplaceIdx int
	HasReplace bool
}

type TblNode struct {
	XMLName xml.Name   `xml:"node"`
	Items   []nodeItem `xml:"item"`
	Type    string     `xml:"item,attr"`
	Xlsx    string     `xml:"xlsx,attr"`
	Page    string     `xml:"page,attr"`
	Out     string     `xml:"out,attr"`
	Code    string     `xml:"code,attr"`
	Key     string     `xml:"key,attr"`
	ItemNum int
}

type ConfigData struct {
	XMLName     xml.Name  `xml:"data"`
	Data        []TblNode `xml:"node"`
	Description string    `xml:",innerxml"`
}

/**
 * As文件结构
 */
type ASReader struct {
	Prop   string
	Reader string
}
type AsVar struct {
	Info string
	Prop string
	Type string
}
type AsStruct struct {
	ClsName string
	Info    string
	Prop    string
	Reader  string
}

/**
 * 获取文件内容
 */
func getBytes(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getContent(path string) (string, error) {
	file, err := getBytes(path)
	if err != nil {
		return "", err
	}
	return string(file[:]), err
}

// func saveContent(path string, str string) error {
// 	return ioutil.WriteFile(path, str, 0644)
// }

var xlsxHash map[string]xlsx.File

/**
 * 获取xlsx，并解析
 */
func getXlsx(path string) (*xlsx.File, error) {
	if xlsxHash == nil {
		xlsxHash = make(map[string]xlsx.File)
	}
	xlsFile, ok := xlsxHash[path]
	if ok {
		return &xlsFile, nil
	}

	t1 := time.Now()
	xlFile, err := xlsx.OpenFile(path)
	if err != nil {
		fmt.Println("error")
		return xlFile, err
	}
	t2 := time.Now()
	fmt.Printf("parse xlsx %s, cost:%fs\n", path, t2.Sub(t1).Seconds())
	xlsxHash[path] = *xlFile
	return xlFile, nil
}

/**
 * 获取列索引
 */
func getRowHeadIdx(head *xlsx.Row, val string) (int, error) {
	for idx, headVal := range head.Cells {
		if headVal.Value == val {
			return idx, nil
		}
	}
	return 0, errors.New("没有列" + val)
}

var pageFieldHash map[string](map[string]int)

/**
 * 属性列缓冲
 */
func getPageField(page *xlsx.Sheet, path []string) (map[string]int, error) {
	if pageFieldHash == nil {
		pageFieldHash = make(map[string](map[string]int))
	}
	key := fmt.Sprintf("%s|%s|%s", path[0], path[1], path[2])
	filedHash, ok := pageFieldHash[key]
	if ok {
		return filedHash, nil
	}
	newFiledHash := make(map[string]int)
	lineIdx, err := getRowHeadIdx(page.Rows[0], path[2])
	if err != nil {
		return newFiledHash, err
	}
	count := page.MaxRow
	for i := 1; i < count; i++ {
		cell := page.Cell(i, lineIdx)
		if cell != nil {
			newFiledHash[cell.Value] = i
		}
	}

	pageFieldHash[key] = newFiledHash

	return newFiledHash, nil
}

/**
 * 从别的表中按替换规则进行替换
 */
func replaceItem(item nodeItem, val string) (string, error) {
	//fmt.Printf("%s Replace:%s\n", item.Name, item.Replace)
	xlsFile, err := getXlsx(item.ReplaceArr[0])
	if err != nil {
		return "", err
	}
	page, err := xlsFile.GetSheetByName(item.ReplaceArr[1])
	if err != nil {
		return "", errors.New("表" + item.ReplaceArr[0] + "中没有页" + item.ReplaceArr[1])
	}
	fieldHash, err := getPageField(page, item.ReplaceArr)
	if err != nil {
		return "", err
	}
	colIdx := item.ReplaceIdx
	if colIdx < 0 {
		idx, err := getRowHeadIdx(page.Rows[0], item.ReplaceArr[3])
		if err != nil {
			return val, nil
		}
		colIdx = idx
	}
	rowIdx, ok := fieldHash[val]
	if !ok {
		return val, errors.New("替换表中没有值" + val)
	}
	return page.Cell(rowIdx, colIdx).Value, nil
}

/**
 * 写入节点数据
 */
func writeItem(tblItem nodeItem, buf *bytes.Buffer, val string) error {
	//fmt.Printf("type:%s, %s:%s\n", tblItem.Type, tblItem.Name, val)
	if tblItem.HasReplace {
		rval, err := replaceItem(tblItem, val)
		if err != nil {
			return err
		}
		val = rval
	}

	if tblItem.Type == "string" {
		err := binary.Write(buf, binary.BigEndian, int16(len(val)))
		if err != nil {
			return err
		}
		buf.WriteString(val)
	} else {
		intVal, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			return err
		}
		switch tblItem.Type {
		case "byte":
			err := binary.Write(buf, binary.BigEndian, int8(intVal))
			if err != nil {
				return err
			}
		case "short":
			err := binary.Write(buf, binary.BigEndian, int16(intVal))
			if err != nil {
				return err
			}
		case "int":
			err := binary.Write(buf, binary.BigEndian, int32(intVal))
			if err != nil {
				return err
			}
		default:
			return errors.New("type error " + tblItem.Type)
		}
	}
	return nil
}

func getAsType(str string) string {
	if str == "string" {
		return "String"
	}
	return "int"
}
func getAsReader(str string) string {
	switch str {
	case "byte":
		return "bytes.readUnsignedByte()"
	case "short":
		return "bytes.readUnsignedShort()"
	case "int":
		return "bytes.readInt()"
	case "string":
		return "bytes.readUTFBytes(bytes.readUnsignedShort())"
	}
	return ""
}
func checkInputType(str string) bool {
	switch str {
	case "byte":
		return true
	case "short":
		return true
	case "int":
		return true
	case "string":
		return true
	}
	return false
}

func createAs(tbl *TblNode) {
	content, err := getContent("template.as")
	if err != nil {
		fmt.Println("template.as is not exist!")
	}
	contents := strings.Split(content, "||")
	fileTpl := contents[0]
	varTpl := contents[1]
	readTpl := contents[2]

	startIdx := strings.LastIndex(tbl.Code, "/") + 1
	endIdx := strings.LastIndex(tbl.Code, ".")
	clsName := tbl.Code[startIdx:endIdx]
	filePath := tbl.Code[0:startIdx]
	os.MkdirAll(filePath, os.ModeDir)

	fileTpl = strings.Replace(fileTpl, "$Info", tbl.Xlsx+" - "+tbl.Page, -1)
	fileTpl = strings.Replace(fileTpl, "$ClsName", clsName, -1)
	fileTpl = strings.Replace(fileTpl, "$TblUrl", tbl.Out[strings.LastIndex(tbl.Out, "/")+1:], -1)
	//fileTpl = strings.Replace(fileTpl, "$ItemNum", strconv.Itoa(tbl.ItemNum), -1)
	if tbl.Key != "" {
		fileTpl = strings.Replace(fileTpl, "$DicType", "Dictionary", -1)
		fileTpl = strings.Replace(fileTpl, "$DicAdd", contents[3], -1)
		fileTpl = strings.Replace(fileTpl, "$DicKey", tbl.Key, -1)
	} else {
		fileTpl = strings.Replace(fileTpl, "$DicType", "Vector.<"+clsName+">", -1)
		fileTpl = strings.Replace(fileTpl, "$DicAdd", contents[4], -1)
	}

	vars := []string{}
	readers := []string{}

	for _, val := range tbl.Items {
		newVar := strings.Replace(varTpl, "$Info", val.Name, -1)
		newVar = strings.Replace(newVar, "$Prop", val.Code, -1)
		newVar = strings.Replace(newVar, "$Type", getAsType(val.Type), -1)
		vars = append(vars, newVar)

		newRead := strings.Replace(readTpl, "$Prop", val.Code, -1)
		newRead = strings.Replace(newRead, "$Reader", getAsReader(val.Type), -1)
		readers = append(readers, newRead)
	}
	fileTpl = strings.Replace(fileTpl, "$Prop", strings.Join(vars, ""), -1)
	fileTpl = strings.Replace(fileTpl, "$Reader", strings.Join(readers, ""), -1)

	errFile := ioutil.WriteFile(tbl.Code, []byte(fileTpl), 0644)
	if errFile != nil {
		fmt.Println("error:", errFile)
	}
}

func arrayIndex(sName []string, pName string) int {
	for idx, _ := range sName {
		if sName[idx] == pName {
			return idx
		}
	}
	return -1
}

func main() {
	fromPath := flag.String("from", "config.xml", "from info")
	flag.Parse()
	configPath := *fromPath

	data, err := getBytes(configPath)
	if err != nil {
		fmt.Printf("error:%v\n", err)
		return
	}
	v := ConfigData{}
	err = xml.Unmarshal(data, &v)
	if err != nil {
		fmt.Printf("error:%v\n", err)
		return
	}
	//解析xml里各节点
	for _, tbl := range v.Data {
		if _, err := os.Stat(tbl.Xlsx); err != nil {
			fmt.Printf("%s is not exist!\n", tbl.Xlsx)
			continue
		}
		//errorType := 0

		xlsFile, err := getXlsx(tbl.Xlsx)
		if err != nil {
			// fmt.Printf("parse xlsx error:%v\n", err)
			fmt.Printf("parse xlsx error\n")
			continue
		}
		page, err := xlsFile.GetSheetByName(tbl.Page)
		if err != nil {
			fmt.Printf("page not exist!\n")
			continue
		}
		//取出索引值
		fmt.Printf("maxRow:%d, maxCol:%d, page:%s\n", page.MaxRow, page.MaxCol, tbl.Page)

		head := page.Rows[0]
		count := len(tbl.Items)
		for i := 0; i < count; i++ {
			tbl.Items[i].Names = strings.Split(tbl.Items[i].Name, ";")
			nameLen := int(len(tbl.Items[i].Names))
			tbl.Items[i].CellIdxs = []int{}
			for j := 0; j < nameLen; j++ {
				tbl.Items[i].CellIdxs = append(tbl.Items[i].CellIdxs, 0)
			}
			tbl.Items[i].Type = strings.ToLower(tbl.Items[i].Type)
			if !checkInputType(tbl.Items[i].Type) {
				fmt.Printf("type is error! in cell %s\n", tbl.Items[i].Name)
				return
			}
		}
		for idx, headVal := range head.Cells {
			for i := 0; i < count; i++ {
				//fmt.Printf("get it -> idx:%d,value:%s\n", i, tbl.Items[i].Name)
				aidx := arrayIndex(tbl.Items[i].Names, headVal.Value)
				if aidx >= 0 {
					tbl.Items[i].CellIdxs[aidx] = idx
					tbl.Items[i].Marked = true
				}
				//if headVal.Value == tbl.Items[i].Name {
				//	tbl.Items[i].CellIdx = idx
				//	tbl.Items[i].Marked = true
				//	//fmt.Printf("head(%d):%s, code:%s\n", tblItem.CellIdx, headVal, tblItem.Code)
				//}
			}
		}
		for idx, tblItem := range tbl.Items {
			if len(tblItem.Replace) > 0 {
				tbl.Items[idx].HasReplace = true
				tbl.Items[idx].ReplaceIdx = -1
				tbl.Items[idx].ReplaceArr = strings.Split(tblItem.Replace, "|")
			}
			if !tblItem.Marked {
				fmt.Printf("列[%s]在表[%s]中不存在\n", tblItem.Name, tbl.Xlsx)
				continue
			}
		}
		buf := new(bytes.Buffer)
		tbl.ItemNum = 0
		fmt.Println(page.MaxRow)
		for i := 1; i < page.MaxRow; i++ {
			line := page.Rows[i]
			if _, ok := line.Cells[0]; ok && line.Cells[0].Value != "" {
				for _, tblItem := range tbl.Items {
					val := ""
					for _, cidx := range tblItem.CellIdxs {
						cellNode := line.Cells[cidx]
						if cellNode != nil {
							val += cellNode.Value
						}
					}
					//val := line.Cells[tblItem.CellIdx].Value
					err := writeItem(tblItem, buf, val)
					if err != nil {
						fmt.Println("binary.Write failed:", err)
					}
				}
				tbl.ItemNum++
			}
		}
		createAs(&tbl)

		startIdx := strings.LastIndex(tbl.Out, "/") + 1
		filePath := tbl.Out[0:startIdx]
		fmt.Printf("save %s at %s\n", tbl.Out, filePath)
		os.MkdirAll(filePath, os.ModeDir)
		errFile := ioutil.WriteFile(tbl.Out, buf.Bytes(), 0644)
		if errFile != nil {
			fmt.Println("error:", errFile)
		}

		fmt.Printf("bytes len:%d\n", len(buf.Bytes()))
	}
}
