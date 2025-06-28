package middleware

import (
	"lighthospital/database"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func OperationLogger(action, module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行请求
		c.Next()

		// 记录操作日志
		session := sessions.Default(c)
		userID := session.Get("user_id")
		username := session.Get("username")

		if userID != nil {
			go func() {
				_, err := database.DB.Exec(`
					INSERT INTO operation_logs (user_id, username, action, module, description, ip, user_agent, created_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
					userID, username, action, module, "", c.ClientIP(), c.Request.UserAgent(), time.Now())
				if err != nil {
					// 日志记录失败不影响主流程
					return
				}
			}()
		}
	}
}
