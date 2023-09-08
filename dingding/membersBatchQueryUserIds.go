package dingding

type Body struct {
	UserId   string `json:"userid"`
	Language string `json:"language"`
}
type Response struct {
	Errcode int    `json:"errcode"`
	Result  Result `json:"result"`
}
type Result struct {
	Mobile string `json:"mobile"`
	Name   string `json:"name"`
}

//func GetUserIds(access_token, OpenConversationId string) (userIds []string, _err error) {
//	olduserIds := []*string{}
//	client, _err := createClient()
//	if _err != nil {
//		return
//	}
//
//	batchQueryGroupMemberHeaders := &dingtalkim_1_0.BatchQueryGroupMemberHeaders{}
//	batchQueryGroupMemberHeaders.XAcsDingtalkAccessToken = tea.String(access_token)
//	batchQueryGroupMemberRequest := &dingtalkim_1_0.BatchQueryGroupMemberRequest{
//		OpenConversationId: tea.String(OpenConversationId),
//		CoolAppCode:        tea.String("COOLAPP-1-102118DC0ABA212C89C7000H"),
//		MaxResults:         tea.Int64(300),
//		NextToken:          tea.String("XXXXX"),
//	}
//	tryErr := func() (_e error) {
//		defer func() {
//			if r := tea.Recover(recover()); r != nil {
//				_e = r
//			}
//		}()
//		result, _err := client.BatchQueryGroupMemberWithOptions(batchQueryGroupMemberRequest, batchQueryGroupMemberHeaders, &util.RuntimeOptions{})
//		if _err != nil {
//			return _err
//		}
//		olduserIds = result.Body.MemberUserIds
//		return
//	}()
//
//	if tryErr != nil {
//		var err = &tea.SDKError{}
//		if _t, ok := tryErr.(*tea.SDKError); ok {
//			err = _t
//		} else {
//			err.Message = tea.String(tryErr.Error())
//		}
//		if !tea.BoolValue(util.Empty(err.Code)) && !tea.BoolValue(util.Empty(err.Message)) {
//			// err 中含有 code 和 message 属性，可帮助开发定位问题
//		}
//
//	}
//	userIds = make([]string, len(olduserIds))
//	for i, id := range olduserIds {
//		userIds[i] = *id
//	}
//	return
//}
//func PostGetUserDetail(access_token string, UserId string) (tele dingding.Tele, err error) {
//	var client *http.Client
//	var request *http.Request
//	var resp *http.Response
//	var body []byte
//	URL := "https://oapi.dingtalk.com/topapi/v2/user/get?access_token=" + access_token
//	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
//		TLSClientConfig: &tls.Config{
//			InsecureSkipVerify: true,
//		},
//	}, Timeout: time.Duration(time.Second * 5)}
//	b := Body{
//		UserId: UserId,
//	}
//	bodymarshal, err := json.Marshal(&b)
//	if err != nil {
//		return
//	}
//	reqBody := strings.NewReader(string(bodymarshal))
//	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
//	if err != nil {
//		return
//	}
//	resp, err = client.Do(request)
//	if err != nil {
//		return
//	}
//	defer resp.Body.Close()
//	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
//	if err != nil {
//		return
//	}
//	r := Response{}
//	err = json.Unmarshal(body, &r)
//	if err != nil {
//		return
//	}
//	if r.Errcode == 33012 {
//		return dingding.Tele{}, errors.New("无效的userId,请检查userId是否正确")
//	} else if r.Errcode == 400002 {
//		return dingding.Tele{}, errors.New("无效的参数,请确认参数是否按要求输入")
//	} else if r.Errcode == -1 {
//		return dingding.Tele{}, errors.New("系统繁忙")
//	}
//	tele.DingUserID = UserId
//	tele.Number = r.Result.Mobile
//	tele.Personname = r.Result.Name
//	return
//}
