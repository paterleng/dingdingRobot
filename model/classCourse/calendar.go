package classCourse

import (
	"crypto/tls"
	"ding/model/common/localTime"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

type Calendar struct {
	Week       int
	WeekNumber int
}

// 获取现在是第几周
func (*Calendar) GetWeek() (week int, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	curTime, _ := (&localTime.MySelfTime{}).GetCurTime(nil)
	//URL := "http://schedule.lzyjykj.com/sys/getWeek?nowday=" + curTime.Format[:10]
	URL := "http://localhost:20080/sys/getWeek?nowday=" + curTime.Format[:10]
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	request, err = http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return
	}
	resp, err = client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return
	}
	r := struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data int    `json:"data"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Code != 200 {
		return 0, errors.New(r.Msg)
	}
	// 此处举行具体的逻辑判断，然后返回即可
	return r.Data, nil
}
