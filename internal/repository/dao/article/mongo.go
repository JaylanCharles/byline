package article

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDBDAO struct {
	client *mongo.Client
	// 代表 byline 的
	database *mongo.Database
	// 代表的是制作库
	col *mongo.Collection
	// 代表的是线上库
	liveCol *mongo.Collection
	// 雪花节点
	node *snowflake.Node
}

func NewMongoDBDAO(db *mongo.Database, node *snowflake.Node) ArticleDAO {
	return &MongoDBDAO{
		col:     db.Collection("articles"),
		liveCol: db.Collection("published_articles"),
		node:    node,
	}
}

func (m *MongoDBDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) GetById(ctx context.Context, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	//id := m.idGen()
	id := m.node.Generate().Int64()
	art.Id = id
	_, err := m.col.InsertOne(ctx, art)
	// 你没有自增主键
	// GLOBAL UNIFY ID (GUID，全局唯一ID）
	return id, err
}

func (m *MongoDBDAO) UpdateById(ctx context.Context, art Article) error {
	// 操作制作库
	filter := bson.M{"id": art.Id, "author_id": art.AuthorId}
	update := bson.D{bson.E{Key: "$set", Value: bson.M{
		"title":   art.Title,
		"content": art.Content,
		"utime":   time.Now().UnixMilli(),
		"status":  art.Status,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// 这边就是校验了 author_id 是不是正确的 ID
	if res.ModifiedCount == 0 {
		return errors.New("更新数据失败")
	}

	return nil
}

func (m *MongoDBDAO) Sync(ctx context.Context, art Article) (int64, error) {
	// NOTE: 没法子引入事务的概念
	// 1. 保存制作库
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id

	// 2. 操作线上库了。upsert 语义
	now := time.Now().UnixMilli()
	art.Utime = now
	update := bson.E{Key: "$set", Value: art}
	upsert := bson.E{Key: "$setOnInsert", Value: bson.D{bson.E{Key: "ctime", Value: now}}}

	filter := bson.M{"id": art.Id}
	_, err = m.liveCol.UpdateOne(ctx, filter, bson.D{update, upsert}, options.UpdateOne().SetUpsert(true))

	return id, err
}

func (m *MongoDBDAO) SyncStatus(ctx context.Context, author, id int64, status uint8) error {
	panic("implement me")
}

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	index := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{Key: "author_id", Value: 1},
				bson.E{Key: "ctime", Value: 1},
			},
			Options: options.Index(),
		},
	}
	_, err := db.Collection("articles").Indexes().
		CreateMany(ctx, index)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().
		CreateMany(ctx, index)
	return err
}
