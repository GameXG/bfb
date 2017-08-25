package chromewebstore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/gamexg/proxyclient"
)

func ChromeUp(appId, token, filepath, proxy string) error {
	type UpdateRes struct {
		UploadState string `json:"uploadState"`
	}

	data, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer data.Close()

	if proxy == "" {
		proxy = "direct://0.0.0.0:0000"
	}
	p, err := proxyclient.NewProxyClient(proxy)
	if err != nil {
		return fmt.Errorf("代理字符串错误，%v", err)
	}

	c := http.Client{}
	c.Transport = &http.Transport{
		Dial: p.Dial,
	}

	req, err := http.NewRequest("PUT",
		fmt.Sprint("https://www.googleapis.com/upload/chromewebstore/v1.1/items/", appId),
		data)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprint("Bearer ", token))
	req.Header.Set("x-goog-api-version", "2")

	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("读 body 错误,%v", err)
	}
	switch res.StatusCode {
	case 200, 201, 204:
	default:
		return fmt.Errorf("服务器回应错误，status：%v  err:%v", res.Status, string(body))
	}
	updateRes := UpdateRes{}
	err = json.Unmarshal(body, &updateRes)
	if err != nil {
		return fmt.Errorf("json序列化错误 body:%v  ,%v", string(body), err)
	}
	if strings.ToLower(updateRes.UploadState) != strings.ToLower("SUCCESS") {
		return fmt.Errorf("上传chrome扩展错误,body:%v", string(body))
	}
	return nil

}

func Publish(appId, token, proxy string) error {
	type PublishRes struct {
		Status       []string `json:"status"`
		StatusDetail []string `json:"statusDetail"`
	}

	if proxy == "" {
		proxy = "direct://0.0.0.0:0000"
	}
	p, err := proxyclient.NewProxyClient(proxy)
	if err != nil {
		return fmt.Errorf("代理字符串错误，%v", err)
	}

	c := http.Client{}
	c.Transport = &http.Transport{
		Dial: p.Dial,
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("https://www.googleapis.com/chromewebstore/v1.1/items/%v/publish", appId),
		strings.NewReader(`{"target":"trustedTesters"}`))
	if err != nil {
		return fmt.Errorf("创建 Request错误 ,%v", err)
	}
	req.Header.Set("Authorization", fmt.Sprint("Bearer ", token))
	req.Header.Set("x-goog-api-version", "2")
	req.Header.Set("Authorization", fmt.Sprint("Bearer ", token))
	//req.Header.Set("publishTarget", "trustedTesters")
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("读 body 错误,%v", err)
	}

	switch res.StatusCode {
	case 200, 201, 204:
	default:
		return fmt.Errorf("服务器回应错误，status：%v  err:%v", res.Status, string(body))
	}

	publishRes := PublishRes{}
	err = json.Unmarshal(body, &publishRes)
	if err != nil {
		return fmt.Errorf("json序列化错误 body:%v  ,%v", string(body), err)
	}
	if reflect.DeepEqual(publishRes.Status, []string{"OK"}) == false {
		return fmt.Errorf("上传chrome扩展错误,body:%v", string(body))
	}

	return nil
}
