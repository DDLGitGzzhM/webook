package startup

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoDB *mongo.Database

func InitMongoDB() *mongo.Database {
	if mongoDB == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		monitor := &event.CommandMonitor{
			Started: func(ctx context.Context,
				startedEvent *event.CommandStartedEvent) {
				fmt.Println(startedEvent.Command)
			},
		}
		opts := options.Client().
			ApplyURI("mongodb://root:example@localhost:27017/").
			SetMonitor(monitor)
		client, err := mongo.Connect(ctx, opts)
		if err != nil {
			panic(err)
		}
		mongoDB = client.Database("webook")
	}
	return mongoDB
}
