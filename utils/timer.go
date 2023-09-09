package utils
import (
	"time"
)

// 定时任务，定时执行某个函数
func PeriodicRoutine(timeDur time.Duration, fu func()) {
	GlobalLimit := make(chan string, 1)
	for {
		GlobalLimit <- "s"
		time.AfterFunc(timeDur, func() {
			fu()
			<-GlobalLimit
		})
	}
}

var TimeTicker *time.Ticker
func Timing( hour *int)  {
	TimeTicker = time.NewTicker(1 * time.Hour)
	for {
		<-TimeTicker.C
		*hour--
	}
}