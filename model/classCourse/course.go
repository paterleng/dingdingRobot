package classCourse

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ByClass []struct {
	Userid     string `json:"userid"`
	UName      string `json:"u_name"`
	ULocation  string `json:"u_location"`
	Group      string `json:"group"`
	UInstitute string `json:"u_institute"`
	UClass     string `json:"u_class"`
}

func GetIsHasCourse(lesson int, startWeek int, userType int, useridList []string, week int) (byClass ByClass, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	useridListString := ""
	for _, userid := range useridList {
		useridListString += userid + ","
	}
	useridListString = useridListString[:len(useridListString)-1]
	URL := fmt.Sprintf("http://8.130.137.7:20080/course/findUserByClass?lesson=%v&page=0&startWeek=%v&userType=%v&useridList=%v&week=%v&pageSize=100", lesson, startWeek, userType, useridListString, week)
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}

	//然后就可以放入具体的request中的
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
		Data struct {
			ByClass `json:"byClass"`
		} `json:"data"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Code != 200 {
		return nil, errors.New(r.Msg)
	}
	// 此处举行具体的逻辑判断，然后返回即可

	return r.Data.ByClass, nil
}
