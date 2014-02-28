package main

import (
	"encoding/xml"
)

type FlaFolder struct {
	XMLName       xml.Name           `xml:"folders"`
	DOMFolderItem []FlaDOMFolderItem `xml:"DOMFolderItem"`
}

type FlaDOMFolderItem struct {
	XMLName    xml.Name `xml:"DOMFolderItem"`
	Name       string   `xml:"name,attr"`
	ItemID     string   `xml:"itemID,attr"`
	IsExpanded bool     `xml:"isExpanded,attr"`
}

type FlaDOMBitmapItem struct {
	XMLName                 xml.Name `xml:"DOMBitmapItem"`
	Name                    string   `xml:"name,attr"`
	ItemID                  string   `xml:"itemID,attr"`
	SourceExternalFilepath  string   `xml:"sourceExternalFilepath,attr"`
	SourceLastImported      string   `xml:"sourceLastImported,attr"`
	ExternalFileCRC32       uint32   `xml:"externalFileCRC32,attr"`
	ExternalFileSize        uint32   `xml:"externalFileSize,attr"`
	UseImportedJPEGData     bool     `xml:"useImportedJPEGData,attr"`
	CompressionType         string   `xml:"compressionType,attr"`
	OriginalCompressionType string   `xml:"originalCompressionType,attr"`
	Quality                 int      `xml:"quality,attr"`
	Href                    string   `xml:"href,attr"`
	BitmapDataHRef          string   `xml:"bitmapDataHRef,attr"`
	FrameRight              int      `xml:"frameRight,attr"`
	FrameBottom             int      `xml:"frameBottom,attr"`
	IsExpanded              bool     `xml:"isExpanded,attr"`
}

type FlaMedia struct {
	XMLName       xml.Name           `xml:"media"`
	DOMBitmapItem []FlaDOMBitmapItem `xml:"DOMBitmapItem"`
}

type FlaInclude struct {
	XMLName       xml.Name `xml:"Include"`
	Href          string   `xml:"href,attr"`
	ItemIcon      string   `xml:"itemIcon,attr"`
	LoadImmediate string   `xml:"loadImmediate,attr"`
	ItemID        string   `xml:"itemID,attr"`
	LastModified  int      `xml:"lastModified,attr"`
}

type FlaSymbols struct {
	XMLName xml.Name     `xml:"symbols"`
	Include []FlaInclude `xml:"Include"`
}

type FlaTimelines struct {
	XMLName xml.Name `xml:"timelines"`
}

type FlaPublishItem struct {
	XMLName     xml.Name `xml:"PublishItem"`
	PublishSize int      `xml:"publishSize,attr"`
	PublishTime int      `xml:"publishTime,attr"`
}

type FlaPublishHistory struct {
	XMLName     xml.Name         `xml:"publishHistory"`
	PublishItem []FlaPublishItem `xml:"PublishItem"`
}

type FlaPrinterSettings struct {
	XMLName xml.Name `xml:"PrinterSettings"`
}

type FlaDocument struct {
	XMLName         xml.Name           `xml:"DOMDocument"`
	Folders         FlaFolder          `xml:"folders"`
	Media           FlaMedia           `xml:"media"`
	Symbols         FlaSymbols         `xml:"symbols"`
	Timelines       FlaTimelines       `xml:"timelines"`
	PrinterSettings FlaPrinterSettings `xml:"PrinterSettings"`
	PublishHistory  FlaPublishHistory  `xml:"publishHistory"`
}

type DOMTimeline struct {
	XMLName xml.Name    `xml:"DOMTimeline"`
	Name    string      `xml:"name"`
	Layers  []DOMLayers `xml:"layers"`
}

type DOMLayers struct {
	XMLName  xml.Name   `xml:"layers"`
	Name     string     `xml:"name,attr"`
	DOMLayer []DOMLayer `xml:"DOMLayer"`
}

type DOMLayer struct {
	XMLName    xml.Name    `xml:"DOMLayer"`
	Name       string      `xml:"name,attr"`
	Color      string      `xml:"color,attr"`
	current    bool        `xml:"current,attr"`
	IsSelected bool        `xml:"isSelected,attr"`
	XMLName    []DomFrames `xml:"frames"`
}

type DomFrames struct {
	XMLName    xml.Name `xml:"frames"`
	Name       string   `xml:"name"`
	Color      string   `xml:"color"`
	current    bool     `xml:"current"`
	IsSelected bool     `xml:"isSelected"`
}
