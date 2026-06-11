package web

// 501001 => 5 代表系统错误  01 代表用户模块  001 代表在这个模块中，具体的错误
// 这个叫做业务错误码，大公司都会规定使用这种方式来返回给前端错误（json）
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
