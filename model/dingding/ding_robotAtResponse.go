package dingding

type RobotAtResp struct {
	ConversationId            string                   `json:"conversationId"`            //会话ID
	AtUsers                   []map[string]interface{} `json:"atUsers"`                   //被@人的信息
	ChatbotCorpId             string                   `json:"catbotCorpId"`              //加密的机器人所在的企业corpId
	ChatbotUserId             string                   `json:"chatbotUserId"`             //加密的机器人ID
	MsgId                     string                   `json:"msgId"`                     //加密的消息ID
	SenderNick                string                   `json:"senderNick"`                //发送者昵称
	IsAdmin                   bool                     `json:"isAdmin"`                   //机器人发布上线后生效
	ConversationType          string                   `json:"conversationType"`          //1：单聊	,2：群聊
	SenderStaffId             string                   `json:"senderStaffId"`             //企业内部群中@该机器人的成员userid
	SessionWebhookExpiredTime int64                    `json:"sessionWebhookExpiredTime"` //当前会话的Webhook地址过期时间
	CreateAt                  int64                    `json:"createAt"`                  //消息的时间戳，单位毫秒
	SenderCorpId              string                   `json:"senderCorpId"`              //企业内部群有的发送者当前群的企业corpId
	SenderId                  string                   `json:"senderId"`                  //使用senderStaffId，作为发送者userid值。
	ConversationTitle         string                   `json:"conversationTitle"`         // 群聊时才有的会话标题
	IsInAtList                bool                     `json:"ssInAtList"`                // 是否在@列表中
	SessionWebhook            string                   `json:"sessionWebhook"`            // 当前会话的Webhook地址
	Text                      map[string]interface{}   `json:"text"`                      //里面content为内容
	Msgtype                   string                   `json:"msgtype"`                   //消息类型
}
