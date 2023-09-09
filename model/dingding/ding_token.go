package dingding

import (
	"context"
	"ding/global"
	"ding/utils"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dingtalkoauth2_1_0 "github.com/alibabacloud-go/dingtalk/oauth2_1_0"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"go.uber.org/zap"
	"time"
)

type DingToken struct {
	Token string `json:"token"`
}

func (t *DingToken) IsLegal() bool {
	//判断一个token是否合法
	return len(t.Token) == utils.TokenLength
}

// redis层下拿到access_token，先从redis中取出来，如果取不到的话，再重新申请一遍
func (t *DingToken) GetAccessToken() (access_token string, err error) {
	//var accessToken string
	var expire1 int64
	fmt.Println(expire1)
	expire, err := global.GLOBAL_REDIS.TTL(context.Background(), utils.AccessToken).Result()
	if err != nil {
		zap.L().Error("判断token剩余生存时间失败", zap.Error(err))
	}
	if expire == -1 || expire == -2 {
		//申请新的token
		access_token, expire1, err = t.GetAccessTokenDing()
		if err != nil {
			zap.L().Error("申请新的token失败", zap.Error(err))
			return
		}
		//将过期时间转换为int64

		//重新设置token和token的过期时间

		err = global.GLOBAL_REDIS.Set(context.Background(), utils.AccessToken, access_token, time.Second*7200).Err()
		if err != nil {
			zap.L().Error("重新设置token和token的过期时间失败", zap.Error(err))
			return
		}
		result, err := global.GLOBAL_REDIS.Get(context.Background(), utils.AccessToken).Result()
		if err != nil {
			zap.L().Error("重新申请后，获取token失败", zap.Error(err))
		}
		access_token = result
		if access_token == "" {
			zap.L().Error("重新申请后，获取token失败")
		}
	} else {
		access_token, err = global.GLOBAL_REDIS.Get(context.Background(), utils.AccessToken).Result()
	}
	//如果err是key不存在的话，应该重新申请一遍
	if err != nil {
		zap.L().Error("从redis从取access_token失败", zap.Error(err))
		return
	}

	return
}
func (t *DingToken) GxpGetAccessToken() (access_token string, err error) {
	//var accessToken string
	var expire1 int64
	fmt.Println(expire1)
	expire, err := global.GLOBAL_REDIS.TTL(context.Background(), utils.GxpAccessToken).Result()
	if err != nil {
		zap.L().Error("判断token剩余生存时间失败", zap.Error(err))
	}
	if expire == -2 {
		//申请新的token
		access_token, expire1, err = t.GxpGetAccessTokenDing()
		if err != nil {
			zap.L().Error("申请新的token失败", zap.Error(err))
			return
		}
		//将过期时间转换为int64

		//重新设置token和token的过期时间

		err = global.GLOBAL_REDIS.Set(context.Background(), utils.GxpAccessToken, access_token, time.Second*7200).Err()
		if err != nil {
			zap.L().Error("重新设置token和token的过期时间失败", zap.Error(err))
			return
		}
		result, err := global.GLOBAL_REDIS.Get(context.Background(), utils.GxpAccessToken).Result()
		if err != nil {
			zap.L().Error("重新申请后，获取token失败", zap.Error(err))
		}
		access_token = result
		if access_token == "" {
			zap.L().Error("重新申请后，获取token失败")
		}
	} else {
		access_token, err = global.GLOBAL_REDIS.Get(context.Background(), utils.GxpAccessToken).Result()
	}
	//如果err是key不存在的话，应该重新申请一遍
	if err != nil {
		zap.L().Error("从redis从取access_token失败", zap.Error(err))
		return
	}

	return
}
func (t *DingToken) GetAccessTokenDing() (access_token string, expireIn int64, _err error) {
	client, _err := CreateClient()
	if _err != nil {
		return
	}
	appKey, _ := global.GLOBAL_REDIS.Get(context.Background(), utils.AppKey).Result()
	appSecret, _ := global.GLOBAL_REDIS.Get(context.Background(), utils.AppSecret).Result()
	getAccessTokenRequest := &dingtalkoauth2_1_0.GetAccessTokenRequest{
		AppKey:    tea.String(appKey),
		AppSecret: tea.String(appSecret),
	}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		result, _err := client.GetAccessToken(getAccessTokenRequest)

		if _err != nil {
			return _err
		}
		access_token = *(result.Body.AccessToken) //把token拿出来
		expireIn = *(result.Body.ExpireIn)
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
func (t *DingToken) GxpGetAccessTokenDing() (access_token string, expireIn int64, _err error) {
	client, _err := CreateClient()
	if _err != nil {
		return
	}
	appKey, _ := global.GLOBAL_REDIS.Get(context.Background(), utils.GxpAppKey).Result()
	appSecret, _ := global.GLOBAL_REDIS.Get(context.Background(), utils.GxpAppSecret).Result()
	getAccessTokenRequest := &dingtalkoauth2_1_0.GetAccessTokenRequest{
		AppKey:    tea.String(appKey),
		AppSecret: tea.String(appSecret),
	}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		result, _err := client.GetAccessToken(getAccessTokenRequest)

		if _err != nil {
			return _err
		}
		access_token = *(result.Body.AccessToken) //把token拿出来
		expireIn = *(result.Body.ExpireIn)
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
func CreateClient() (_result *dingtalkoauth2_1_0.Client, _err error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	_result, _err = dingtalkoauth2_1_0.NewClient(config)
	return _result, _err
}
