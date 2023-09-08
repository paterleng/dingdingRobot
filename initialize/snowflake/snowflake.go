package snowflake

//我们把这个东西包装成一个模块即可，我们的项目很小，
//以后多个业务都要生成id的时候，我们封装成一个服务部署后，可以共同分布式使用，那个项目用到了，就调用即可
import (
	"time"

	sf "github.com/bwmarrin/snowflake"
)

var node *sf.Node

func Init(startTime string, machineID int64) (err error) {
	var st time.Time
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		return
	}
	sf.Epoch = st.UnixNano() / 1000000
	node, err = sf.NewNode(machineID)
	return
}
func GenID() int64 {
	return node.Generate().Int64()
}
