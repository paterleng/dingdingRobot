package dingding

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

/**
 * 使用 Token 初始化账号Client
 * @return Client
 * @throws Exception
 */
func createClient() (_result *dingtalkim_1_0.Client, _err error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	_result = &dingtalkim_1_0.Client{}
	_result, _err = dingtalkim_1_0.NewClient(config)
	return _result, _err
}

func GetOpenConverstaionId(access_token, chatId string) (openConverstaionId string, _err error) {
	client, _err := createClient()
	if _err != nil {
		return
	}

	chatIdToOpenConversationIdHeaders := &dingtalkim_1_0.ChatIdToOpenConversationIdHeaders{}
	chatIdToOpenConversationIdHeaders.XAcsDingtalkAccessToken = tea.String(access_token)
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		result, _err := client.ChatIdToOpenConversationIdWithOptions(tea.String(chatId), chatIdToOpenConversationIdHeaders, &util.RuntimeOptions{})
		if _err != nil {
			return _err
		}
		openConverstaionId = *(result.Body.OpenConversationId)
		return nil
	}()

	if tryErr != nil {
		var err = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			err = _t
		} else {
			err.Message = tea.String(tryErr.Error())
		}
		if !tea.BoolValue(util.Empty(err.Code)) && !tea.BoolValue(util.Empty(err.Message)) {
			// err 中含有 code 和 message 属性，可帮助开发定位问题
		}

	}
	return
}
