package dingding

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"ding/global"
	myselfRedis "ding/initialize/redis"
	"ding/model/params/ding"
	"ding/model/system"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const secret = "liwenzhou.com"

var wg = sync.WaitGroup{}
var hrefs []string
var blogs []string
var start int64 = 1676822400
var end int64 = 1677427200
var jin *DingUser

type Strs []string

type DingUser struct {
	UserId              string ` gorm:"primaryKey;foreignKey:UserId" json:"userid"`
	DingRobots          []DingRobot
	Deleted             gorm.DeletedAt
	Name                string                `json:"name"`
	Mobile              string                `json:"mobile"`
	Password            string                `json:"password"`
	DeptIdList          []int                 `json:"dept_id_list" gorm:"-"` //所属部门
	DeptList            []DingDept            `json:"dept_list" gorm:"many2many:user_dept"`
	AuthorityId         uint                  `json:"authorityId" gorm:"default:888;comment:用户角色ID"`
	Authority           system.SysAuthority   `json:"authority" gorm:"foreignKey:AuthorityId;references:AuthorityId;comment:用户角色"`
	Authorities         []system.SysAuthority `json:"authorities" gorm:"many2many:sys_user_authority;"`
	Title               string                `json:"title"` //职位
	JianShuAddr         string                `json:"jianshu_addr"`
	BlogAddr            string                `json:"blog_addr"`
	LeetCodeAddr        string                `json:"leet_code_addr"`
	AuthToken           string                `json:"auth_token" gorm:"-"`
	DingToken           `json:"ding_token" gorm:"-"`
	JianShuArticleURL   Strs `gorm:"type:longtext" json:"jian_shu_article_url"`
	BlogArticleURL      Strs `gorm:"type:longtext" json:"blog_article_url"`
	IsExcellentJianBlog bool `json:"is_excellentBlogJian" `
	Admin               bool `gorm:"-"  json:"admin"`
}

//用户签到
//如果dateStr没有传，那就是签到，如果传了特定日期，可以进行补签
//返回连续签到的次数
//用户签到
func (d *DingUser) Sign(year, uporDown, startWeek, weekDay, MNE int) (ConsecutiveSignNum int, err error) {
	//MNE 是上午下午晚上 1 2 3
	//构建redis中的key //singKey  user:sign:5:2023:1:19周        用户5在2023上半年第19周签到的记录
	//构建redis中的key //singKey  user:sign:5:2023:2:19周        用户5在2023下半年第19周签到的记录
	if weekDay == 0 {
		weekDay = 7
	}
	key := fmt.Sprintf(myselfRedis.UserSign+"%v:%v:%v:%v", d.UserId, year, uporDown, startWeek)
	//根据date能够判断出来，现在是第几周的上午下午晚上等
	offset := int64((weekDay-1)*3 + MNE - 1)
	IsSigned := global.GLOBAL_REDIS.GetBit(context.Background(), key, offset).Val()
	if IsSigned == 1 {
		ConsecutiveSignNum, _ = d.GetConsecutiveSignNum(year, uporDown, startWeek, weekDay, MNE)
		return ConsecutiveSignNum, errors.New("当前日期已经打卡签到，无需再次打卡签到")
	}
	//用户没有签到，设置成签到即可
	i, err := global.GLOBAL_REDIS.SetBit(context.Background(), key, offset, 1).Result()
	if err != nil || i != 1 {
		//此处返回的是设置前的值
		zap.L().Error("签到时操作redis中的位图失败", zap.Error(err))
	}
	ConsecutiveSignNum, _ = d.GetConsecutiveSignNum(year, uporDown, startWeek, weekDay, MNE)
	return
}

//统计用户在当前周连续签到的次数
func (d *DingUser) GetConsecutiveSignNum(year, uporDown, startWeek, weekDay, MNE int) (ConsecutiveSignNum int, err error) {
	if startWeek == 0 {
		startWeek = 7
	}
	key := fmt.Sprintf(myselfRedis.UserSign+"%v:%v:%v:%v", d.UserId, year, uporDown, startWeek)
	//bitfield可以操作多个位 bitfile user:sign:2023:1:19 u7 0  //从索引零开始，往后面统计7天的
	//cmd := global.GLOBAL_REDIS.Do(context.Background(), "BITFIELD", key, "GET", "u"+strconv.Itoa(weekDay), "0").
	list, err := global.GLOBAL_REDIS.BitField(context.Background(), key, "GET", "u"+strconv.Itoa(weekDay), "0").Result()
	if err != nil || list == nil || len(list) == 0 || list[0] == 0 {
		return 0, nil
	}
	// 此处获得的值是经过二进制转化过来的，总共有21个字节，如果长度是21个字节的话，可能会非常的大，我们如何处理非常大的值呢？
	//具体思路可以使用位运算，具体博客链接
	v := list[0]
	for i := weekDay; i > 0; i-- {
		for j := 0; j < 3; j++ {
			//如果这个很大的数字转化为二进制之后，左移动一位，右移动一位，如果还等于自己，说明最后一位是0，表示没有签到
			if v>>1<<1 == v {
				if !(i == weekDay && j == MNE) {
					//低位为0 且 非当天早中晚应该签到的时间，签到中断
					break
				}
			} else {
				//说明签到了
				ConsecutiveSignNum++
			}
		}
		//将v右移一位，并重新复制，相当于最低位提前了一天
		v = v >> 1
	}
	return
}

type days struct {
	morning bool
	midday  bool
	night   bool
}

//统计用户当前周签到的详情情况
func (d *DingUser) GetWeekSignDetail(year, uporDown, startWeek int) (result map[int][]bool, err error) {
	result = make(map[int][]bool, 0)
	//if year == 0 || uporDown == 0 || startWeek == 0 {
	//	curTime, _ := (&localTime.MySelfTime{}).GetCurTime(nil)
	//
	//}
	//使用bitFiled来获取int64，然后使用位运算计算结果
	key := fmt.Sprintf(myselfRedis.UserSign+"%v:%v:%v:%v", d.UserId, year, uporDown, startWeek)
	fmt.Println(key)
	list, err := global.GLOBAL_REDIS.BitField(context.Background(), key, "GET", "u"+strconv.Itoa(21), "0").Result()
	if err != nil || list == nil || len(list) == 0 || list[0] == 0 {
		zap.L().Error("使用redis中的bitmap失败", zap.Error(err))
		return nil, errors.New("使用redis中的bitmap失败")
	}
	v := list[0]
	//110001111111111101111000
	//for i := 1; i <= 8; i++ {
	//	if v>>1<<1 == v {
	//		//说明没有签到
	//		result[i] = append(result[i], false)
	//	} else {
	//		//说明签到了
	//		result[i] = append(result[i], true)
	//	}
	//	v = v >> 1
	//	x = x >> 1
	//}
	for i := 7; i > 0; i-- {
		for j := 0; j < 3; j++ {
			if v>>1<<1 == v {
				//说明没有签到
				result[i] = append(result[i], false)
			} else {
				//说明签到了
				result[i] = append(result[i], true)
				//result[i][j] = true
			}
			v = v >> 1
		}
	}
	return
}

//统计用户一周的签到次数（非连续）
func (d *DingUser) GetWeekSignNum(year, uporDown, startWeek int) (WeekSignNum int64, err error) {
	//需要使用redis中的bitmap中bigcount方法来统计
	//构建key
	key := fmt.Sprintf(myselfRedis.UserSign+"%v:%v:%v:%v", d.UserId, year, uporDown, startWeek)

	bitCount := &redis.BitCount{
		Start: 0, //都设置成0就是涵盖整个bitmap
		End:   0,
	}
	WeekSignNum, err = global.GLOBAL_REDIS.BitCount(context.Background(), key, bitCount).Result()
	if err != nil {
		zap.L().Error("使用redis的BitCount失败", zap.Error(err))
		return
	}
	return
}

//通过userid查询部门id
func GetDeptByUserId(UserId string) (user *DingUser) {
	err := global.GLOAB_DB.Where("user_id = ?", UserId).Preload("DeptList").First(&user).Error
	if err != nil {
		zap.L().Error("通过userid查询用户失败", zap.Error(err))
		return
	}
	return
}
func (d *DingUser) SendFrequencyLeave(start int) error {
	fmt.Println("推送个人请假频率")
	return nil
}
func (d *DingUser) CountFrequencyLeave(startWeek int, result map[string][]DingAttendance) error {
	fmt.Println("存储个人请假频率")
	return nil
}

type JinAndBlog struct {
	UserId            string `gorm:"primary_key" json:"id"`
	Name              string `json:"name"`
	JianShuArticleURL Strs   `gorm:"type:longtext" json:"jian_shu_article_url"`
	BlogArticleURL    Strs   `gorm:"type:longtext" json:"blog_article_url"`
	IsExcellent       bool   `json:"is_excellent"`
}

func (d *DingUser) GetUserByUserId() (user DingUser, err error) {
	err = global.GLOAB_DB.Where("user_id = ?", d.UserId).First(&user).Error
	return
}
func (d *DingUser) GetUserInfo() (err error) {
	err = global.GLOAB_DB.Where("user_id = ?", d.UserId).Preload("Authority").Preload("Authorities").First(&d).Error
	return
}

func (m *DingUser) UserAuthorityDefaultRouter(user *DingUser) {
	var menuIds []string
	err := global.GLOAB_DB.Model(&system.SysAuthorityMenu{}).Where("sys_authority_authority_id = ?", user.AuthorityId).Pluck("sys_base_menu_id", &menuIds).Error
	if err != nil {
		return
	}
	var am system.SysBaseMenu
	err = global.GLOAB_DB.First(&am, "name = ? and id in (?)", user.Authority.DefaultRouter, menuIds).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user.Authority.DefaultRouter = "404"
	}
}
func (d *DingUser) Login() (user *DingUser, err error) {
	user = &DingUser{
		Mobile:   d.Mobile,
		Password: d.Password,
	}
	//判断该用户是否存在
	err = global.GLOAB_DB.Model(DingUser{}).Where("mobile", d.Mobile).First(user).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	//判断密码是否正确
	if user.Password != d.Password {
		return nil, errors.New("密码错误")
	}
	//此处的Login函数传递的是一个指针类型的数据
	opassword := user.Password //此处是用户输入的密码，不一定是对的
	err = global.GLOAB_DB.Where(&DingUser{Mobile: user.Mobile}).Preload("Authorities").Preload("Authority").First(user).Error
	if err != nil {
		zap.L().Error("登录时查询数据库失败", zap.Error(err))
		return
	}
	//如果到了这里还没有结束的话，那就说明该用户至少是存在的，于是我们解析一下密码
	//password := encryptPassword(opassword)
	password := opassword
	//拿到解析后的密码，我们看看是否正确
	if password != user.Password {
		return nil, errors.New("密码错误")
	}
	d.UserAuthorityDefaultRouter(user)
	//如果能到这里的话，那就登录成功了
	return

}
func encryptPassword(oPassword string) string {
	h := md5.New()
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum([]byte(oPassword)))
}

//https://open.dingtalk.com/document/isvapp/query-user-details
func (d *DingUser) GetUserDetailByUserId() (user DingUser, err error) {
	var client *http.Client
	var request *http.Request
	var resp *http.Response
	var body []byte
	URL := "https://oapi.dingtalk.com/topapi/v2/user/get?access_token=" + d.DingToken.Token
	client = &http.Client{Transport: &http.Transport{ //对客户端进行一些配置
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}, Timeout: time.Duration(time.Second * 5)}
	//此处是post请求的请求题，我们先初始化一个对象
	b := struct {
		UserId string `json:"userid"`
	}{UserId: d.UserId}

	//然后把结构体对象序列化一下
	bodymarshal, err := json.Marshal(&b)
	if err != nil {
		return
	}
	//再处理一下
	reqBody := strings.NewReader(string(bodymarshal))
	//然后就可以放入具体的request中的
	request, err = http.NewRequest(http.MethodPost, URL, reqBody)
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
		DingResponseCommon
		User DingUser `json:"result"` //必须大写，不然的话，会被忽略，从而反序列化不上
	}{}
	//把请求到的结构反序列化到专门接受返回值的对象上面
	err = json.Unmarshal(body, &r)
	if err != nil {
		return
	}
	if r.Errcode != 0 {
		return DingUser{}, errors.New(r.Errmsg)
	}
	// 此处举行具体的逻辑判断，然后返回即可

	return r.User, nil

}

// ImportUserToMysql 把钉钉用户信息导入到数据库中
func (d *DingUser) ImportUserToMysql() error {
	return global.GLOAB_DB.Transaction(func(tx *gorm.DB) error {
		token, err := (&DingToken{}).GetAccessToken()
		if err != nil {
			zap.L().Error("从redis中取出token失败", zap.Error(err))
			return err
		}
		var deptIdList []int
		err = tx.Model(&DingDept{}).Select("dept_id").Find(&deptIdList).Error
		if err != nil {
			zap.L().Error("从数据库中取出所有部门id失败", zap.Error(err))
			return err
		}
		for i := 0; i < len(deptIdList); i++ {
			Dept := &DingDept{DeptId: deptIdList[i], DingToken: DingToken{Token: token}}
			DeptUserList, _, err := Dept.GetUserListByDepartmentID(0, 100)

			if err != nil {
				zap.L().Error("获取部门成员失败", zap.Error(err))
				continue
			}
			tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "title"}),
			}).Create(&DeptUserList)
			if err != nil {
				zap.L().Error(fmt.Sprintf("存储部门:%s成员到数据库失败", Dept.Name), zap.Error(err))
				continue
			}
		}
		return err
	})

}

func (d *DingUser) FindDingUsers(name, mobile string) (us []DingUser, err error) {
	db := global.GLOAB_DB.Model(&DingUser{})
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if mobile != "" {
		db = db.Where("mobile like ?", "%"+mobile+"%")
	}
	err = db.Select("user_id", "name", "mobile").Find(&us).Error
	//keys, err := global.GLOBAL_REDIS.Keys(context.Background(), "user*").Result()
	//往redis中做一份缓存
	//for i := 0; i < len(us); i++ {
	//	batchData := make(map[string]interface{})
	//	batchData["name"] = us[i].Name
	//	_, err := global.GLOBAL_REDIS.HMSet(context.Background(), "user:"+us[i].UserId, batchData).Result()
	//	if err != nil {
	//		zap.L().Error("把数据缓存到redis中失败", zap.Error(err))
	//	}
	//}
	return
}

// UpdateDingUserAddr 根据用户id修改其简书和博客地址
func (d *DingUser) UpdateDingUserAddr(userParam ding.UserAndAddrParam) error {
	return global.GLOAB_DB.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Model(&DingUser{}).Where("user_id = ?", userParam.UserId).Updates(DingUser{BlogAddr: userParam.BlogAddr, JianShuAddr: userParam.JianShuAddr}).Error; err != nil {
			zap.L().Error("更新用户博客和简书地址失败", zap.Error(err))
			return
		}
		return
	})
}

func (d *DingUser) GoCrawlerDingUserJinAndBlog() (err error) {
	//spec = "00 03,33,45 08,14,21 * * ?"
	//spec := "50 30 21 * * 1"
	//task := func() {
	err = global.GLOAB_DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&DingUser{}).UpdateColumn("is_excellent_jian_blog", false).Error
	if err != nil {
		zap.L().Error("爬取文章之前清空之前的优秀简书博客人员", zap.Error(err))
	}
	wg.Add(2)
	//go UpdateDingUserHref()
	//UpdateDingUserJianShu()
	//go UpdateDingUserBlog()
	UpdateDingUserBlog()
	wg.Wait()
	zap.L().Info("爬取完毕，已成功存入数据库")
	//}
	//taskId, err := global.GLOAB_CORN.AddFunc(spec, task)

	if err != nil {
		zap.L().Error("启动爬虫爬取文文章地址失败", zap.Error(err))
	}
	//zap.L().Info(fmt.Sprintf("启动爬虫爬取文文章地址定时任务成功（非爬虫成功），定时任务id:%v", taskId))
	return err
}
func UpdateDingUserJianShu() {
	//获取所有人的博客和简书主页链接
	urls, err := (&DingUser{}).FindDingUserAddr()
	if err != nil {
		zap.L().Error("获取简书链接错误", zap.Error(err))
		return
	}

	for _, v := range urls {
		falg := true
		if v.JianShuAddr == "" {
			//为了避免偶然因素
			err := v.UpdateDingUserHref(hrefs, v.UserId)
			if err != nil {
				fmt.Println("空简书链接未清空")
			}
			continue
		}
		client := http.Client{}
		v.JianShuAddr = strings.ReplaceAll(v.JianShuAddr, "\n", "")
		//strings.Replace(v.JianShuAddr, "\n", "", -1)
		req, err := http.NewRequest("GET", v.JianShuAddr, nil)
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Content-Type", "text/html; charset=utf-8")
		req.Header.Set("Keep-Alive", "timeout=30")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		//解析网页
		docDetail, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		docDetail.Find("#list-container > ul > li ").Each(func(i int, selection *goquery.Selection) {
			if falg == false {
				return
			}
			//获取文章链接
			attr, exists := selection.Find(".content > a").Attr("href")
			//获取时间，格式为utc时间
			text, exits := selection.Find(".content > .meta > span.time").Attr("data-shared-at")
			if exits == true {
				//拿到当前时刻的时间戳和最近一篇简书的时间戳
				timeUnix := time.Now().Unix()
				timeUnix = 1678022322
				t1 := UTCTransLocal(text)
				unix := switchTime(t1)
				//先把之前的文章进行一下清空
				err = v.UpdateDingUserHref(hrefs, v.UserId)
				if err != nil {
					zap.L().Error("清空记录失败", zap.Error(err))
				}
				for end < timeUnix {
					start += 604800
					end += 604800
				}
				if unix >= start && unix < end {
					zap.L().Info(fmt.Sprintf("%v简书文章爬取成功，文章链接：%v,文章发布时间：%v", v.Name, attr, text))
					if exists == true {
						str := "https://www.jianshu.com" + attr
						hrefs = append(hrefs, str)
					}
				} else {
					zap.L().Info(fmt.Sprintf("%v简书文章爬取成功，旦不满足时间要求，结束爬取，文章链接：%v,文章发布时间：%v", v.Name, attr, text))
					falg = false
					return
				}

			}

		})

		err = v.UpdateDingUserHref(hrefs, v.UserId)
		if err != nil {
			zap.L().Error("更新简书数组到数据库出错", zap.Error(err))
			return
		}
		hrefs = []string{}
	}
	wg.Done()
}
func UpdateDingUserBlog() {

	urls, err := jin.FindDingUserAddr()
	if err != nil {
		zap.L().Error("获取简书链接错误", zap.Error(err))
		return
	}
	for _, v := range urls {
		if v.Name != "闫佳鹏" {
			continue
		}
		if v.BlogAddr == "" {
			err = v.UpdateDingUserBlog(blogs, v.UserId)
			if err != nil {
				fmt.Println("空博客链接未清空")
			}
			zap.L().Info(fmt.Sprintf("%v博客链接是空，直接跳过", v.Name))
			continue
		}
		target := v.BlogAddr + "/article/list"
		htmls, err := GetHtml(target)
		if err != nil {
			fmt.Println(err)
			panic("Get target ERROR!!!")
		}

		var html string
		html = strings.Replace(string(htmls), "\n", "", -1)
		html = strings.Replace(string(htmls), " ", "", -1)

		//fmt.Println(html)
		reBlog := regexp.MustCompile(`<div class="article-item-box csdn-tracking-statistics(.*?)</div>`)
		reLink := regexp.MustCompile(`href="(.*?)"`)
		reTime := regexp.MustCompile(`<span class="date">(.*?)</span>`)

		articles := reBlog.FindAllString(html, -1)
		if articles == nil || len(articles) == 0 {
			zap.L().Info(fmt.Sprintf("%s本周未写博客", v.Name))
			continue
		}
		for _, value := range articles {
			BlogLink := reLink.FindAllStringSubmatch(value, -1)

			BlogTime := reTime.FindAllStringSubmatch(value, -1)

			timeUnix := time.Now().Unix()
			timeUnix = 1678024664
			t1 := UTCTransLocal(BlogTime[0][1])
			unix := switchTime(t1)

			err = v.UpdateDingUserBlog(blogs, v.UserId)
			if err != nil {
				zap.L().Error("清空博客数据失败", zap.Error(err))
				return
			}
			for end < timeUnix {
				start += 604800
				end += 604800
			}
			if unix >= start && unix < end {
				hrefs = append(hrefs, BlogLink[0][1])
				zap.L().Info(fmt.Sprintf("%v博客文章爬取成功，文章链接：%v,文章发布时间：%v", v.Name, BlogLink, BlogTime))
			} else {
				zap.L().Info(fmt.Sprintf("%v博客文章爬取成功，旦不满足时间要求，结束爬取，文章链接：%v,文章发布时间：%v", v.Name, BlogLink, BlogTime))
				break
			}
		}
		err = v.UpdateDingUserBlog(blogs, v.UserId)
		if err != nil {
			zap.L().Error("更新博客数组到数据库出错", zap.Error(err))
			return
		}
		blogs = []string{}
	}
	wg.Done()
}

func switchTime(ans string) (unix int64) {
	loc, _ := time.LoadLocation("Local")
	theTime, err := time.ParseInLocation("2006-01-02 15:04:05", ans, loc)
	if err == nil {
		unix = theTime.Unix()
		return unix
	}
	return
}

func UTCTransLocal(utcTime string) string {
	loc, _ := time.LoadLocation("Local")
	t, _ := time.ParseInLocation("2006-01-02T15:04:05+08:00", utcTime, loc)
	return t.Local().Format("2006-01-02 15:04:05")
}

func GetHtml(URL string) (html []byte, err error) {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    10 * time.Second,
		DisableCompression: true,
		// Proxy:              http.ProxyURL(proxyUrl),
	}

	req, err := http.NewRequest("GET", URL, nil)
	req.Header.Add("UserAgent", " Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.41")

	client := &http.Client{
		Transport: tr, /*使用transport参数*/
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	html, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return html, err
}

func (d *DingUser) FindDingUserAddr() (addrs []DingUser, err error) {
	var address []DingUser
	err = global.GLOAB_DB.Model(&DingUser{}).Select("jian_shu_addr", "blog_addr", "user_id", "name").Find(&address).Error
	if err != nil {
		zap.L().Error("获取钉钉用户的简书或博客链接失败", zap.Error(err))
		return
	}
	return address, nil
}

func (d *DingUser) UpdateDingUserHref(jins Strs, id string) (err error) {
	err = global.GLOAB_DB.Model(&DingUser{}).Where("user_id = ?", id).UpdateColumns(map[string]interface{}{
		"jian_shu_article_url": jins,
	}).Error
	if err != nil {
		zap.L().Error("在mysql中更新这周简书链接失败", zap.Error(err))
		return
	}
	return nil
}

func (d *DingUser) UpdateDingUserBlog(blogs Strs, id string) (err error) {
	err = global.GLOAB_DB.Model(&DingUser{}).Where("user_id = ?", id).UpdateColumns(map[string]interface{}{
		"blog_article_url": blogs,
	}).Error
	if err != nil {
		zap.L().Error("在mysql中更新这周博客链接失败", zap.Error(err))
		return
	}
	return nil
}

// 获取二维码buf，chatId, title
func (u *DingUser) GetQRCodeInWindows(c *gin.Context) (buf []byte, chatId, title string, err error) {
	zap.L().Info("进入到了chromedp")
	d := data{}
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
		chromedp.Flag("headless", false),
		chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
		chromedp.Flag("ignore-certificate-errors", true),      //忽略错误
		chromedp.Flag("disable-web-security", true),           //禁用网络安全标志
		chromedp.Flag("disable-extensions", true),             //开启插件支持
		chromedp.Flag("disable-default-apps", true),
		chromedp.NoFirstRun, //设置网站不是首次运行
		chromedp.WindowSize(1921, 1024),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"), //设置UserAgent
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	//defer cancel()
	print(cancel)

	// 创建上下文实例
	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// 创建超时上下文
	ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	// navigate to a page, wait for an element, click

	// capture screenshot of an element

	// capture entire browser viewport, returning png with quality=90
	var html string
	fmt.Println("开始运行chromedp")
	err = chromedp.Run(ctx,
		//打开网页
		chromedp.Navigate(`https://open-dev.dingtalk.com/apiExplorer?spm=ding_open_doc.document.0.0.20bf4063FEGqWg#/jsapi?api=biz.chat.chooseConversationByCorpId`),
		//定位登录按钮
		chromedp.Click(`document.querySelector(".ant-btn.ant-btn-primary")`, chromedp.ByJSPath),
		//等二维码出现
		chromedp.WaitVisible(`document.querySelector(".ant-modal")`, chromedp.ByJSPath),
		//截图
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, _, _, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			buf, err = page.CaptureScreenshot().
				WithQuality(90).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			username, _ := c.Get(global.CtxUserNameKey)
			fmt.Println(username)
			err = ioutil.WriteFile(fmt.Sprintf("./Screenshot_%s.png", username), buf, 0644)
			if err != nil {
				zap.L().Error("二维码写入失败", zap.Error(err))
			}
			zap.L().Info("写入二维码成功", zap.Error(err))
			return nil
		}),
		//等待用户扫码连接成功
		chromedp.WaitVisible(`document.querySelector(".connect-info")`, chromedp.ByJSPath),
		//chromedp.SendKeys(`document.querySelector("#corpId")`, "caonima",chromedp.ByJSPath),
		//设置输入框中的值为空
		chromedp.SetValue(`document.querySelector("#corpId")`, "", chromedp.ByJSPath),
		//chromedp.Click(`document.querySelector(".ant-btn.ant-btn-primary")`, chromedp.ByJSPath),
		//chromedp.Clear(`#corpId`,chromedp.ByID),
		//输入正确的值
		chromedp.SendKeys(`document.querySelector("#corpId")`, "ding7625646e1d05915a35c2f4657eb6378f", chromedp.ByJSPath),
		//点击发起调用按钮
		chromedp.Click(`document.querySelector(".ant-btn.ant-btn-primary")`, chromedp.ByJSPath),

		chromedp.WaitVisible(`document.querySelector("#dingapp > div > div > div.api-explorer-wrap > div.api-info > div > div.ant-tabs-content.ant-tabs-content-animated.ant-tabs-top-content > div.ant-tabs-tabpane.ant-tabs-tabpane-active > div.debug-result > div.code-mirror > div.code-content > div > div > div.CodeMirror-scroll > div.CodeMirror-sizer > div > div > div > div.CodeMirror-code > div:nth-child(2) > pre > span > span.cm-tab")`, chromedp.ByJSPath),
		//自定义函数进行爬虫
		chromedp.ActionFunc(func(ctx context.Context) error {
			//b := chromedp.WaitEnabled(`document.querySelector("#dingapp > div > div > div.api-explorer-wrap > div.api-info > div > div.ant-tabs-content.ant-tabs-content-animated.ant-tabs-top-content > div.ant-tabs-tabpane.ant-tabs-tabpane-active > div.debug-result > div.code-mirror > div.code-content > div > div > div.CodeMirror-scroll > div.CodeMirror-sizer > div > div > div > div.CodeMirror-code > div > pre")`, chromedp.ByJSPath)
			//b.Do(ctx)
			a := chromedp.OuterHTML(`document.querySelector("body")`, &html, chromedp.ByJSPath)
			a.Do(ctx)
			dom, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				fmt.Println("123", err.Error())
				return err
			}
			var data string
			dom.Find("#dingapp > div > div > div.api-explorer-wrap > div.api-info > div > div.ant-tabs-content.ant-tabs-content-animated.ant-tabs-top-content > div.ant-tabs-tabpane.ant-tabs-tabpane-active > div.debug-result > div.code-mirror > div.code-content > div > div > div.CodeMirror-scroll > div.CodeMirror-sizer > div > div > div > div.CodeMirror-code > div > pre").Each(func(i int, selection *goquery.Selection) {
				data = data + selection.Text()
				selection.Next()
			})
			data = strings.ReplaceAll(data, " ", "")
			data = strings.ReplaceAll(data, "\n", "")
			reader := strings.NewReader(data)
			bytearr, err := ioutil.ReadAll(reader)

			err1 := json.Unmarshal(bytearr, &d)
			if err1 != nil {

			}
			return nil
		}),
	)
	if err != nil {
		zap.L().Error("使用chromedp失败", zap.Error(err))
		return nil, "", "", err
	}
	if &d == nil {
		return nil, "", "", err
	}
	return buf, d.Result.ChatId, d.Result.Title, err
}

var ChromeCtx context.Context

func GetChromeCtx(focus bool) context.Context {
	if ChromeCtx == nil || focus {
		allocOpts := chromedp.DefaultExecAllocatorOptions[:]
		allocOpts = append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.DisableGPU,
			//chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
			//chromedp.Flag("headless", false),
			chromedp.Flag("blink-settings", "imagesEnabled=false"), //开启图像界面,重点是开启这个
			chromedp.Flag("ignore-certificate-errors", true),       //忽略错误
			chromedp.Flag("disable-web-security", true),            //禁用网络安全标志
			chromedp.Flag("disable-extensions", true),              //开启插件支持
			chromedp.Flag("accept-language", `zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,zh-TW;q=0.6`),
			//chromedp.Flag("disable-default-apps", true),
			//chromedp.NoFirstRun, //设置网站不是首次运行
			chromedp.WindowSize(1921, 1024),
			chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"), //设置UserAgent
		)

		if checkChromePort() {
			// 不知道为何，不能直接使用 NewExecAllocator ，因此增加 使用 ws://172.17.0.7:9222/ 来调用
			c, _ := chromedp.NewRemoteAllocator(context.Background(), "ws://172.17.0.7:9222/")
			ChromeCtx, _ = chromedp.NewContext(c)
		} else {
			c, _ := chromedp.NewExecAllocator(context.Background(), allocOpts...)
			ChromeCtx, _ = chromedp.NewContext(c)
		}
	}
	return ChromeCtx
}
func (u *DingUser) GetQRCodeInLinux(c *gin.Context) (buf []byte, chatId, title string, err error) {
	timeCtx, cancel := context.WithTimeout(GetChromeCtx(false), 5*time.Minute)
	defer cancel()
	d := data{}
	var html string
	zap.L().Info("开始运行chromedp")
	err = chromedp.Run(timeCtx,
		//打开网页
		chromedp.Navigate(`https://open-dev.dingtalk.com/apiExplorer?spm=ding_open_doc.document.0.0.20bf4063FEGqWg#/jsapi?api=biz.chat.chooseConversationByCorpId`),
		//定位登录按钮
		chromedp.Click(`document.querySelector(".ant-btn.ant-btn-primary")`, chromedp.ByJSPath),
		//等二维码出现
		chromedp.WaitVisible(`document.querySelector(".ant-modal")`, chromedp.ByJSPath),
		//截图
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, _, _, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			buf, err = page.CaptureScreenshot().
				WithQuality(90).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			username, _ := c.Get(global.CtxUserNameKey)
			err = ioutil.WriteFile(fmt.Sprintf("./Screenshot_%s.png", username), buf, 0644)
			if err != nil {
				zap.L().Error("二维码写入失败", zap.Error(err))
			}
			return nil
		}),
		//等待用户扫码连接成功

		chromedp.WaitVisible(`document.querySelector(".connect-info")`, chromedp.ByJSPath),
		//chromedp.SendKeys(`document.querySelector("#corpId")`, "caonima",chromedp.ByJSPath),
		//设置输入框中的值为空
		chromedp.SetValue(`document.querySelector("#corpId")`, "", chromedp.ByJSPath),
		//chromedp.Click(`document.querySelector(".ant-btn.ant-btn-primary")`, chromedp.ByJSPath),
		//chromedp.Clear(`#corpId`,chromedp.ByID),
		//输入正确的值
		chromedp.SendKeys(`document.querySelector("#corpId")`, "ding7625646e1d05915a35c2f4657eb6378f", chromedp.ByJSPath),
		//点击发起调用按钮
		chromedp.Click(`document.querySelector("#dingapp > div > div > div.api-explorer-wrap > div.param-list > div > div.api-param-footer > button")`, chromedp.ByJSPath),
		//chromedp.Click(`document.querySelector(".ant-btn.ant-btn-primary")`, chromedp.ByJSPath),
		chromedp.WaitVisible(`document.querySelector("#dingapp > div > div > div.api-explorer-wrap > div.api-info > div > div.ant-tabs-content.ant-tabs-content-animated.ant-tabs-top-content > div.ant-tabs-tabpane.ant-tabs-tabpane-active > div.debug-result > div.code-mirror > div.code-content > div > div > div.CodeMirror-scroll > div.CodeMirror-sizer > div > div > div > div.CodeMirror-code > div:nth-child(2) > pre > span > span.cm-tab")`, chromedp.ByJSPath),
		//自定义函数进行爬虫
		chromedp.ActionFunc(func(ctx context.Context) error {
			//b := chromedp.WaitEnabled(`document.querySelector("#dingapp > div > div > div.api-explorer-wrap > div.api-info > div > div.ant-tabs-content.ant-tabs-content-animated.ant-tabs-top-content > div.ant-tabs-tabpane.ant-tabs-tabpane-active > div.debug-result > div.code-mirror > div.code-content > div > div > div.CodeMirror-scroll > div.CodeMirror-sizer > div > div > div > div.CodeMirror-code > div > pre")`, chromedp.ByJSPath)
			//b.Do(ctx)
			a := chromedp.OuterHTML(`document.querySelector("body")`, &html, chromedp.ByJSPath)
			a.Do(ctx)
			dom, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				fmt.Println("123", err.Error())
				return err
			}
			var data string
			dom.Find("#dingapp > div > div > div.api-explorer-wrap > div.api-info > div > div.ant-tabs-content.ant-tabs-content-animated.ant-tabs-top-content > div.ant-tabs-tabpane.ant-tabs-tabpane-active > div.debug-result > div.code-mirror > div.code-content > div > div > div.CodeMirror-scroll > div.CodeMirror-sizer > div > div > div > div.CodeMirror-code > div > pre").Each(func(i int, selection *goquery.Selection) {
				data = data + selection.Text()
				selection.Next()
			})
			data = strings.ReplaceAll(data, " ", "")
			data = strings.ReplaceAll(data, "\n", "")
			reader := strings.NewReader(data)
			bytearr, err := ioutil.ReadAll(reader)

			err1 := json.Unmarshal(bytearr, &d)
			if err1 != nil {

			}
			return nil
		}),
	)
	if err != nil {
		zap.L().Error("使用chromedp失败", zap.Error(err))
		return nil, "", "", err
	}
	if &d == nil {
		return nil, "", "", err
	}
	return buf, d.Result.ChatId, d.Result.Title, err
}
func checkChromePort() bool {
	addr := net.JoinHostPort("", "9222")
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
func (u *DingUser) GetRobotList() (RobotList []DingRobot, err error) {
	//err = global.GLOAB_DB.Where("ding_user_id = ?", u.UserId).Find(&RobotList).Error
	err = global.GLOAB_DB.Model(u).Association("DingRobots").Find(&RobotList)
	return
}

func (u *DingRobot) GetOpenConversationId(access_token, chatId string) (openConversationId string, _err error) {
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
		openConversationId = *(result.Body.OpenConversationId)
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
