package redis

//ding:activeTask:id
func GetTaskKey(task_id string) string {
	return Perfix + ActiveTask + task_id
}

