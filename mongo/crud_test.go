package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestMongoDB(t *testing.T) {
	// 初始化客户端
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			t.Log(evt.Command)
		},
	}
	opts := options.Client().ApplyURI("mongodb://root:root@localhost:27017/").SetMonitor(monitor)
	client, err := mongo.Connect(opts)
	assert.NoError(t, err)

	col := client.Database("byline").Collection("articles")

	// 插入数据
	insertRes, err := col.InsertOne(ctx, Article{
		Id:       1,
		Title:    "我的标题",
		Content:  "我的内容",
		AuthorId: 123,
	})
	assert.NoError(t, err)

	oid := insertRes.InsertedID.(bson.ObjectID)
	t.Log("插入ID", oid)

	// 查找数据
	var art Article
	filter := bson.D{bson.E{Key: "id", Value: 1}}
	err = col.FindOne(ctx, filter).Decode(&art)
	assert.NoError(t, err)
	t.Log(art)

	art = Article{}
	err = col.FindOne(ctx, Article{
		Id: 1,
	}).Decode(&art)
	assert.NoError(t, err)
	t.Log(art)

	// 更新数据
	set := bson.D{bson.E{Key: "$set", Value: bson.E{Key: "title", Value: "新的标题"}}}

	updateRes, err := col.UpdateOne(ctx, filter, set)
	assert.NoError(t, err)
	t.Log("更新文档数量", updateRes.ModifiedCount)

	updateManyRes, err := col.UpdateMany(ctx, filter,
		bson.D{bson.E{Key: "$set",
			Value: Article{Content: "新的内容"}}})
	assert.NoError(t, err)
	t.Log("更新文档数量", updateManyRes.ModifiedCount)

	// 创建索引
	indexRes, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{bson.E{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("idx_id"),
	})
	assert.NoError(t, err)
	t.Log("创建索引", indexRes)

	// 删除数据
	delRes, err := col.DeleteMany(ctx, filter)
	assert.NoError(t, err)
	t.Log("删除文档数量", delRes.DeletedCount)
}

type Article struct {
	Id       int64  `bson:"id,omitempty"`
	Title    string `bson:"title,omitempty"`
	Content  string `bson:"content,omitempty"`
	AuthorId int64  `bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	// 更新时间
	Utime int64 `bson:"utime,omitempty"`
}
