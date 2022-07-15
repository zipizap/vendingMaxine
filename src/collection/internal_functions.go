package collection

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
)

// https://gist.github.com/alex-ant/aeaaf497055590dacba760af24839b8d
func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	resData = resB.Bytes()

	return
}

func gZipData(data []byte) (compressedData []byte, err error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	_, err = gz.Write(data)
	if err != nil {
		return
	}

	if err = gz.Flush(); err != nil {
		return
	}

	if err = gz.Close(); err != nil {
		return
	}

	compressedData = b.Bytes()

	return
}

func encode_bytes_to_gzB64(bytes []byte) (gzB64 string, err error) {
	gz, err := gZipData(bytes)
	if err != nil {
		return "", err
	}
	gzB64 = base64.StdEncoding.EncodeToString(gz)
	return gzB64, nil
}

func decode_gzB64_to_bytes(gzB64 string) (bytes []byte, err error) {
	gz, err := base64.StdEncoding.DecodeString(gzB64)
	if err != nil {
		return nil, err
	}
	bytes, err = gUnzipData(gz)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
