package domain

import "time"

const (
	// 未知状态
	ArticleStatusUnknow ArticleStatus = iota
	// 未发表
	ArticleStatusUnpublished
	// 已发表
	ArticleStatusPublished
	// 仅自己可见
	ArticleStatusPrivate
)

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
	Ctime   time.Time
	Utime   time.Time
}

func (a Article) Abstract() string {
	str := []rune(a.Content)
	// 只取部分作为摘要
	if len(str) > 128 {
		str = str[:128]
	}
	return string(str)
}

type Author struct {
	Id   int64
	Name string
}

type ArticleStatus uint8

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

func (s ArticleStatus) Valid() bool {
	return s.ToUint8() > 0
}
