package nextcloud

import (
	"encoding/xml"
	"net/url"
	"strings"
)

// File is a NextCloud remote file
type File struct {
	Path string
	Size int64
}

// Files is an array of File
type Files []File

// UnmarshalXML parses the XML response from the DAV server and transforms it to an array of file
func (files *Files) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	cxd := struct {
		XMLName   xml.Name `xml:"multistatus"`
		Responses []struct {
			XMLName    xml.Name `xml:"response"`
			Href       string   `xml:"href"`
			Collection xml.Name `xml:"propstat>prop>resourcetype>collection"`
			Size       int64    `xml:"propstat>prop>getcontentlength"`
		} `xml:"response"`
	}{}

	if err := d.DecodeElement(&cxd, &start); err != nil {
		return err
	}

	// Iterate through results, appending them to the files list
	for _, resp := range cxd.Responses[1:] {
		// Do not process directories
		if resp.Collection.Local == "collection" {
			continue
		}

		href := strings.ReplaceAll(resp.Href, "/public.php/webdav/", "")
		decodedHref, err := url.QueryUnescape(href)
		if err != nil {
			return err
		}

		*files = append(*files, File{decodedHref, resp.Size})
	}

	return nil
}
