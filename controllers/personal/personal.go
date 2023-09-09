package personal

import (
	"context"
	"crypto/tls"
	"ding/global"
	"ding/model"
	"ding/response"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

func Jk(c *gin.Context) {
	var p_book_name string
	c.ShouldBindJSON(&p_book_name)
	if p_book_name == "" {
		response.FailWithMessage("书名有误", c)
	}
	//应该是golang发送请求，拿到数据，反序列化到一个对象上面，然后绑定到数组上面。
	req := larkbitable.NewListAppTableRecordReqBuilder().
		AppToken("bascntcTTBjdMv1WUYsHYBkveRc").
		TableId("tblPuUs30HOrTWdi").
		Filter(fmt.Sprintf("CurrentValue.[当前阶段] = %s &&CurrentValue.[进展] != \"已完成\"", p_book_name)).Build()
	list, err := global.GLOBAL_Feishu.Bitable.AppTableRecord.List(context.Background(), req)
	if err != nil {

	}
	datas := list.Data.Items
	names := make([]string, len(datas))
	for i, data := range datas {
		name := data.Fields["姓名"]
		names[i] = name.(string)
	}

	//参数校验
	var p model.Jk
	p.Names = names
	//if err = c.ShouldBindJSON(&p); err != nil { //只能判断数据格式是不是json和数据类型
	//	zap.L().Error("Jk(京科的接口) with invalid param", zap.Error(err))
	//	errs, ok := err.(validator.ValidationErrors)
	//	if !ok {
	//		response.ResponseError(c, response.CodeInvalidParam)
	//		return
	//	}
	//	response.ResponseErrorWithMsg(c, response.CodeInvalidParam, RemoveTopStruct(errs.Translate(Trans)))
	//	return
	//}
	err = model.JkFunc(c, &p)
	if err != nil {
		fmt.Println(err)
	}

}
func Zjq(c *gin.Context) {
	//req := larksheets.NewFindSpreadsheetSheetReqBuilder().
	//	SpreadsheetToken("shtcnE5HudLyOfDdQIQUpi1Nm36").
	//	SheetId("770a26").
	//	Find(larksheets.NewFindBuilder().
	//		FindCondition(larksheets.NewFindConditionBuilder().Range("770a26!A1:M11").Build()).
	//		Find("").
	//		Build()).
	//	Build()
	//
	//resp, err := global.GLOBAL_Feishu.Sheets.SpreadsheetSheet.Find(context.Background(), req)
	//fmt.Println(resp.Data.FindResult)
	//if err != nil {
	//
	//}
	//使用golang封装一个请求
	type V struct {
		Values [][]string `json:"values"`
	}
	var client *http.Client   //封装客户端
	var request *http.Request //封装请求
	var resp *http.Response   //封装响应
	var body []byte
	urlForUserID := "https://open.feishu.cn/open-apis/sheets/v2/spreadsheets/shtcnE5HudLyOfDdQIQUpi1Nm36/values/770a26!A1:M11" //拼接URL
	client = &http.Client{Transport: &http.Transport{                                                                          //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	request, err := http.NewRequest(http.MethodPost, urlForUserID, nil)
	if err != nil {
		return
	}
	request.Header.Add("Content-Type", "application/json; charset=utf-8")
	request.Header.Add("Authorization", "Bearer t-g104aiagCFW7KH67N66S3PVILWTQ56WA5AGY3J6Y")
	resp, err = client.Do(request)
	if err != nil {
		zap.L().Error("请求发送错误", zap.Error(err))
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	Value := V{}
	err = json.Unmarshal(body, Value)
	if err != nil {

	}
}
func Lxy(c *gin.Context) {

}
