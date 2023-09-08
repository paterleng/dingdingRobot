package global

import lark "github.com/larksuite/oapi-sdk-go/v3"

func InitFeishu()  {
	var Feishu = lark.NewClient("cli_a3b0280db9f8d00e", "NM6tCipeAkvCWtmkVHhOFhmuPmu2yyPy") // 默认配置为自建应用
	GLOBAL_Feishu = Feishu
}

