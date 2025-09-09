package robot

import "encoding/xml"

type ShareLinkMessage struct {
	XMLName   xml.Name `xml:"appmsg"`
	Title     string   `xml:"title"`
	Des       string   `xml:"des"`
	Type      int      `xml:"type"`
	Url       string   `xml:"url"`
	ThumbUrl  string   `xml:"thumburl"`
	AppAttach string   `xml:"appattach"`
}
