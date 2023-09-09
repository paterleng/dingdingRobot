package dingding

import (
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

type DingGroup struct {
	OpenConversationID string
	ChatID             string
	Name               string
	Token              DingToken
}

func (g *DingGroup) GetOpenConversationID() string {
	client, _err := createClient()
	if _err != nil {
		return g.OpenConversationID
	}

	chatIdToOpenConversationIdHeaders := &dingtalkim_1_0.ChatIdToOpenConversationIdHeaders{}
	chatIdToOpenConversationIdHeaders.XAcsDingtalkAccessToken = tea.String(g.Token.Token)
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		result, _err := client.ChatIdToOpenConversationIdWithOptions(tea.String(g.ChatID), chatIdToOpenConversationIdHeaders, &util.RuntimeOptions{})
		if _err != nil {
			return _err
		}
		g.OpenConversationID = *(result.Body.OpenConversationId)
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
	return g.OpenConversationID
}
