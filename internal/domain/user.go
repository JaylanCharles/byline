package domain

// User 领域对象
// BO
type User struct {
	Email    string
	Password string
	// 不需要 comfirmPassword ，因为web层已经检验过了
}
