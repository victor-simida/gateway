package server
import (
	"gateway/web"
	"gateway/config"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"io/ioutil"
	"gateway/mylog"
	"crypto/tls"
	"io"
	"time"
	"net"
)

var httpClient *http.Client
var httpsClient *http.Client
func init() {
	httpClient = http.DefaultClient
	httpsClient = http.DefaultClient

	/*https默认不校验tls*/
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	httpsClient.Transport = tr

	for _, v := range config.Settings.HttpAddr {
		url := fmt.Sprintf("/gateway/%s", v.Suffix)
		mylog.LOG.I("Http retransfer %s", url)
		web.RegisterHandler(url, httpHandler)
	}

	for _, v := range config.Settings.HttpsAddr {
		url := fmt.Sprintf("/gateway/%s", v.Suffix)
		mylog.LOG.I("Https retransfer %s", url)
		web.RegisterHandler(url, httpsHandler)
	}

	for _, v := range config.Settings.HttpsBatchAddr {
		url := fmt.Sprintf("/gateway/%s**", v.Suffix)
		mylog.LOG.I("Https Batch retransfer %s", url)
		web.RegisterHandler(url, httpsBatchHandler)
	}

	for _, v := range config.Settings.HttpBatchAddr {
		url := fmt.Sprintf("/gateway/%s**", v.Suffix)
		mylog.LOG.I("Http Batch retransfer %s", url)
		web.RegisterHandler(url, httpBatchHandler)
	}
}

/*******************************************
*函数名：httpsHandler
*作用：转换地址并改为https请求
*时间：2016/1/19 20:08
*入参：
*返回值：
*******************************************/
func httpsHandler(w http.ResponseWriter, req *http.Request) {
	var urlStr string
	var host string
	for _, v := range config.Settings.HttpsAddr {
		if !strings.Contains(req.URL.Path, v.Suffix) {
			continue
		}
		suffix := strings.Split(req.RequestURI, "gateway/")[1]
		urlStr = fmt.Sprintf("https://%s/%s", v.Prefix, suffix)
		host = v.Host
		break
	}
	if urlStr == "" {
		mylog.LOG.E("urlStr is empty,req.URL.Path:%s", req.URL.Path)
		w.Write([]byte{})
		return
	}

	mylog.LOG.I("httpsHandler req:%+v,urlStr:%s", *req, urlStr)
	result, statusCode, err := TransToHttps(urlStr, req, host)
	if err != nil {
		result = []byte{}
		mylog.LOG.E("cloneRequest Error:%s", err.Error())
	}

	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
		return
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
func httpHandler(w http.ResponseWriter, req *http.Request) {
	var urlStr string
	var host string
	for _, v := range config.Settings.HttpAddr {
		if !strings.Contains(req.URL.Path, v.Suffix) {
			continue
		}
		suffix := strings.Split(req.RequestURI, "gateway/")[1]
		urlStr = fmt.Sprintf("http://%s/%s", v.Prefix, suffix)
		host = v.Host
		break
	}
	if urlStr == "" {
		mylog.LOG.E("urlStr is empty,req.URL.Path:%s", req.URL.Path)
		w.Write([]byte{})
		return
	}

	mylog.LOG.I("httpHandler req:%+v,urlStr:%s", *req, urlStr)
	result, statusCode, err := TransToHttp(urlStr, req, host)
	if err != nil {
		result = []byte{}
		mylog.LOG.E("cloneRequest Error:%s", err.Error())
	}

	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
		return
	}
	w.Write(result)

}

/*******************************************
*函数名：httpsBatchHandler
*作用：监听一系列路由，转换地址并改为https请求
*时间：2016/1/19 20:08
*入参：
*返回值：
*******************************************/
func httpsBatchHandler(w http.ResponseWriter, req *http.Request) {
	var urlStr string
	var host string
	for _, v := range config.Settings.HttpsBatchAddr {
		if !strings.Contains(req.URL.Path, v.Suffix) {
			continue
		}
		suffix := strings.Split(req.RequestURI, "gateway/")[1]
		urlStr = fmt.Sprintf("https://%s/%s", v.Prefix, suffix)
		host = v.Host
		mylog.LOG.I("httpsBatchHandler urlStr:%s", urlStr)
		break
	}
	if urlStr == "" {
		mylog.LOG.E("urlStr is empty,req.URL.Path:%s", req.URL.Path)
		w.Write([]byte{})
		return
	}

	mylog.LOG.I("httpsBatchHandler req:%+v,urlStr:%s", *req, urlStr)
	result, statusCode, err := TransToHttps(urlStr, req, host)
	if err != nil {
		result = []byte{}
		mylog.LOG.E("cloneRequest Error:%s", err.Error())
	}

	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
		return
	}
	w.Write(result)
}

/*******************************************
*函数名：httpBatchHandler
*作用：监听一系列路由，转换地址并改为http请求
*时间：2016/1/19 20:08
*入参：
*返回值：
*******************************************/
func httpBatchHandler(w http.ResponseWriter, req *http.Request) {
	var urlStr string
	var host string
	for _, v := range config.Settings.HttpBatchAddr {
		if !strings.Contains(req.URL.Path, v.Suffix) {
			continue
		}
		suffix := strings.Split(req.RequestURI, "gateway/")[1]
		host = v.Host
		urlStr = fmt.Sprintf("http://%s/%s", v.Prefix, suffix)
		mylog.LOG.I("httpBatchHandler urlStr:%s", urlStr)
		break
	}
	if urlStr == "" {
		mylog.LOG.E("urlStr is empty,req.URL.Path:%s", req.URL.Path)
		w.Write([]byte{})
		return
	}

	mylog.LOG.I("httpBatchHandler req:%+v,urlStr:%s", *req, urlStr)
	result, statusCode, err := TransToHttp(urlStr, req, host)
	if err != nil {
		result = []byte{}
		mylog.LOG.E("cloneRequest Error:%s", err.Error())
	}

	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
		return
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
func TransToHttps(urlStr string, req *http.Request, host string) ([]byte, int, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		mylog.LOG.E("urlStr parse error:%s", err.Error())
		return []byte{}, http.StatusNotFound, err
	}

	temp := *req

	new_req := &temp
	new_req.URL = u
	if host != "" {
		new_req.Host = host
	} else {
		new_req.Host = u.Host
	}
	new_req.RequestURI = ""            //RequestURI需要设置为空，否则会报错


	mylog.LOG.I("TransToHttps New Req:%+v", *new_req)
	resp, err := httpsClient.Do(new_req)
	if err != nil {
		mylog.LOG.E("client do error:%s", err.Error())
		return []byte{}, http.StatusNotFound, err
	}

	defer func (){
		io.Copy(ioutil.Discard,resp.Body)
		resp.Body.Close()
	} ()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mylog.LOG.E("ioutil.ReadAll error:%s", err.Error())
		return []byte{}, resp.StatusCode, err
	}
	mylog.LOG.I("Return response:%s %v", string(body), resp.StatusCode)
	return body, resp.StatusCode, nil
}

/*******************************************
*函数名：TransToHttp
*作用：克隆请求结构体，并改为发起新的http请求
*时间：2016/1/19 20:25
*入参：
*返回值：
*******************************************/
func TransToHttp(urlStr string, req *http.Request, host string) ([]byte, int, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		mylog.LOG.E("urlStr parse error:%s", err.Error())
		return []byte{}, http.StatusNotFound, err
	}

	temp := *req

	new_req := &temp
	new_req.URL = u
	if host != "" {
		new_req.Host = host
	} else {
		new_req.Host = u.Host
	}
	new_req.RequestURI = ""            //RequestURI需要设置为空，否则会报错

	mylog.LOG.I("TransToHttp New Req:%+v", new_req.Header)
	resp, err := httpClient.Do(new_req)
	if err != nil {
		mylog.LOG.E("client do error:%s", err.Error())
		return []byte{}, http.StatusNotFound, err
	}

	defer func (){
		io.Copy(ioutil.Discard,resp.Body)
		resp.Body.Close()
	} ()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mylog.LOG.E("ioutil.ReadAll error:%s", err.Error())
		return []byte{}, resp.StatusCode, err
	}
	mylog.LOG.I("Return response:%s %v", string(body), resp.StatusCode)
	return body, resp.StatusCode, nil
}

