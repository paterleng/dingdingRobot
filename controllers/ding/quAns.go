package ding

import (
	"context"
	"ding/global"
	"ding/model/dingding"
	"ding/response"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

func GetRedisRoad(data *Data, UserId string) (redisRoad string) {
	redisRoad = "learningData:"
	if data.Type == 1 {
		//公共存储
		redisRoad += "public:" + UserId + ":"
	} else if data.Type == 2 {
		deptList := dingding.GetDeptByUserId(UserId).DeptList
		for _, dept := range deptList {
			redisRoad = "learningData:"
			if dept.DeptId == data.DeptId {
				//部门存储
				redisRoad += "dept:" + strconv.Itoa(data.DeptId) + ":" + UserId + ":"
				return
			}
		}

	} else if data.Type == 3 {
		//个人存储
		redisRoad += "personal:" + UserId + ":"
	}
	return
}

//上传资源
func UpdateData(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if err != nil {
		zap.L().Error("token获取userid失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	var data *Data
	err = c.ShouldBindJSON(&data)
	if err != nil {
		zap.L().Error("JSON绑定错误", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	ctx := context.Background()
	redisRoad := GetRedisRoad(data, UserId)
	if redisRoad == "learningData:" {
		response.FailWithMessage("您不是该部门的成员，不能添加资源", c)
		return
	}
	result, err := global.GLOBAL_REDIS.HExists(ctx, redisRoad, data.DataName).Result()
	if err != nil {
		zap.L().Error("查询redis中是否存在该名字失败", zap.Error(err))
	}
	if result {
		zap.L().Info("已存在该文件名称")
		response.FailWithMessage("已存在该文件名称", c)
	} else {
		err = global.GLOBAL_REDIS.HSet(ctx, redisRoad, data.DataName, data.DataLink).Err()
		if err != nil {
			zap.L().Error("将文件名称链接存储进redis中失败", zap.Error(err))
			response.FailWithMessage("参数错误", c)
			return
		}
	}
	response.OkWithMessage("上传成功", c)
}

//删除资源
func DeleteData(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if err != nil {
		zap.L().Error("token获取userid失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	var data *Data
	err = c.ShouldBindJSON(&data)
	if err != nil {
		zap.L().Error("JSON绑定错误", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	ctx := context.Background()
	redisRoad := GetRedisRoad(data, UserId)
	if redisRoad == "learningData:" {
		response.FailWithMessage("这不是您上传的资源，您没有权限删除", c)
		return
	} else {
		exist, err := global.GLOBAL_REDIS.HExists(ctx, redisRoad, data.DataName).Result()
		if err != nil {
			zap.L().Error("从redis中查询是否存在失败", zap.Error(err))
			response.FailWithMessage("参数错误", c)
			return
		}
		dataLink, err := global.GLOBAL_REDIS.HGet(ctx, redisRoad, data.DataName).Result()
		if err != nil {
			zap.L().Error("从redis中没有查到该DataName", zap.Error(err))
			response.FailWithMessage("这不是您上传的资源，您没有权限删除", c)
			return
		}
		user, err := (&dingding.DingUser{UserId: UserId}).GetUserByUserId()
		if user.Name != data.UserName || !exist || data.DataLink != dataLink {
			response.FailWithMessage("这不是您上传的资源，您没有权限删除", c)
			return
		}
		err = global.GLOBAL_REDIS.HDel(ctx, redisRoad, data.DataName).Err()
		if err != nil {
			zap.L().Error("将文件名称链接从redis中删除失败", zap.Error(err))
			response.FailWithMessage("参数错误", c)
			return
		}
	}
	response.OkWithMessage("删除成功", c)
}

//修改资源
func PutData(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if err != nil {
		zap.L().Error("token获取userid失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	var data *Data
	err = c.ShouldBindJSON(&data)
	if err != nil {
		zap.L().Error("JSON绑定错误", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	ctx := context.Background()
	redisRoad := GetRedisRoad(data, UserId)
	user, err := (&dingding.DingUser{UserId: UserId}).GetUserByUserId()
	//判断是否已存在该键
	exist, err := global.GLOBAL_REDIS.HExists(ctx, redisRoad, data.OldDataName).Result()
	if err != nil {
		zap.L().Error("从redis中查询是否存在失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	if !exist || data.UserName != user.Name {
		response.FailWithMessage("这不是您上传的资源，您没有权限删除", c)
		return
	} else {
		exist, err := global.GLOBAL_REDIS.HExists(ctx, redisRoad, data.DataName).Result()
		if err != nil {
			zap.L().Error("从redis中查询是否存在失败", zap.Error(err))
			response.FailWithMessage("参数错误", c)
			return
		}
		oldDataLink, err := global.GLOBAL_REDIS.HGet(ctx, redisRoad, data.OldDataName).Result()
		if err != nil {
			zap.L().Error("从redis中没有查到该DataName", zap.Error(err))
			response.FailWithMessage("这不是您上传的资源，您没有权限删除", c)
			return
		}
		if exist && oldDataLink == data.DataLink {
			response.FailWithMessage("没有进行修改或已存在该键", c)
			return
		}
		err = global.GLOBAL_REDIS.HDel(ctx, redisRoad, data.OldDataName).Err()
		if err != nil {
			zap.L().Error("将文件名称链接从redis中删除失败", zap.Error(err))
			response.FailWithMessage("参数错误", c)
			return
		}
		err = global.GLOBAL_REDIS.HSet(ctx, redisRoad, data.DataName, data.DataLink).Err()
		if err != nil {
			zap.L().Error("将文件名称链接存储进redis中失败", zap.Error(err))
			response.FailWithMessage("参数错误", c)
			return
		}
	}
	response.OkWithMessage("成功", c)
}

//查询资源
func GetData(c *gin.Context) {
	UserId, err := global.GetCurrentUserId(c)
	if err != nil {
		zap.L().Error("token获取userid失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	var data *Data
	err = c.ShouldBindJSON(&data)
	if err != nil {
		zap.L().Error("JSON绑定错误", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	ctx := context.Background()
	redisRoad := "learningData:"
	if data.Type == 1 {
		redisRoad += "public*"
	} else if data.Type == 2 {
		redisRoad += "dept:" + strconv.Itoa(data.DeptId) + ":*"
	} else if data.Type == 3 {
		redisRoad += "personal:" + UserId + ":*"
	}
	//result, err := global.GLOBAL_REDIS.Keys(context.Background(), "learningData:dept:546623914*").Result()
	fmt.Println(redisRoad)
	allRedisRoad, err := global.GLOBAL_REDIS.Keys(context.Background(), redisRoad).Result()
	if err != nil {
		zap.L().Error("从redis读取公共数据失败", zap.Error(err))
		response.FailWithMessage("参数错误", c)
		return
	}
	//var AllDatas []map[string]map[string]string
	var AllDatas []dingding.Result
	for _, s := range allRedisRoad {
		split := strings.Split(s, ":")
		userId := split[len(split)-1-1]
		user, err := (&dingding.DingUser{UserId: userId}).GetUserByUserId()
		AllData, err := global.GLOBAL_REDIS.HGetAll(ctx, s).Result()
		if err != nil {
			zap.L().Error("从redis读取失败", zap.Error(err))
			response.FailWithMessage("参数错误", c)
			return
		}

		//AllDatas = append(AllDatas, AllData)
		//userData := make(map[string]map[string]string)
		//userData[user.Name] = AllData

		//AllDatas = append(AllDatas, userData)

		for dataName, dataLink := range AllData {
			r := dingding.Result{
				Name:     user.Name,
				DataName: dataName,
				DataLink: dataLink,
			}
			AllDatas = append(AllDatas, r)
		}

	}

	response.OkWithDetailed(AllDatas, "成功", c)
}
