package oauth2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gamexg/proxyclient"
	"github.com/go-xweb/log"
)

// https://developer.chrome.com/webstore/using_webstore_api
// 仅仅返回 code 授权地址，需要用户通过浏览器访问
func GetCodeUrl(clientID string) string {
	return fmt.Sprint("https://accounts.google.com/o/oauth2/auth?response_type=code&scope=https://www.googleapis.com/auth/chromewebstore&client_id=" + clientID + "&redirect_uri=urn:ietf:wg:oauth:2.0:oob")
}

//
// 返回值
//   1    	access_token			30分钟内有效，超时后需要通过  RefreshToken(refresh_token) 获得新的。
//	 2		refresh_token			只有第一次授权才会获得，如果第一次忘了保存，需要取消授权后再次执行授权。
func GetRefreshToken(client_id, client_secret, code, proxy string) (string, string, error) {
	log.Debug("刷新 Token ...")
	if proxy == "" {
		proxy = "direct://0.0.0.0:0000"
	}
	p, err := proxyclient.NewProxyClient(proxy)
	if err != nil {
		return "", "", fmt.Errorf("代理字符串错误，%v", err)
	}

	c := http.Client{}
	c.Transport = &http.Transport{
		Dial: p.Dial,
	}
	// "client_id=$CLIENT_ID&client_secret=$CLIENT_SECRET&code=$CODE&grant_type=authorization_code&redirect_uri=urn:ietf:wg:oauth:2.0:oob"
	vs := url.Values{
		"client_id":     {client_id},
		"client_secret": {client_secret},
		"code":          {code},
		"grant_type":    {"refresh_token"},
		"redirect_uri":  {"urn:ietf:wg:oauth:2.0:oob"},
	}
	res, err := c.PostForm("https://accounts.google.com/o/oauth2/token", vs)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", fmt.Errorf("读 body 错误,%v", err)
	}

	if res.StatusCode != 200 {
		return "", "", fmt.Errorf("服务器回应错误，status：%v  err:%v", res.Status, string(body))
	}

	r := struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"` // Bearer
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}{}

	err = json.Unmarshal(body, &r)
	if err != nil {
		return "", "", fmt.Errorf("json序列化错误 body:%v  ,%v", string(body), err)
	}

	return r.AccessToken, r.RefreshToken, nil
}

func RefreshToken(ClientId, ClientSecret, RefreshToken, proxy string) (string, error) {
	log.Debug("刷新 Token ...")
	if proxy == "" {
		proxy = "direct://0.0.0.0:0000"
	}
	p, err := proxyclient.NewProxyClient(proxy)
	if err != nil {
		return "", fmt.Errorf("代理字符串错误，%v", err)
	}

	c := http.Client{}
	c.Transport = &http.Transport{
		Dial: p.Dial,
	}
	vs := url.Values{
		"client_id":     {ClientId},
		"client_secret": {ClientSecret},
		"refresh_token": {RefreshToken},
		"grant_type":    {"refresh_token"},
	}
	res, err := c.PostForm("https://www.googleapis.com/oauth2/v4/token", vs)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("读 body 错误,%v", err)
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("服务器回应错误，status：%v  err:%v", res.Status, string(body))
	}

	r := struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}{}

	err = json.Unmarshal(body, &r)
	if err != nil {
		return "", fmt.Errorf("json序列化错误 body:%v  ,%v", string(body), err)
	}

	return r.AccessToken, nil
}
