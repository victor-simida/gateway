package server
import (
	"gateway/web"
	"gateway/config"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"io/ioutil"
	"crypto/tls"
)

func init() {
	for _,v := range config.Settings.HttpAddr {
		url := fmt.Sprintf("/gateway/%s",v.Suffix)
		web.RegisterHandler(url, httpHandler)
	}

	for _,v := range config.Settings.HttpsAddr {
		url := fmt.Sprintf("/gateway/%s",v.Suffix)
		web.RegisterHandler(url, httpsHandler)
	}
}

/*******************************************
*函数名：httpsHandler
*作用：转换地址并改为https请求
*时间：2016/1/19 20:08
*入参：
*返回值：
*******************************************/
func httpsHandler(w http.ResponseWriter, req *http.Request){
	var urlStr string
	for _, v := range config.Settings.HttpsAddr {
		if !strings.Contains(req.URL.Path, v.Suffix) {
			continue
		}
		urlStr = fmt.Sprintf("https://%s/%s", v.Prefix, v.Suffix)
		break
	}
	if urlStr == "" {
		w.Write([]byte{})
		return
	}

	result,err := TransToHttps(urlStr,req)
	if err != nil {
		result = []byte{}
	}
	w.Write(result)
}

/*******************************************
*函数名：httpHandler
*作用：转换地址并改为http请求
*时间：2016/1/19 20:08
*入参：
*返回值：
*******************************************/
func httpHandler(w http.ResponseWriter, req *http.Request){
	var urlStr string
	for _, v := range config.Settings.HttpAddr {
		if !strings.Contains(req.URL.Path, v.Suffix) {
			continue
		}
		urlStr = fmt.Sprintf("http://%s/%s",  v.Prefix, v.Suffix)
		break
	}
	if urlStr == "" {
		w.Write([]byte{})
		return
	}

	result,err := TransToHttp(urlStr,req)
	if err != nil {
		result = []byte{}
	}
	w.Write(result)

}

/*******************************************
*函数名：TransToHttps
*作用：克隆请求结构体，并改为发起新的https请求
*时间：2016/1/19 20:25
*入参：
*返回值：
*******************************************/
func TransToHttps(urlStr string, req *http.Request) ([]byte, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return []byte{}, err
	}

	temp := *req

	new_req := &temp
	new_req.URL = u
	new_req.Host = u.Host
	new_req.RequestURI = ""			//RequestURI需要设置为空，否则会报错

	/*https默认不校验tls*/
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport:tr}
	resp, err := client.Do(new_req)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return body,nil
}

/*******************************************
*函数名：TransToHttp
*作用：克隆请求结构体，并改为发起新的http请求
*时间：2016/1/19 20:25
*入参：
*返回值：
*******************************************/
func TransToHttp(urlStr string, req *http.Request) ([]byte, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return []byte{}, err
	}

	temp := *req

	new_req := &temp
	new_req.URL = u
	new_req.Host = u.Host
	new_req.RequestURI = ""			//RequestURI需要设置为空，否则会报错
	client := &http.Client{}

	resp, err := client.Do(new_req)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return body,nil
}
