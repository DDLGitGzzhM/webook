package startup

import (
	"log"
	"time"

	"github.com/olivere/elastic/v7"

	"webook/webook/search/repository/dao"
)

// InitESClient 集成测试用 ES 客户端。
func InitESClient() *elastic.Client {
	const timeout = 10 * time.Second
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
		elastic.SetHealthcheckTimeoutStartup(timeout),
		elastic.SetTraceLog(log.Default()),
	}
	client, err := elastic.NewClient(opts...)
	if err != nil {
		panic(err)
	}
	err = dao.InitES(client)
	if err != nil {
		panic(err)
	}
	return client
}
