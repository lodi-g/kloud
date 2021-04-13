package nextcloud

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strings"
)

const propfindPayload = `<?xml version="1.0"?>
<d:propfind xmlns:d="DAV:">
	<d:prop>
		<d:resourcetype />
		<d:getcontentlength />
	</d:prop>
</d:propfind>`

// GetRemoteFiles returns a list of the files in the remote NC server with their size
func (c *Client) GetRemoteFiles() (map[string]int64, error) {
	req, err := http.NewRequest("PROPFIND", c.server+"/public.php/webdav", strings.NewReader(propfindPayload))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.shareID, "")
	req.Header.Set("Depth", "10")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var files Files
	if err := xml.Unmarshal(body, &files); err != nil {
		return nil, err
	}

	ret := map[string]int64{}
	for _, file := range files {
		ret[file.Path] = file.Size
	}

	return ret, nil
}

// DownloadFile downloads a single file from its path and return the file's contents
func (c *Client) DownloadFile(fileName string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.server+"/public.php/webdav/"+fileName, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.shareID, "")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
