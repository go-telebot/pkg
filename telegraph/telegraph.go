package telegraph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

// Upload uploads a file to Telegraph using the passed reader as a source.
func Upload(r io.Reader) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "file")
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(part, r); err != nil {
		return "", err
	}
	if err = writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, "https://telegra.ph/upload", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var uploads []struct {
		Path string `json:"src"`
	}
	if err := json.Unmarshal(data, &uploads); err != nil {
		m := map[string]string{}
		if err := json.Unmarshal(data, &m); err != nil {
			return "", err
		}
		return "", fmt.Errorf("telegraph: %s", m["error"])
	}

	return "https://telegra.ph/" + uploads[0].Path, nil
}

// UploadFile uploads a file to Telegraph from a disk.
func UploadFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return Upload(file)
}
