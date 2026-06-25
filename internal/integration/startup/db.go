package startup

import (
	"context"
	"fmt"
	"time"

	"github.com/JaylanCharles/byline/internal/repository/dao"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var mongoDB *mongo.Database

func InitMongoDB() *mongo.Database {
	if mongoDB == nil {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		monitor := &event.CommandMonitor{
			Started: func(ctx context.Context,
				startedEvent *event.CommandStartedEvent) {
				fmt.Println(startedEvent.Command)
			},
		}
		opts := options.Client().
			ApplyURI("mongodb://root:root@localhost:27017/").
			SetMonitor(monitor)
		client, err := mongo.Connect(opts)
		if err != nil {
			panic(err)
		}

		mongoDB = client.Database("byline")
	}
	return mongoDB
}

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/byline"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
