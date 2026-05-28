package web

import (
	"errors"
	"net/http"

	"github.com/JaylanCharles/byline/internal/domain"
	"github.com/JaylanCharles/byline/internal/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// UserHandler 用来定义与用户有关的路由
type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)

	// go 自带的 regexp 包，不支持很复杂的正则表达式，使用第三方包，然后重命名一下hhhh
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None) // 第二个参数随便传
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)

	return &UserHandler{
		emailExp:    emailExp,
		passwordExp: passwordExp,
		svc:         svc,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.Login)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.Profile)
}

// 这里入参使用ctx *gin.Context 而不是 使用server *gin.Engine
// 是因为	server.POST("/users/signup", u.SignUp) POST要求第二个入参使用type HandlerFunc func(*Context)
func (u *UserHandler) SignUp(ctx *gin.Context) {
	// 使用内部结构体 因为不想让别的方法看见这个结构体
	type SignUpReq struct {
		Email string `json:"email"`
		// 有些人感觉ConfirmPassword不用后端接收处理，前端校验就可以了，是可以的
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	// Bind 方法会根据 Content-Type 来解析你的数据到 req
	// 解析错了，就会直接写回一个 400 的错误
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		// 说明正则表达式有问题
		ctx.String(http.StatusOK, "系统错误") // 不要将具体错误返回给前端，因为这是内部错误
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "你的邮箱格式不对")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		// 记录日志
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码必须大于8位，包含数字、特殊字符")
		return
	}

	// 数据库操作
	// 调用一下 service 的方法
	// 注意 err 一定要规范处理
	err = u.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	// 使用 errors.Is() 是最佳实践
	if errors.Is(err, service.ErrUserDuplicateEmail) {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}
func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 参数少，可以不使用 domain.User
	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		// 记录日志
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 登录成功了
	sess := sessions.Default(ctx)
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		// 生产环境要将这两个参数开启
		//Secure:   true, // 规定使用 https 协议
		//HttpOnly: true, // 这个 cookie 只能被服务器通过 HTTP 请求读写，浏览器里的 JavaScript 不能访问它
		MaxAge: 300, // 单位：秒
	})
	if err := sess.Save(); err != nil { // 设置一次加一个 cookie 在浏览器端可以看见，一次 save() 增加一个 cookie
		panic(err)
	}

	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		MaxAge: -1,
	})
	sess.Save() // 设置一次加一个 cookie 在浏览器端可以看见，一次 save() 增加一个 cookie
	ctx.String(http.StatusOK, "退出登录成功")
	return
}
func (u *UserHandler) Edit(ctx *gin.Context) {

}
func (u *UserHandler) Profile(ctx *gin.Context) {

}
