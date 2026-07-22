package mongo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, cce *event.CommandStartedEvent) { // 每个命令执行之前
			fmt.Println(cce.Command.String())
		},
	}
	opts := options.Client().ApplyURI("mongodb://root:example@localhost:27017").
		SetMonitor(monitor)

	client, err := mongo.Connect(ctx,
		opts)
	require.Nil(t, err)

	mdb := client.Database("webook")
	col := mdb.Collection("articles")

	res, err := col.InsertOne(ctx, Article{
		Id:       1,
		Title:    "我的标题",
		Content:  "我的内容",
		AuthorId: 123,
		Ctime:    time.Now().UnixMilli(),
		Utime:    time.Now().UnixMilli(),
		Status:   1,
	})
	require.Nil(t, err)
	t.Log("insteredId", res.InsertedID) // 文档 ID

	filter := bson.D{bson.E{Key: "id", Value: 1}}
	var article Article
	err = col.FindOne(ctx, filter).Decode(&article)
	require.Nil(t, err)
	t.Log("article", article)

	sets := bson.D{bson.E{Key: "$set", Value: bson.D{{Key: "title", Value: "新的标题"}}}}
	one, err := col.UpdateOne(ctx, filter, sets)
	if err != nil {
		return
	}
	require.Nil(t, err)
	t.Log("one", one)
}

type Article struct {
	Id       int64
	Title    string
	Content  string
	AuthorId int64
	Status   uint8
	Ctime    int64
	Utime    int64
}
