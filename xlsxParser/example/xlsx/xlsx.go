package xlsx

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

type Sheet struct {
	Name     string
	filename string
	id       string
	relId    string
	Rows     []*Row
	File     *File
	MaxRow   int
	MaxCol   int
}

type cell struct {
	Name  string
	Type  string
	Value string
}

type Row struct {
	Num   int
	Cells map[int]*cell
}

type xlsxWorkbookRels struct {
	Relationships []xlsxWorkbookRelation `xml:"Relationship"`
}

// xmlxWorkbookRelation maps sheet id and xl/worksheets/sheet%d.xml
type xlsxWorkbookRelation struct {
	Id     string `xml:",attr"`
	Target string `xml:",attr"`
}

type File struct {
	filepath        string
	compressedFiles []zip.File
	Worksheets      []*Sheet
	sharedStrings   []string
	rels            map[string]string
}

func readWorkbook(d *xml.Decoder, s *File) []*Sheet {
	worksheets := make([]*Sheet, 0, 5)
	var (
		err   error
		token xml.Token
	)

	for {
		token, err = d.Token()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		switch x := token.(type) {
		case xml.StartElement:
			switch x.Name.Local {
			case "sheet":
				ws := new(Sheet)
				ws.File = s
				for _, a := range x.Attr {
					switch a.Name.Local {
					case "name":
						ws.Name = a.Value
					case "sheetId":
						ws.id = a.Value
					case "id":
						ws.relId = a.Value
					}
				}
				worksheets = append(worksheets, ws)
			}
		}
	}
	return worksheets
}

func readStrings(d *xml.Decoder, s *File) {
	var (
		err   error
		data  []byte
		token xml.Token
	)
	for {
		token, err = d.Token()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		switch x := token.(type) {
		case xml.StartElement:
			switch x.Name.Local {
			case "sst":
				// root element
				for i := 0; i < len(x.Attr); i++ {
					if x.Attr[i].Name.Local == "uniqueCount" {
						count, err := strconv.Atoi(x.Attr[i].Value)
						if err != nil {
							log.Fatal(err)
						}
						s.sharedStrings = make([]string, 0, count)
					}
				}
			default:
				// log.Println(x.Name.Local)
			}
		case xml.CharData:
			data = x.Copy()
		case xml.EndElement:
			switch x.Name.Local {
			case "t":
				s.sharedStrings = append(s.sharedStrings, string(data))
			}
		}

	}
}

func OpenFile(path string) (*File, error) {
	xlsx := new(File)
	xlsx.filepath = path

	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		switch f.Name {
		case "xl/workbook.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			xlsx.Worksheets = readWorkbook(xml.NewDecoder(rc), xlsx)
			rc.Close()
		case "xl/sharedStrings.xml":
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			readStrings(xml.NewDecoder(rc), xlsx)
			rc.Close()
		case "xl/_rels/workbook.xml.rels":
			rels, err := readWorkbookRelationsFromZipFile(f)
			if err != nil {
				panic(err)
			}
			xlsx.rels = rels
		}
	}
	return xlsx, nil
}

func readWorkbookRelationsFromZipFile(workbookRels *zip.File) (map[string]string, error) {
	var sheetXMLMap map[string]string
	var wbRelationships *xlsxWorkbookRels
	var rc io.ReadCloser
	var decoder *xml.Decoder
	var err error

	rc, err = workbookRels.Open()
	if err != nil {
		return nil, err
	}
	decoder = xml.NewDecoder(rc)
	wbRelationships = new(xlsxWorkbookRels)
	err = decoder.Decode(wbRelationships)
	if err != nil {
		return nil, err
	}
	sheetXMLMap = make(map[string]string)
	for _, rel := range wbRelationships.Relationships {
		if strings.HasSuffix(rel.Target, ".xml") && strings.HasPrefix(rel.Target, "worksheets/") {
			sheetXMLMap[rel.Id] = strings.Replace(rel.Target[len("worksheets/"):], ".xml", "", 1)
		}
	}
	return sheetXMLMap, nil
}

// func readWorksheetXML(dec *xml.Decoder, s *File) (map[int]*Row, error) {
func readWorksheetXML(dec *xml.Decoder, s *File) ([]*Row, error) {
	// rows := make(map[int]*Row)
	rows := []*Row{}
	var (
		err         error
		token       xml.Token
		rownum      int
		currentCell *cell
		currentRow  *Row
	)

	for {
		token, err = dec.Token()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		switch x := token.(type) {
		case xml.StartElement:
			switch x.Name.Local {
			case "row":
				currentRow = &Row{}
				currentRow.Cells = make(map[int]*cell)
				for _, a := range x.Attr {
					if a.Name.Local == "r" {
						rownum, err = strconv.Atoi(a.Value)
						if err != nil {
							return nil, err
						}
						//rownum--
					}
				}
				currentRow.Num = rownum
				rows = append(rows, currentRow)
				//rows[rownum-1] = currentRow
			case "c":
				currentCell = &cell{}
				var cellnumber rune
				for _, a := range x.Attr {
					switch a.Name.Local {
					case "r":
						for _, v := range a.Value {
							if v >= 'A' && v <= 'Z' {
								//cellnumber = cellnumber*26 + v - 'A'
								cellnumber = cellnumber*26 + v - 'A' + 1
							}
						}
					case "t":
						if a.Value == "s" {
							currentCell.Type = "s"
						}
					}

				}
				currentRow.Cells[int(cellnumber)-1] = currentCell
			}
		case xml.EndElement:
			switch x.Name.Local {
			case "c":
				currentCell = nil
			}
		case xml.CharData:
			if currentCell != nil {
				val := string(x.Copy())
				if currentCell.Type == "s" {
					valInt, _ := strconv.Atoi(val)
					currentCell.Value = s.sharedStrings[valInt]
				} else {
					currentCell.Value = val
				}
			}
		}
	}
	return rows, nil
}

func (ws *Sheet) readWorksheetZIP() error {
	r, err := zip.OpenReader(ws.File.filepath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == ws.filename {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			rows, err := readWorksheetXML(xml.NewDecoder(rc), ws.File)
			ws.Rows = rows
			ws.MaxRow = len(rows)
			if len(rows) > 0 {
				ws.MaxCol = len(rows[0].Cells)
			}
			//for _, row := range rows {
			//	if len(row.Cells) > ws.MaxCol {
			//		ws.MaxCol = len(row.Cells)
			//	}
			//}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ws *Sheet) Cell(row, column int) *cell {
	xrow := ws.Rows[row]
	if xrow == nil {
		return nil
	}
	if xrow.Cells[column] == nil {
		return nil
	}
	return xrow.Cells[column]
}

func (s *File) GetSheetByIdx(number int) (*Sheet, error) {
	if number >= len(s.Worksheets) || number < 0 {
		return nil, errors.New("Index out of range")
	}
	ws := s.Worksheets[number]
	if ws.Rows == nil {
		ws.filename = fmt.Sprintf("xl/worksheets/sheet%s.xml", ws.id)
		err := ws.readWorksheetZIP()
		if err != nil {
			return nil, err
		}
	}
	return ws, nil
}

func (s *File) GetSheetByName(name string) (*Sheet, error) {
	for _, ws := range s.Worksheets {
		if ws.Name == name {
			pageKey, ok := s.rels[ws.relId]
			if !ok {
				return nil, errors.New("page " + name + " not exist")
			}
			if ws.Rows == nil {
				ws.filename = fmt.Sprintf("xl/worksheets/%s.xml", pageKey)
				err := ws.readWorksheetZIP()
				if err != nil {
					return nil, err
				}
			}
			return ws, nil
		}
	}
	return nil, errors.New("Index out of range")

}
