package app

import (
	"bytes"
	e "drafter/exception"
	. "drafter/setting"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

type qiniuUploadResponse struct {
	Key  string `json:"key"`
	Hash string `json:"hash"`
}

type uploadImageResponse struct {
	Url string `json:"url"`
}

// a multipart/form-data body require 'file' property
func UploadImageHandler(ctx *context) (err error) {
	f, header, err := ctx.Req.FormFile("file")
	if err != nil {
		return
	}
	token := generateQiniuToken("mrzie", Settings.Qiniu.AK, Settings.Qiniu.SK)

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	fw, err := w.CreateFormFile("file", header.Filename)
	if err != nil {
		return
	}
	_, err = io.Copy(fw, f)
	if err != nil {
		return
	}
	w.WriteField("token", token)
	w.Close()
	req, err := http.NewRequest("POST", "http://up.qiniup.com/", buf)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	var client http.Client
	res, err := client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode >= 300 || res.StatusCode < 200 {
		return e.UploadImageFail()
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	res.Body.Close()
	var qiniuRes qiniuUploadResponse
	err = json.Unmarshal(resBody, &qiniuRes)
	if err != nil {
		return
	}
	return ctx.SendJson(uploadImageResponse{Url: "http://" + Settings.Qiniu.Domain + "/" + qiniuRes.Key})
	// io.Copy(os.Stderr, res.Body) // Replace this with Status.Code check
	// return
}

func generateQiniuToken(bucket string, accessKey string, secretKey string) string {
	// 简单上传凭证
	// 默认凭证有效期一小时
	putPolicy := storage.PutPolicy{
		Scope:   bucket,
		SaveKey: "image/${etag}",
	}
	mac := qbox.NewMac(accessKey, secretKey)
	return putPolicy.UploadToken(mac)
}
