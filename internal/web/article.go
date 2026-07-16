package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	intrv1 "github.com/JaylanCharles/byline/api/proto/gen/intr/v1"
	"github.com/JaylanCharles/byline/internal/domain"
	"github.com/JaylanCharles/byline/internal/service"
	ijwt "github.com/JaylanCharles/byline/internal/web/jwt"
	"github.com/JaylanCharles/byline/pkg/ginx"
	"github.com/JaylanCharles/byline/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc     service.ArticleService
	intrSvc intrv1.InteractiveServiceClient
	l       logger.Logger
	biz     string
}

func NewArticleHandler(svc service.ArticleService, l logger.Logger, intrSvc intrv1.InteractiveServiceClient) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		l:       l,
		biz:     "article",
		intrSvc: intrSvc,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles") // 就是习惯使用复数了，也算不上好的实践
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)
	g.POST("/list", ginx.WrapBodyAndToken[ListReq, ijwt.UserClaims](h.List))
	g.GET("/detail/:id", ginx.WrapToken[ijwt.UserClaims](h.Detail))

	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	// 点赞和取消点赞都是这个接口
	pub.POST("/like", ginx.WrapBodyAndToken[LikeReq, ijwt.UserClaims](h.Like))
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 参数校验，这里不写了

	c := ctx.MustGet("claims")
	claims, ok := c.(ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}

	//调用svc
	id, err := h.svc.Save(ctx, req.toDomain(claims.Uid))

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 这里应该打日志
		h.l.Error("报错帖子失败", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: id,
	})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 参数校验，这里不写了

	c := ctx.MustGet("claims")
	claims, ok := c.(ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}

	//调用svc
	id, err := h.svc.Publish(ctx, req.toDomain(claims.Uid))

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 这里应该打日志
		h.l.Error("发表帖子失败", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: id,
	})
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 参数校验，这里不写了

	c := ctx.MustGet("claims")
	claims, ok := c.(ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}

	//调用svc
	err := h.svc.Withdraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 这里应该打日志
		h.l.Error("隐藏帖子失败", logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: req.Id,
	})
}

func (h *ArticleHandler) List(ctx *gin.Context, req ListReq, uc ijwt.UserClaims) (ginx.Result, error) {
	res, err := h.svc.List(ctx, uc.Uid, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	// 在列表页，不显示全文，只显示一个"摘要"
	// 比如说，简单的摘要就是前几句话
	// 强大的摘要是 AI 帮你生成的
	return ginx.Result{
		Data: slice.Map[domain.Article, ArticleVO](res,
			func(idx int, src domain.Article) ArticleVO {
				return ArticleVO{
					Id:       src.Id,
					Title:    src.Title,
					Abstract: src.Abstract(),
					Status:   src.Status.ToUint8(),
					// 这个列表请求，不需要返回内容
					//Content: src.Content,
					// 这个是创作者看自己的文章列表，也不需要这个字段
					//Author: src.Author
					Ctime: src.Ctime.Format(time.DateTime),
					Utime: src.Utime.Format(time.DateTime),
				}
			}),
	}, nil
}

func (h *ArticleHandler) Detail(ctx *gin.Context, usr ijwt.UserClaims) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		//ctx.JSON(http.StatusOK, )
		//a.l.Error("前端输入的 ID 不对", logger.Error(err))
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, err
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		//ctx.JSON(http.StatusOK, )
		//a.l.Error("获得文章信息失败", logger.Error(err))
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	// 这是不借助数据库查询来判定的方法
	if art.Author.Id != usr.Uid {
		return ginx.Result{
			Code: 4,
			// 也不需要告诉前端究竟发生了什么
			Msg: "输入有误",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", usr.Uid)
	}
	return ginx.Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,
			// 不需要这个摘要信息
			//Abstract: art.Abstract(),
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 这个是创作者看自己的文章列表，也不需要这个字段
			//Author: art.Author
			Ctime: art.Ctime.Format(time.DateTime),
			Utime: art.Utime.Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		h.l.Error("前端输入的 ID 不对", logger.Error(err))
		return
	}

	uc := ctx.MustGet("claims").(ijwt.UserClaims)

	var eg errgroup.Group
	var art domain.Article
	eg.Go(func() error {
		art, err = h.svc.GetPublishedById(ctx, id, uc.Uid)
		return err
	})

	var getResp *intrv1.GetResponse
	eg.Go(func() error {
		// 要在这里获得这篇文章的计数
		// 这个地方可以容忍错误
		getResp, err = h.intrSvc.Get(ctx, &intrv1.GetRequest{
			Biz: h.biz, BizId: id, Uid: uc.Uid,
		})
		// 这种是容错的写法
		//if err != nil {
		//	// 记录日志
		//}
		//return nil
		return err
	})

	// 在这儿等，要保证前面两个
	err = eg.Wait()
	if err != nil {
		// 代表查询出错了
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	// 增加阅读计数。
	go func() {
		// 开一个 goroutine，异步去执行
		_, er := h.intrSvc.IncrReadCnt(ctx, &intrv1.IncrReadCntRequest{
			Biz: h.biz, BizId: art.Id,
		})
		if er != nil {
			h.l.Error("增加阅读计数失败",
				logger.Int64("aid", art.Id),
				logger.Error(err))
		}
	}()

	intr := getResp.Intr

	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 要把作者信息带出去
			Author:     art.Author.Name,
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
			Liked:      intr.Liked,
			Collected:  intr.Collected,
			LikeCnt:    intr.LikeCnt,
			ReadCnt:    intr.ReadCnt,
			CollectCnt: intr.CollectCnt,
		},
	})
}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc ijwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		_, err = h.intrSvc.Like(ctx, &intrv1.LikeRequest{
			Biz: h.biz, BizId: req.Id, Uid: uc.Uid,
		})
	} else {
		_, err = h.intrSvc.CancelLike(ctx, &intrv1.CancelLikeRequest{
			Biz: h.biz, BizId: req.Id, Uid: uc.Uid,
		})
	}
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{Msg: "OK"}, nil
}
