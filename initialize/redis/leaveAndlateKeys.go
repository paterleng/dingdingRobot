package redis

/**
*
* @author yth
* @language go
* @since 2023/2/25 21:30
 */

const (
	KeyDeptAveLeave = "leave:" // 根据各部门平均请假次数排序的集合
	KeyDeptAveLate  = "late:"  // 根据各部门平均迟到次数排序的集合
)
