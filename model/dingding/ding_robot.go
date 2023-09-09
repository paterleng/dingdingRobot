package dingding

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"ding/global"
	"ding/utils"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrorSpecInvalid = errors.New("定时规则可能不合法")
)

type DingRobot struct {
	RobotId            string         `gorm:"primaryKey;foreignKey:RobotId" json:"robot_id"` //机器人的token
	Deleted            gorm.DeletedAt `json:"deleted"`                                       //软删除字段
	Type               string         `json:"type"`                                          //机器人类型，1为企业内部机器人，2为自定义webhook机器人
	TypeDetail         string         `json:"type_detail"`                                   //具体机器人类型
	ChatBotUserId      string         `json:"chat_bot_user_id"`                              //加密的机器人id，该字段无用
	Secret             string         `json:"secret"`                                        //如果是自定义成机器人， 则存在此字段
	DingUserID         string         `json:"ding_user_id"`                                  // 机器人所属用户id
	UserName           string         `json:"user_name"`                                     //机器人所属用户名
	DingUsers          []DingUser     `json:"ding_users" gorm:"many2many:user_robot"`        //机器人@多个人，一个人可以被多个机器人@
	ChatId             string         `json:"chat_id"`                                       //机器人所在的群聊chatId
	OpenConversationID string         `json:"open_conversation_id"`                          //机器人所在的群聊openConversationID
	Tasks              []Task         `gorm:"foreignKey:RobotId;references:RobotId"`         //机器人拥有多个任务
	Name               string         `json:"name"`                                          //机器人的名称
	DingToken          `json:"ding_token" gorm:"-"`
	IsShared           int `json:"is_shared"`
}

func (r *DingRobot) GetSharedRobot() (Robots []DingRobot, err error) {
	err = global.GLOAB_DB.Where("is_shared = ?", 1).Find(&Robots).Error
	return
}
func (r *DingRobot) PingRobot() (err error) {
	robot, err := r.GetRobotByRobotId()
	if err != nil {
		zap.L().Error("通过robot_id获取robot失败", zap.Error(err))
		return
	}
	if robot.RobotId == "" {
		zap.L().Error("测试机器人发送消息失败，机器人id或者secret为空")
		return
	}
	p := &ParamCronTask{}
	p.MsgText.Msgtype = "text"
	p.MsgText.Text.Content = "测试"
	err = robot.SendMessage(p)
	if err != nil {
		zap.L().Error("测试机器人发送消息失败", zap.Error(err))
		return
	}
	return
}

type ResponseSendMessage struct {
	DingResponseCommon
}

func (r *DingRobot) AddDingRobot() (err error) {
	err = global.GLOAB_DB.Create(r).Error
	return
}
func (r *DingRobot) RemoveRobot() (err error) {
	err = global.GLOAB_DB.Delete(r).Error
	return
}
func (r *DingRobot) RemoveRobots(Robots []DingRobot) (err error) {
	err = global.GLOAB_DB.Delete(Robots).Error
	return
}
func (r *DingRobot) CreateOrUpdateRobot() (err error) {
	err = global.GLOAB_DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(r).Error
	return
}
func (r *DingRobot) GetRobotByRobotId() (robot *DingRobot, err error) {
	err = global.GLOAB_DB.Where("robot_id = ?", r.RobotId).First(&robot).Error
	return
}

type MySendParam struct {
	MsgParam  string   `json:"msgParam"`
	MsgKey    string   `json:"msgKey"`
	RobotCode string   `json:"robotCode"`
	UserIds   []string `json:"userIds"`
}

func (r *DingRobot) GxpSingleChat(p *ParamChat) (err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	var b MySendParam
	b.RobotCode = p.RobotCode
	b.MsgKey = p.MsgKey
	b.RobotCode = p.RobotCode
	b.UserIds = p.UserIds
	b.MsgParam = fmt.Sprintf("{       \"content\": \"%s\"   }", p.MsgParam)

	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return nil
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return nil
	}
	token, err := r.DingToken.GxpGetAccessToken()
	if err != nil {
		return err
	}
	request.Header.Set("x-acs-dingtalk-access-token", token)
	request.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(request)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return nil
	}
	h := struct {
		Code                      string   `json:"code"`
		Message                   string   `json:"message"`
		ProcessQueryKey           string   `json:"processQueryKey"`
		InvalidStaffIdList        []string `json:"invalidStaffIdList"`
		FlowControlledStaffIdList []string `json:"flowControlledStaffIdList"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &h)
	if err != nil {
		return nil
	}
	if h.Code != "" {
		return errors.New(h.Message)
	}
	// 此处举行具体的逻辑判断，然后返回即可

	return nil

}

// 钉钉机器人单聊
func (r *DingRobot) ChatSendMessage(p *ParamChat) error {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	var b MySendParam
	b.RobotCode = "dingepndjqy7etanalhi"
	if p.MsgKey == "sampleText" {
		b.MsgKey = p.MsgKey
		b.RobotCode = "dingepndjqy7etanalhi"
		b.UserIds = p.UserIds
		b.MsgParam = fmt.Sprintf("{       \"content\": \"%s\"   }", p.MsgParam)

	} else if strings.Contains(p.MsgKey, "sampleActionCard") {
		b.MsgKey = p.MsgKey
		b.RobotCode = "dingepndjqy7etanalhi"
		b.UserIds = p.UserIds
		b.MsgParam = p.MsgParam
	}

	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return nil
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return nil
	}
	token, err := r.DingToken.GetAccessToken()
	if err != nil {
		return err
	}
	request.Header.Set("x-acs-dingtalk-access-token", token)
	request.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(request)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return nil
	}
	h := struct {
		Code                      string   `json:"code"`
		Message                   string   `json:"message"`
		ProcessQueryKey           string   `json:"processQueryKey"`
		InvalidStaffIdList        []string `json:"invalidStaffIdList"`
		FlowControlledStaffIdList []string `json:"flowControlledStaffIdList"`
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &h)
	if err != nil {
		return nil
	}
	if h.Code != "" {
		return errors.New(h.Message)
	}
	// 此处举行具体的逻辑判断，然后返回即可

	return nil
}

type MySendGroupParam struct {
	MsgParam           string `json:"msgParam"`
	MsgKey             string `json:"msgKey"`
	RobotCode          string `json:"robotCode"`
	OpenConversationId string `json:"openConversationId"`
	CoolAppCode        string `json:"coolAppCode"`
}

func (r *DingRobot) ChatSendGroupMessage(p *ParamChat) (map[string]interface{}, error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	var res map[string]interface{}
	URL := "https://api.dingtalk.com/v1.0/robot/groupMessages/send"
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	var b MySendGroupParam
	b.RobotCode = "dingepndjqy7etanalhi"
	b.CoolAppCode = "COOLAPP-1-102118DC0ABA212C89C7000H"
	//b.OpenConversationId = "cidNOZESlAdvOGV/s3CVZxdlQ=="
	b.OpenConversationId = p.OpenConversationId
	if p.MsgKey == "sampleText" {
		b.MsgKey = p.MsgKey
		b.RobotCode = "dingepndjqy7etanalhi"
		b.MsgParam = fmt.Sprintf("{       \"content\": \"%s\"   }", p.MsgParam)
	} else if strings.Contains(p.MsgKey, "sampleActionCard") {
		b.MsgKey = p.MsgKey
		b.RobotCode = "dingepndjqy7etanalhi"
		b.MsgParam = p.MsgParam
	}

	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return res, nil
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
	if err != nil {
		return res, nil
	}
	token, err := r.DingToken.GetAccessToken()
	if err != nil {
		return res, err
	}
	request.Header.Set("x-acs-dingtalk-access-token", token)
	request.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(request)
	if err != nil {
		return res, nil
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body) //把请求到的body转化成byte[]
	if err != nil {
		return res, nil
	}

	//h := struct {
	//	Code                      string   `json:"code"`
	//	Message                   string   `json:"message"`
	//	ProcessQueryKey           string   `json:"processQueryKey"`
	//	InvalidStaffIdList        []string `json:"invalidStaffIdList"`
	//	FlowControlledStaffIdList []string `json:"flowControlledStaffIdList"`
	//}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &res)
	if err != nil {
		return res, nil
	}
	return res, nil
}
func (r *DingRobot) CronSend(c *gin.Context, p *ParamCronTask) (err error, task Task) {
	robotId := r.RobotId
	//转化时候出现问题
	spec, detailTimeForUser, err := HandleSpec(p)
	if p.Spec != "" {
		spec = p.Spec
	}
	tid := "0"
	UserId := ""
	if c != nil {
		UserId, err = global.GetCurrentUserId(c)
	}

	if err != nil {
		UserId = ""
	}
	CurrentUser, err := (&DingUser{UserId: UserId}).GetUserByUserId()
	if err != nil {
		CurrentUser = DingUser{}
	}
	r, err = (&DingRobot{RobotId: r.RobotId}).GetRobotByRobotId()
	if err != nil {
		zap.L().Error("通过机器人的robot_id获取机器人失败，是一个没有注册的机器人", zap.Error(err))
	}
	//a := "纪检部通知[广播][广播]:  \n断水断电不断简书，为谋其事必为总结[爱意]  \n[钉子]各期负责人于下周一晚上20：00前在钉钉简书小程序中标记优秀简书\n  \n[钉子]检查为机器人检查，大家要及时发表文章\n [灵感][灵感]注意：    \n  \n[对勾]简书严禁抄袭，坚持原创，我们会根据相关字段进行严格检查的哦[爱意]\n[对勾]纪检部同时也会对简书进行抽查[猫咪]\n[对勾]简书字数不能低于400字\n[钉子][钉子]重点！！！到周日20：00后财务部人员会在大群里面对简书未完成人员发起群收款[惊愕][惊愕]\n大家要注意了哦！！\n[灵感][灵感]提醒:   \n  \n [对勾]简书以及博客的时间为本周内，否则会被标记为未登记！\n  \n[爱意]希望大家多多参与简书投稿及评论互动[捧脸]并在此相互学习和借鉴哟[猫咪][猫咪]@所有人 "
	if errors.Is(err, gorm.ErrRecordNotFound) {
		r = &DingRobot{RobotId: robotId}
		err = nil
	}

	//到了这里就说明这个用户有这个小机器人
	//crontab := cron.New(cron.WithSeconds()) //精确到秒
	//spec := "* 30 22 * * ?" //cron表达式，每五秒一次
	if p.MsgText == nil && p.MsgLink == nil && p.MsgMarkDown == nil {
		p.MsgText.Msgtype = "text"
		p.RepeatTime = "立即发送"
	}

	if p.MsgText.Msgtype == "text" {
		if (p.RepeatTime) == "立即发送" { //这个判断说明我只想单纯的发送一条消息，不用做定时任务
			zap.L().Info("进入即时发送消息模式")
			err = r.SendMessage(p)
			if err != nil {
				return err, Task{}
			} else {
				zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", CurrentUser.Name, r.Name))
			}

			return err, task
		} else { //我要做定时任务
			tasker := func() {}
			zap.L().Info("进入定时任务模式")
			tasker = func() {
				err := r.SendMessage(p)
				if err != nil {
					return
				} else {
					zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", CurrentUser.Name, r.Name))
				}
			}
			TaskID, err := global.GLOAB_CORN.AddFunc(spec, tasker)
			tid = strconv.Itoa(int(TaskID))
			if err != nil {
				zap.L().Error("定时任务启动失败", zap.Error(err))
				err = ErrorSpecInvalid
				return err, Task{}
			}
			nextTime := global.GLOAB_CORN.Entry(TaskID).Next
			//把定时任务添加到数据库中
			task = Task{
				TaskID:            tid,
				TaskName:          p.TaskName,
				UserId:            CurrentUser.UserId,
				UserName:          CurrentUser.Name,
				RobotId:           r.RobotId,
				RobotName:         r.Name,
				Secret:            r.Secret,
				DetailTimeForUser: detailTimeForUser, //给用户看的
				Spec:              spec,              //cron后端定时规则
				FrontRepeatTime:   p.RepeatTime,      // 前端给的原始数据
				FrontDetailTime:   p.DetailTime,
				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
				NextTime:          nextTime,
			}
			err = (&task).InsertTask()
			if err != nil {
				zap.L().Info(fmt.Sprintf("定时任务插入数据库数据失败!用户名：%s,机器名 ： %s,定时规则：%s ,失败原因", CurrentUser.Name, r.Name, p.DetailTime, zap.Error(err)))
				return err, Task{}
			}
			zap.L().Info(fmt.Sprintf("定时任务插入数据库数据成功!用户名：%s,机器名 ： %s,定时规则：%s", CurrentUser.Name, r.Name, p.DetailTime))
		}
	} else if p.MsgLink.Msgtype == "link" {
		if (p.RepeatTime) == "立即发送" { //这个判断说明我只想单纯的发送一条消息，不用做定时任务
			zap.L().Info("进入即时发送消息模式")
			err := r.SendMessage(p)
			if err != nil {
				return err, Task{}
			} else {
				zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", CurrentUser.Name, r.Name))
			}
			//定时任务
			task = Task{
				TaskID:            tid,
				TaskName:          p.TaskName,
				UserId:            CurrentUser.UserId,
				UserName:          CurrentUser.Name,
				RobotId:           r.RobotId,
				RobotName:         r.Name,
				Secret:            r.Secret,
				DetailTimeForUser: detailTimeForUser, //给用户看的
				Spec:              spec,              //cron后端定时规则
				FrontRepeatTime:   p.RepeatTime,      // 前端给的原始数据
				FrontDetailTime:   p.DetailTime,
				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
				//MsgLink:           p.MsgLink,
				//MsgMarkDown:       p.MsgMarkDown,
			}
			return err, task
		} else { //我要做定时任务
			tasker := func() {}
			zap.L().Info("进入定时任务模式")
			tasker = func() {
				err := r.SendMessage(p)
				if err != nil {
					return
				} else {
					zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", CurrentUser.Name, r.Name))
				}
			}
			TaskID, err := global.GLOAB_CORN.AddFunc(spec, tasker)
			nextTime := global.GLOAB_CORN.Entry(TaskID).Next

			tid = strconv.Itoa(int(TaskID))
			if err != nil {
				err = ErrorSpecInvalid
				return err, Task{}
			}
			//把定时任务添加到数据库中
			task = Task{
				TaskID:            tid,
				TaskName:          p.TaskName,
				UserId:            CurrentUser.UserId,
				UserName:          CurrentUser.Name,
				RobotId:           r.RobotId,
				RobotName:         r.Name,
				Secret:            r.Secret,
				DetailTimeForUser: detailTimeForUser, //给用户看的
				Spec:              spec,              //cron后端定时规则
				FrontRepeatTime:   p.RepeatTime,      // 前端给的原始数据
				FrontDetailTime:   p.DetailTime,
				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
				//MsgLink:           p.MsgLink,
				//MsgMarkDown:       p.MsgMarkDown,
				NextTime: nextTime,
			}
			err = (&task).InsertTask()
			if err != nil {
				zap.L().Info(fmt.Sprintf("定时任务插入数据库数据失败!用户名：%s,机器名 ： %s,定时规则：%s ,失败原因", CurrentUser.Name, r.Name, p.DetailTime, zap.Error(err)))
				return err, Task{}
			}
			zap.L().Info(fmt.Sprintf("定时任务插入数据库数据成功!用户名：%s,机器名 ： %s,定时规则：%s", CurrentUser.Name, r.Name, p.DetailTime))
		}
	} else if p.MsgMarkDown.Msgtype == "markdown" {
		if err != nil {
			zap.L().Error("通过人名查询电话号码失败", zap.Error(err))
			return
		}
		if (p.RepeatTime) == "立即发送" { //这个判断说明我只想单纯的发送一条消息，不用做定时任务
			zap.L().Info("进入即时发送消息模式")
			err := r.SendMessage(p)
			if err != nil {
				return err, Task{}
			} else {
				zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", CurrentUser.Name, r.Name))
			}
			//定时任务
			task = Task{
				TaskID:            tid,
				TaskName:          p.TaskName,
				UserId:            CurrentUser.UserId,
				UserName:          CurrentUser.Name,
				RobotId:           r.RobotId,
				RobotName:         r.Name,
				Secret:            r.Secret,
				DetailTimeForUser: detailTimeForUser, //给用户看的
				Spec:              spec,              //cron后端定时规则
				FrontRepeatTime:   p.RepeatTime,      // 前端给的原始数据
				FrontDetailTime:   p.DetailTime,
				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
				//MsgLink:           p.MsgLink,
				//MsgMarkDown:       p.MsgMarkDown,
			}
			return err, task
		} else { //我要做定时任务
			tasker := func() {}
			zap.L().Info("进入定时任务模式")
			tasker = func() {
				err := r.SendMessage(p)
				if err != nil {
					return
				} else {
					zap.L().Info(fmt.Sprintf("发送消息成功！发送人:%s,对应机器人:%s", CurrentUser.Name, r.Name))
				}
			}
			TaskID, err := global.GLOAB_CORN.AddFunc(spec, tasker)
			tid = strconv.Itoa(int(TaskID))
			if err != nil {
				err = ErrorSpecInvalid
				return err, Task{}
			}
			//把定时任务添加到数据库中
			task = Task{
				TaskID:            tid,
				TaskName:          p.TaskName,
				UserId:            CurrentUser.UserId,
				UserName:          CurrentUser.Name,
				RobotId:           r.RobotId,
				RobotName:         r.Name,
				Secret:            r.Secret,
				DetailTimeForUser: detailTimeForUser, //给用户看的
				Spec:              spec,              //cron后端定时规则
				FrontRepeatTime:   p.RepeatTime,      // 前端给的原始数据
				FrontDetailTime:   p.DetailTime,
				MsgText:           p.MsgText, //到时候此处只会存储一个MsgText的id字段
				//MsgLink:           p.MsgLink,
				//MsgMarkDown:       p.MsgMarkDown,
			}
			err = (&task).InsertTask()
			if err != nil {
				zap.L().Info(fmt.Sprintf("定时任务插入数据库数据失败!用户名：%s,机器名 ： %s,定时规则：%s ,失败原因", CurrentUser.Name, r.Name, p.DetailTime, zap.Error(err)))
				return err, Task{}
			}
			zap.L().Info(fmt.Sprintf("定时任务插入数据库数据成功!用户名：%s,机器名 ： %s,定时规则：%s", CurrentUser.Name, r.Name, p.DetailTime))
		}
	}

	global.GLOAB_CORN.Start()

	return err, task

}

// SendMessage Function to send message
//
//goland:noinspection GoUnhandledErrorResult
func (t *DingRobot) SendMessage(p *ParamCronTask) error {
	b := []byte{}
	//我们需要在文本，链接，markdown三种其中的一个
	if p.MsgText.Msgtype == "text" {
		msg := map[string]interface{}{}
		atMobileStringArr := make([]string, len(p.MsgText.At.AtMobiles))
		for i, atMobile := range p.MsgText.At.AtMobiles {
			atMobileStringArr[i] = atMobile.AtMobile
		}
		atUserIdStringArr := make([]string, len(p.MsgText.At.AtUserIds))
		for i, AtuserId := range p.MsgText.At.AtUserIds {
			atUserIdStringArr[i] = AtuserId.AtUserId
		}
		msg = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": p.MsgText.Text.Content,
			},
		}
		if p.MsgText.At.IsAtAll {
			msg["at"] = map[string]interface{}{
				"isAtAll": p.MsgText.At.IsAtAll,
			}
		} else {
			msg["at"] = map[string]interface{}{
				"atMobiles": atMobileStringArr, //字符串切片类型
				"atUserIds": atUserIdStringArr,
				"isAtAll":   p.MsgText.At.IsAtAll,
			}
		}
		b, _ = json.Marshal(msg)

	} else if p.MsgLink.Msgtype == "link" {
		//直接序列化
		b, _ = json.Marshal(p.MsgLink)
	} else if p.MsgMarkDown.Msgtype == "markdown" {
		msg := map[string]interface{}{}
		atMobileStringArr := make([]string, len(p.MsgMarkDown.At.AtMobiles))
		for i, atMobile := range p.MsgMarkDown.At.AtMobiles {
			atMobileStringArr[i] = atMobile.AtMobile
		}
		msg = map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": p.MsgMarkDown.MarkDown.Title,
				"text":  p.MsgMarkDown.MarkDown.Text,
			},
		}
		if p.MsgText.At.IsAtAll {
			msg["at"] = map[string]interface{}{
				"isAtAll": p.MsgText.At.IsAtAll,
			}
		} else {
			msg["at"] = map[string]interface{}{
				"atMobiles": atMobileStringArr, //字符串切片类型
				"isAtAll":   p.MsgText.At.IsAtAll,
			}
		}
		b, _ = json.Marshal(msg)
	} else {
		msg := map[string]interface{}{}
		atMobileStringArr := make([]string, len(p.MsgText.At.AtMobiles))
		for i, atMobile := range p.MsgText.At.AtMobiles {
			atMobileStringArr[i] = atMobile.AtMobile
		}
		atUserIdStringArr := make([]string, len(p.MsgText.At.AtUserIds))
		for i, AtuserId := range p.MsgText.At.AtUserIds {
			atUserIdStringArr[i] = AtuserId.AtUserId
		}
		msg = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": p.MsgText.Text.Content,
			},
		}
		if p.MsgText.At.IsAtAll {
			msg["at"] = map[string]interface{}{
				"isAtAll": p.MsgText.At.IsAtAll,
			}
		} else {
			msg["at"] = map[string]interface{}{
				"atMobiles": atMobileStringArr, //字符串切片类型
				"atUserIds": atUserIdStringArr,
				"isAtAll":   p.MsgText.At.IsAtAll,
			}
		}
		b, _ = json.Marshal(msg)
	}

	var resp *http.Response
	var err error
	if t.Type == "1" || t.Secret == "" {
		resp, err = http.Post(t.getURLV2(), "application/json", bytes.NewBuffer(b))
	} else {
		resp, err = http.Post(t.getURL(), "application/json", bytes.NewBuffer(b))
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	date, err := ioutil.ReadAll(resp.Body)
	r := ResponseSendMessage{}
	err = json.Unmarshal(date, &r)
	if err != nil {
		return err
	}
	if r.Errcode != 0 {
		fmt.Println(r.Errmsg)
		return errors.New(r.Errmsg)
	}

	return nil
}

func (t *DingRobot) hmacSha256(stringToSign string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (t *DingRobot) getURL() string {
	url := "https://oapi.dingtalk.com/robot/send?access_token=" + t.RobotId //拼接token路径
	timestamp := time.Now().UnixNano() / 1e6                                //以毫秒为单位
	formatTimeStr := time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
	zap.L().Info(fmt.Sprintf("当时时间戳对应的时间是：%s", formatTimeStr))
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, t.Secret)
	sign := t.hmacSha256(stringToSign, t.Secret)
	url = fmt.Sprintf("%s&timestamp=%d&sign=%s", url, timestamp, sign) //把timestamp和sign也拼接在一起
	return url
}
func (t *DingRobot) getURLV2() string {
	url := "https://oapi.dingtalk.com/robot/send?access_token=" + t.RobotId //拼接token路径
	return url
}

//	func (t *DingRobot) StopTask(id string) (err error) {
//		task := Task{
//			TaskID: id,
//		}
//		taskID, err := mysql.StopTask(task)
//
//		if errors.Is(err, mysql.ErrorNotHasTask) {
//			return mysql.ErrorNotHasTask
//		}
//		global.GLOAB_CORN.Remove(cron.EntryID(taskID))
//		return err
//	}
func (*DingRobot) SendSessionWebHook(p *ParamReveiver) (err error) {
	var msg map[string]interface{}
	//如果@机器人的消息包含考勤，且包含三期或者四期，再加上时间限制
	robot := &DingRobot{}
	if strings.Contains(p.Text.Content, "打字邀请码") {
		code, _, err := robot.GetInviteCode()
		if err != nil {
			zap.L().Error("申请新的TypingInviationCode失败", zap.Error(err))
			return err
		}
		msg = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": utils.TypingInviationSucc + ": " + code,
			},
		}
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	var resp *http.Response

	resp, err = http.Post(p.SessionWebhook, "application/json", bytes.NewBuffer(b))

	defer resp.Body.Close()
	date, err := ioutil.ReadAll(resp.Body)
	fmt.Println(date)
	if err != nil {
		return err
	}
	return nil
}

func (*DingRobot) GxpSendSessionWebHook(p *ParamReveiver) (err error) {

	//var msg map[string]interface{}
	getTime, _ := time.Parse("15:04", time.Now().Format("15:04"))
	startHour := ""
	startMin := ""
	if 0 <= utils.StartHour && utils.StartHour < 10 {
		startHour = "0" + strconv.Itoa(utils.StartHour)
	} else {
		startHour = strconv.Itoa(utils.StartHour)
	}
	if 0 <= utils.StartMin && utils.StartMin < 10 {
		startMin = "0" + strconv.Itoa(utils.StartMin)
	} else {
		startMin = strconv.Itoa(utils.StartMin)
	}
	startTime := startHour + ":" + startMin
	startTimer, _ := time.Parse("15:04", startTime)
	endHour := ""
	endMin := ""
	if 0 <= utils.EndHour && utils.EndHour < 10 {
		endHour = "0" + strconv.Itoa(utils.EndHour)
	} else {
		endHour = strconv.Itoa(utils.EndHour)
	}
	if 0 <= utils.EndMin && utils.EndMin < 10 {
		endMin = "0" + strconv.Itoa(utils.EndMin)
	} else {
		endMin = strconv.Itoa(utils.EndMin)
	}
	endTime := endHour + ":" + endMin
	endTimer, _ := time.Parse("15:04", endTime)
	var msg map[string]interface{}
	if getTime.Before(startTimer) {
		msg = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": fmt.Sprintf("未到报备时间，请%s后重新报备", startTimer.Format("15:04")),
			},
		}
	} else if getTime.After(endTimer) {
		return
	} else {
		//如果@机器人的消息包含考勤，且包含三期或者四期，再加上时间限制
		//if strings.Contains(p.Text.Content, "到") || strings.Contains(p.Text.Content, "宿舍") || strings.Contains(p.Text.Content, "寝室") {
		//
		//} else {
		//	r := Record{TongXinUserID: p.SenderStaffId, IsAtRobot: true, IsInRoom: false, Content: p.Text.Content}
		//	err = global.GLOAB_DB1.Where("id = ?", p.SenderStaffId).Create(&r).Error
		//	if err != nil {
		//		zap.L().Error("发送其他信息，存入数据库失败", zap.Error(err))
		//	}
		//}
		r := Record{TongXinUserID: p.SenderStaffId, IsAtRobot: true, IsInRoom: true, Content: p.Text.Content}
		err = global.GLOAB_DB1.Where("id = ?", p.SenderStaffId).Create(&r).Error
		if err != nil {
			zap.L().Error("发送到宿舍后，存入数据库失败", zap.Error(err))
		}
		msg = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": "收到,学习辛苦了[送花花]",
			},
		}
		msg["at"] = map[string][]string{
			"atUserIds": []string{p.SenderStaffId},
		}
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	var resp *http.Response

	resp, err = http.Post(p.SessionWebhook, "application/json", bytes.NewBuffer(b))

	defer resp.Body.Close()
	date, err := ioutil.ReadAll(resp.Body)
	fmt.Println(date)
	if err != nil {
		return err
	}
	return nil
}
func TypingInviation() (TypingInvitationCode string, expire time.Duration, err error) {
	zap.L().Info("进入到了chromedp，开始申请")
	timeCtx, cancel := context.WithTimeout(GetChromeCtx(false), 5*time.Minute)
	defer cancel()
	//opts := append(
	//	chromedp.DefaultExecAllocatorOptions[:],
	//	chromedp.NoDefaultBrowserCheck,                        //不检查默认浏览器
	//	chromedp.Flag("headless", false),                      // 禁用chrome headless（禁用无窗口模式，那就是开启窗口模式）
	//	chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
	//	chromedp.Flag("ignore-certificate-errors", true),      //忽略错误
	//	chromedp.Flag("disable-web-security", true),           //禁用网络安全标志
	//	chromedp.Flag("disable-extensions", true),             //开启插件支持
	//	chromedp.Flag("disable-default-apps", true),
	//	chromedp.NoFirstRun, //设置网站不是首次运行
	//	chromedp.WindowSize(1921, 1024),
	//	chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"), //设置UserAgent
	//)
	//
	//allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	//defer cancel()
	//
	////创建上下文实例
	//timeCtx, cancel := chromedp.NewContext(
	//	allocCtx,
	//	chromedp.WithLogf(log.Printf),
	//)
	//defer cancel()
	// 创建超时上下文
	var html string
	//ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)

	err = chromedp.Run(timeCtx,
		chromedp.Navigate("https://dazi.kukuw.com/"),
		//点击“我的打字“按钮
		chromedp.Click(`document.getElementById("globallink").getElementsByTagName("a")[5]`, chromedp.ByJSPath),
		// 锁定用户名框并填写内容
		chromedp.WaitVisible(`document.querySelector("#name")`, chromedp.ByJSPath),
		chromedp.SetValue(`document.querySelector("#name")`, "闫佳鹏", chromedp.ByJSPath),
		//锁定密码框并填写内容
		chromedp.WaitVisible(`document.querySelector("#pass")`, chromedp.ByJSPath),
		chromedp.SetValue(`document.querySelector("#pass")`, "123456", chromedp.ByJSPath),
		//点击登录按钮
		chromedp.WaitVisible(`document.querySelector(".button").firstElementChild`, chromedp.ByJSPath),
		chromedp.Click(`document.querySelector(".button").firstElementChild`, chromedp.ByJSPath),
		//点击发布竞赛
		chromedp.WaitVisible(`document.querySelector("a.groupnew")`, chromedp.ByJSPath),
		chromedp.Click(`document.querySelector("a.groupnew")`, chromedp.ByJSPath),
		chromedp.Sleep(time.Second),
		////点击所要打字的文章
		chromedp.WaitVisible(`document.querySelector("a#select_b.select_b")`, chromedp.ByJSPath),
		chromedp.Click(`document.querySelector("a#select_b.select_b")`, chromedp.ByJSPath),
		chromedp.WaitVisible(`document.querySelector("a.sys.on")`, chromedp.ByJSPath),
		chromedp.Click(`document.querySelector("a.sys.on")`, chromedp.ByJSPath),
		//设置比赛时间2分钟
		chromedp.Evaluate(`document.querySelector("#set_time").value=10`, nil),
		//选择有效期
		chromedp.Evaluate("document.querySelector(\"select#youxiaoqi\").value = document.querySelector(\"#youxiaoqi > option:nth-child(5)\").value", nil),
		//设置成为不公开
		chromedp.Click(`document.querySelectorAll("input#gongkai")[1]`, chromedp.ByJSPath),
		//点击发布按钮
		chromedp.Click(`document.querySelectorAll(".artnew table tr td input")[7]`, chromedp.ByJSPath),
		chromedp.WaitVisible(`document.querySelectorAll("#my_main .art_table td")[9].childNodes[0]`, chromedp.ByJSPath),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("打字吗出现了")
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("爬取前:", TypingInvitationCode)
			a := chromedp.OuterHTML(`document.querySelector("body")`, &html, chromedp.ByJSPath)
			err := a.Do(ctx)
			if err != nil {
				zap.L().Error("chromedp获取页面全部数据失败", zap.Error(err))
				return err
			}
			dom, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				zap.L().Error("chromedp获取页面全部数据后，转化成dom失败", zap.Error(err))
				return err
			}
			//dom.Find(`#my_main > .art_table > tbody > tr:nth-child(2) > td:nth-child(4) > span`).Each(func(i int, selection *goquery.Selection) {
			//	TypingInvitationCode = TypingInvitationCode + selection.Text()
			//	fmt.Println("爬取后:",TypingInvitationCode)
			//	selection.Next()
			//})
			TypingInvitationCode = dom.Find(`#my_main > .art_table > tbody > tr:nth-child(2) > td:nth-child(4) > span`).First().Text()
			if TypingInvitationCode == "" {
				zap.L().Error("爬取打字邀请码失败")
				return err
			}
			_, err = global.GLOBAL_REDIS.Set(context.Background(), utils.ConstTypingInvitationCode, TypingInvitationCode, time.Second*60*60*5).Result() //5小时过期时间
			if err != nil {
				zap.L().Error("爬取打字邀请码后存入redis失败", zap.Error(err))
			}
			return err
		}),
	)

	if err != nil {
		zap.L().Error("chromedp.Run有误", zap.Error(err))
		return "", time.Second * 0, err
	} else {
		zap.L().Info(fmt.Sprintf("chromedp.Run无误，成功获取打字邀请码:%v", TypingInvitationCode), zap.Error(err))
		return TypingInvitationCode, time.Second * 60 * 60 * 5, err
	}

}
func (d *DingRobot) GetInviteCode() (code string, expire time.Duration, err error) {
	//如果@机器人的消息包含考勤，且包含三期或者四期，再加上时间限制
	//去redis中取一下打字邀请码
	var TypingInviationCode string
	var expire1 int64
	fmt.Println(expire1)
	expire, err = global.GLOBAL_REDIS.TTL(context.Background(), utils.ConstTypingInvitationCode).Result()
	if err != nil {
		zap.L().Error("判断token剩余生存时间失败", zap.Error(err))
	}
	//如果redis里面没有的话
	if expire == -2 {
		zap.L().Error("redis中无打字码，去申请", zap.Error(err))
		//申请新的TypingInviationCode并已经存入redis
		TypingInviationCode, expire, err = TypingInviation()
		if err != nil || TypingInviationCode == "" {
			zap.L().Error("申请新的TypingInviationCode失败", zap.Error(err))
			return TypingInviationCode, time.Second * 0, err
		}

	} else {
		//从redis从取到邀请码
		TypingInviationCode = global.GLOBAL_REDIS.Get(context.Background(), utils.ConstTypingInvitationCode).Val()
		if len(TypingInviationCode) != 5 {
			zap.L().Error("申请新的TypingInviationCode失败", zap.Error(err))
			return TypingInviationCode, expire, errors.New("申请新的TypingInviationCode失败")
		}
	}
	return TypingInviationCode, expire, nil

}
func HandleSpec(p *ParamCronTask) (spec, detailTimeForUser string, err error) {
	spec = ""
	detailTimeForUser = ""
	n := len(p.DetailTime)
	if p.RepeatTime == "仅发送一次" {
		second := p.DetailTime[n-2:]
		minute := p.DetailTime[n-5 : n-3]
		hour := p.DetailTime[n-8 : n-6]
		//year := p.DetailTime[:4]
		month := p.DetailTime[5:7]
		day := p.DetailTime[8:10]
		week := "?" //问号代表放弃周
		spec = second + " " + minute + " " + hour + " " + day + " " + month + " " + week
		detailTimeForUser = "仅在" + p.DetailTime + "发送一次"
	}
	if string([]rune(p.RepeatTime)[0:3]) == "周重复" {
		M := map[string]string{"0": "周日", "1": "周一", "2": "周二", "3": "周三", "4": "周四", "5": "周五", "6": "周六"}
		detailTimeForUser = "周重复 ："
		weeks := strings.Split(p.RepeatTime, "/")[1:]
		week := ""
		for i := 0; i < len(weeks); i++ {
			detailTimeForUser += M[weeks[i]]
			week += weeks[i] + ","
		}
		week = week[0 : len(week)-1]
		HMS := strings.Split(p.DetailTime, ":")
		second := HMS[2]
		minute := HMS[1]
		hour := HMS[0]
		month := "*" //每个月的每个星期都发送
		day := "?"   //选了星期就要放弃具体的某一天
		detailTimeForUser += hour + "：" + minute + "：" + second
		spec = second + " " + minute + " " + hour + " " + day + " " + month + " " + week
	}

	if string([]rune(p.RepeatTime)[0:3]) == "月重复" {
		var daymap map[int]string
		daymap = make(map[int]string)
		for i := 1; i <= 31; i++ {
			daymap[i] += strconv.Itoa(i) + "号"
		}
		//字符串数组
		days := strings.Split(p.RepeatTime, "/")[1:]
		detailTimeForUser = "月重复 ："
		day := ""
		for i := 0; i < len(days); i++ {
			atoi, _ := strconv.Atoi(days[i])
			detailTimeForUser += daymap[atoi]
			day += days[i] + ","
		}
		day = day[0 : len(day)-1]
		HMS := strings.Split(p.DetailTime, ":")
		second := HMS[2]
		minute := HMS[1]
		hour := HMS[0]
		month := "*" //每个月的每个星期都发送
		week := "?"
		detailTimeForUser += hour + ":" + minute + ":" + second
		spec = second + " " + minute + " " + hour + " " + day + " " + month + " " + week
	}

	if spec == "" || detailTimeForUser == "" {
		return spec, detailTimeForUser, errors.New("cron定时规则转化错误")
	}
	return spec, detailTimeForUser, nil
}

// 获取机器人所在的群聊的userIdList ，前提是获取到OpenConversationId，获取到OpenConverstaionId的前提是获取到二维码
func (r *DingRobot) GetGroupUserIds() (userIds []string, _err error) {
	//所需参数access_token, OpenConversationId string
	olduserIds := []*string{}
	client, _err := createClient()
	if _err != nil {
		return
	}

	batchQueryGroupMemberHeaders := &dingtalkim_1_0.BatchQueryGroupMemberHeaders{}
	batchQueryGroupMemberHeaders.XAcsDingtalkAccessToken = tea.String(r.DingToken.Token)
	batchQueryGroupMemberRequest := &dingtalkim_1_0.BatchQueryGroupMemberRequest{
		OpenConversationId: tea.String(r.OpenConversationID),
		CoolAppCode:        tea.String("COOLAPP-1-102118DC0ABA212C89C7000H"),
		MaxResults:         tea.Int64(300),
		NextToken:          tea.String("XXXXX"),
	}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		result, _err := client.BatchQueryGroupMemberWithOptions(batchQueryGroupMemberRequest, batchQueryGroupMemberHeaders, &util.RuntimeOptions{})
		if _err != nil {
			return _err
		}
		olduserIds = result.Body.MemberUserIds
		return
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
	userIds = make([]string, len(olduserIds))
	for i, id := range olduserIds {
		userIds[i] = *id
	}
	return
}

func createClient() (_result *dingtalkim_1_0.Client, _err error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	_result, _err = dingtalkim_1_0.NewClient(config)
	return _result, _err
}

func GetImage(c *gin.Context) { //显示图片的方法
	imageName := c.Query("imageName")     //截取get请求参数，也就是图片的路径，可是使用绝对路径，也可使用相对路径
	file, _ := ioutil.ReadFile(imageName) //把要显示的图片读取到变量中
	c.Writer.WriteString(string(file))    //关键一步，写给前端
}
func (t *DingRobot) StopTask(taskId string) (err error) {
	//先来判断一下是否拥有这个定时任务
	var task Task
	err = global.GLOAB_DB.Where("task_id", taskId).First(&task).Error
	if err != nil {
		zap.L().Info("通过taskId查找定时任务失败", zap.Error(err))
		return err
	}
	taskID, err := strconv.Atoi(task.TaskID)
	if err != nil {
		return err
	}
	//到了这里就说明我有这个定时任务，我要移除这个定时任务
	err = global.GLOAB_DB.Delete(&task).Error
	if err != nil {
		zap.L().Error("删除定时任务失败", zap.Error(err))
		return err
	}
	//global.GLOAB_CORN.
	global.GLOAB_CORN.Remove(cron.EntryID(taskID))
	return err
}
func (t *DingRobot) GetTaskList(RobotId string) (tasks []Task, err error) {
	err = global.GLOAB_DB.Model(&DingRobot{RobotId: RobotId}).Unscoped().Association("Tasks").Find(&tasks) //通过机器人的id拿到机器人，拿到机器人后，我们就可以拿到所有的任务
	if err != nil {
		zap.L().Error("通过机器人robot_id拿到该机器人的所有定时任务失败", zap.Error(err))
		return
	}
	return
}
func (t *DingRobot) RemoveTask(taskId string) (err error) {
	//先来判断一下是否拥有这个定时任务
	var task Task
	err = global.GLOAB_DB.Unscoped().Where("id = ?", taskId).First(&task).Error
	if err != nil {
		zap.L().Info("通过taskId查找定时任务失败", zap.Error(err))
		return err
	}
	taskID, err := strconv.Atoi(task.TaskID)
	if err != nil {
		return err
	}
	//到了这里就说明我有这个定时任务，我要移除这个定时任务
	err = global.GLOAB_DB.Unscoped().Delete(&task).Error
	if err != nil {
		zap.L().Error("删除定时任务失败", zap.Error(err))
		return err
	}
	global.GLOAB_CORN.Remove(cron.EntryID(taskID))
	return err
}
func (t *DingRobot) GetUnscopedTaskByID(id string) (task Task, err error) {
	err = global.GLOAB_DB.Unscoped().Preload("MsgText.At.AtMobiles").Preload("MsgText.At.AtUserIds").Preload("MsgText.Text").First(&task, id).Error
	if err != nil {
		zap.L().Error("通过主键id查询定时任务失败", zap.Error(err))
		return
	}
	return
}
func (t *DingRobot) ReStartTask(id string) (task Task, err error) {
	err = global.GLOAB_DB.Model(&Task{}).First(&task, id).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Task{}, errors.New("该定时任务没有暂停，所以无法重启")
	}
	task, err = t.GetUnscopedTaskByID(id)
	//根据这个id主键查询到被删除的数据
	err = global.GLOAB_DB.Unscoped().Model(&task).Update("deleted_at", nil).Error //这个地方必须加上Unscoped()，否则不报错，但是却无法更新
	p := ParamCronTask{
		MsgText:     task.MsgText,
		MsgLink:     task.MsgLink,
		MsgMarkDown: task.MsgMarkDown,
		RobotId:     task.RobotId,
	}
	d := DingRobot{
		RobotId: task.RobotId,
		Secret:  task.Secret,
	}
	tasker := func() {
		err := d.SendMessage(&p)
		if err != nil {
			//zap.L().Error(fmt.Sprintf("恢复任务失败！发送人:%s,对应机器人:%s", username, robotname), zap.Error(err))
			return
		} else {
			//zap.L().Info(fmt.Sprintf("恢复任务成功！发送人:%s,对应机器人:%s", username, robotname))
		}
	}
	//	// 添加定时任务
	TaskID, err := global.GLOAB_CORN.AddFunc(task.Spec, tasker)
	if err != nil {
		//zap.L().Error("项目重启后恢复定时任务失败,失败原因：", zap.Error(err))
		//zap.L().Error(fmt.Sprintf("该任务所属人：%s,所属机器人：%s,"+
		//"人物名：%s,任务具体消息:%s,任务具体定时规则：%s", username, robotname, message, detailTimeForUser))
		return
	}
	tid := int(TaskID)
	oldId := task.TaskID
	err = global.GLOAB_DB.Table("tasks").Where("task_id = ? ", oldId).Update("task_id", tid).Error
	if err != nil {
		//zap.L().Error("重启项目后更新任务id失败", zap.Error(err))
		return
	}
	return
}

func (t *DingRobot) EditTaskContent(p *EditTaskContentParam) (err error) {
	//根据任务id查询该任务的msg
	task := Task{
		Model: gorm.Model{ID: p.ID},
	}
	err = global.GLOAB_DB.Preload("MsgText.At.AtMobiles").Preload("MsgText.At.AtUserIds").Preload("MsgText.Text").First(&task).Error
	if err != nil {
		zap.L().Error("EditTaskContent err", zap.Error(err))
		return
	}
	task.MsgText.Text.Content = p.Content
	err = global.GLOAB_DB.Save(&task.MsgText.Text).Error
	if err != nil {
		zap.L().Error("EditTaskContent err", zap.Error(err))
		return
	}
	//杀死旧任务
	oldtaskId, _ := strconv.Atoi(task.TaskID)
	global.GLOAB_CORN.Remove(cron.EntryID(oldtaskId))
	//启动新任务
	paramCronTask := ParamCronTask{
		MsgText:     task.MsgText,
		MsgLink:     task.MsgLink,
		MsgMarkDown: task.MsgMarkDown,
		RobotId:     task.RobotId,
	}
	d := DingRobot{
		RobotId: task.RobotId,
		Secret:  task.Secret,
	}
	tasker := func() {
		err := d.SendMessage(&paramCronTask)
		if err != nil {
			zap.L().Error(fmt.Sprintf("重启定时任务失败"), zap.Error(err))
			return
		}
	}
	taskId, err := global.GLOAB_CORN.AddFunc(task.Spec, tasker)
	err = global.GLOAB_DB.Table("tasks").Where("task_id = ? ", task.TaskID).Update("task_id", taskId).Error
	if err != nil {
		zap.L().Error("重启项目后更新任务id失败", zap.Error(err))
		return
	}
	return
}

// 获取所有的公共机器人
func GetAllPublicRobot() (robot []DingRobot, err error) {
	//IsShare值为1为公共机器人
	err = global.GLOAB_DB.Where("is_shared=?", 1).Find(&robot).Error
	if err != nil {
		zap.L().Error("服务繁忙", zap.Error(err))
		return nil, err
	}
	return robot, err
}

func AlterResultByRobot(p *ParamAlterResultByRobot) (err error) {
	err = global.GLOAB_DB.Table("ding_depts").Where("dept_id", p.DeptId).Update("robot_token", p.Token).Error
	return
}

type result struct {
	ChatId string `json:"chatId"`
	Title  string `json:"title"`
}
type data struct {
	Result result `json:"result"`
}

func (t *DingRobot) RobotSendInviteCode(resp *RobotAtResp) error {
	code, expire, err := t.GetInviteCode()
	if err != nil {
		zap.L().Error("获取邀请码失败", zap.Error(err))
	}
	content := fmt.Sprintf(
		"欢迎加入闫佳鹏的打字邀请比赛\n网站: https://dazi.kukuw.com/\n邀请码: %v\n比赛剩余时间: %v",
		code, expire)
	if err != nil {
		content = "获取失败！"
	}
	param := &ParamChat{
		MsgKey:    "sampleText",
		MsgParam:  content,
		RobotCode: "dingepndjqy7etanalhi",
		UserIds:   []string{resp.SenderStaffId},
	}
	err = t.ChatSendMessage(param)
	if err != nil {
		zap.L().Error("单聊中发送打字邀请码错误" + err.Error())
		return err
	}
	return nil
}

func (t *DingRobot) RobotSendGroupInviteCode(resp *RobotAtResp) (res map[string]interface{}, err error) {
	code, expire, err := t.GetInviteCode()
	if err != nil {
		zap.L().Error("获取邀请码失败", zap.Error(err))
	}
	content := fmt.Sprintf(
		"欢迎加入闫佳鹏的打字邀请比赛\n网站: https://dazi.kukuw.com/\n邀请码: %v\n比赛剩余时间: %v",
		code, expire)
	if err != nil {
		content = "获取失败！"
	}
	param := &ParamChat{
		MsgKey:             "sampleText",
		MsgParam:           content,
		RobotCode:          "dingepndjqy7etanalhi",
		UserIds:            []string{resp.SenderStaffId},
		OpenConversationId: resp.ConversationId,
	}
	res, err = t.ChatSendGroupMessage(param)
	if err != nil {
		zap.L().Error("单聊中发送打字邀请码错误" + err.Error())
	}
	return res, nil
}
func (t *DingRobot) RobotSendGroupWater(resp *RobotAtResp) (res map[string]interface{}, err error) {
	param := &ParamChat{
		MsgKey:             "sampleText",
		MsgParam:           "送水师傅电话: 15236463964",
		RobotCode:          "dingepndjqy7etanalhi",
		UserIds:            []string{resp.SenderStaffId},
		OpenConversationId: resp.ConversationId,
	}
	res, err = t.ChatSendGroupMessage(param)
	if err != nil {
		zap.L().Error("发送送水师傅电话失败" + err.Error())
	}
	return res, nil
}
func (t *DingRobot) RobotSendGroupCard(resp *RobotAtResp) (res map[string]interface{}, err error) {
	param := &ParamChat{
		MsgKey: "sampleActionCard2",
		MsgParam: "{\n" +
			"        \"title\": \"帮助\",\n" +
			"        \"text\": \"请问你是否在查找以下功能\",\n" +
			"        \"actionTitle1\": \"送水电话号码\",\n" +
			fmt.Sprintf("'actionURL1':'dtmd://dingtalkclient/sendMessage?content=%s',\n", url.QueryEscape("送水电话号码")) +
			"        \"actionTitle2\": \"打字邀请码\",\n" +
			fmt.Sprintf("'actionURL2':'dtmd://dingtalkclient/sendMessage?content=%s',\n", url.QueryEscape("打字邀请码")) +
			"    }",
		RobotCode:          "dingepndjqy7etanalhi",
		UserIds:            []string{resp.SenderStaffId},
		OpenConversationId: resp.ConversationId,
	}
	res, err = t.ChatSendGroupMessage(param)
	if err != nil {
		zap.L().Error("发送chatSendMessage错误" + err.Error())
	}
	return res, nil
}

type Result struct {
	Name     string `json:"name"`
	DataName string `json:"data_name"`
	DataLink string `json:"data_link"`
}

//机器人问答发送卡片给个人https://open.dingtalk.com/document/isvapp/the-internal-robot-of-the-enterprise-realizes-the-interaction-in
func (t *DingRobot) RobotSendCardToPerson(resp *RobotAtResp, dataByStr []Result) (err error) {
	cardLen := len(dataByStr)
	if cardLen <= 5 {
	} else {
		cardLen = 5
	}
	action := ""
	var param *ParamChat
	if cardLen == 1 {
		//for i, data := range dataByStr {
		//	action += "        \"actionTitle" + strconv.Itoa(i+1) + "\": \"" + data.DataName + "\",\n" +
		//		fmt.Sprintf("'actionURL%d':'dtmd://dingtalkclient/sendMessage?content=%s',\n", i+1, url.QueryEscape(data.DataName))
		//}

		//action = fmt.Sprintf("\"singleTitle\": \"%s\",\n     \"singleURL\": \"%s\"", dataByStr[0].DataName, dataByStr[0].DataLink)
		for _, data := range dataByStr {
			action += "        \"singleTitle" + "\": \"" + data.DataName + "\",\n" +
				fmt.Sprintf("'singleURL':'dtmd://dingtalkclient/sendMessage?content=%s',\n", url.QueryEscape(data.DataName))
		}
		param = &ParamChat{
			MsgKey: "sampleActionCard",
			MsgParam: "{\n" +
				"        \"title\": \"资料\",\n" +
				"        \"text\": \"请问你是否在查找以下资料\",\n" +
				action +
				"    }",
			RobotCode: "dingepndjqy7etanalhi",
			UserIds:   []string{resp.SenderStaffId},
		}

	} else {
		for i, data := range dataByStr {
			action += "        \"actionTitle" + strconv.Itoa(i+1) + "\": \"" + data.DataName + "\",\n" +
				fmt.Sprintf("'actionURL%d':'dtmd://dingtalkclient/sendMessage?content=%s',\n", i+1, url.QueryEscape(data.DataName))
		}
		param = &ParamChat{
			MsgKey: "sampleActionCard" + strconv.Itoa(cardLen),
			MsgParam: "{\n" +
				"        \"title\": \"资料\",\n" +
				"        \"text\": \"请问你是否在查找以下资料\",\n" +
				action +
				"    }",
			RobotCode: "dingepndjqy7etanalhi",
			UserIds:   []string{resp.SenderStaffId},
		}
	}

	fmt.Println(action)

	err = t.ChatSendMessage(param)
	if err != nil {
		zap.L().Error("发送chatSendCardToPerson错误" + err.Error())
	}
	return
}

//机器人问答发送信息给个人
func (t *DingRobot) RobotSendMessageToPerson(resp *RobotAtResp, dataByStr []Result) (err error) {
	msg := ""
	if len(dataByStr) == 0 {
		msg = "您所查询的资源里没有此类资源"
	} else if len(dataByStr) == 1 {
		msg = dataByStr[0].DataLink
	} else {
		msg = "查询结果如下：\n"
		for _, data := range dataByStr {
			msg += "上传资料人员：" + data.Name + "\n" + "资源名称：" + data.DataName + "\n" + "资源内容：" + data.DataLink + "\n"
		}
	}
	param := &ParamChat{
		MsgKey:    "sampleText",
		MsgParam:  msg,
		RobotCode: "dingepndjqy7etanalhi",
		UserIds:   []string{resp.SenderStaffId},
	}
	err = t.ChatSendMessage(param)
	if err != nil {
		zap.L().Error("发送chatSendMessageToPerson错误" + err.Error())
	}
	return
}
