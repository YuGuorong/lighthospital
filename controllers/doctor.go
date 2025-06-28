package controllers

import (
	"lighthospital/database"
	"lighthospital/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type DoctorController struct{}

// 管理员权限校验（假设已在路由组加中间件）

// 列表
func (dc *DoctorController) List(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, username, name FROM users WHERE role = 'doctor' ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询医生列表失败"})
		return
	}
	defer rows.Close()
	var doctors []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Name); err == nil {
			doctors = append(doctors, u)
		}
	}
	c.JSON(http.StatusOK, gin.H{"doctors": doctors})
}

// 单个
func (dc *DoctorController) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var u models.User
	err := database.DB.QueryRow("SELECT id, username, name FROM users WHERE id = ? AND role = 'doctor'", id).Scan(&u.ID, &u.Username, &u.Name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "医生不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"doctor": u})
}

// 新增
func (dc *DoctorController) Create(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Name == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	// 检查用户名唯一
	var exists int
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", req.Username).Scan(&exists)
	if exists > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	_, err := database.DB.Exec("INSERT INTO users (username, name, password, role) VALUES (?, ?, ?, 'doctor')", req.Username, req.Name, string(hash))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "添加成功"})
}

// 修改
func (dc *DoctorController) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	// 检查用户名唯一（排除自己）
	var exists int
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? AND id != ?", req.Username, id).Scan(&exists)
	if exists > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}
	var err error
	if req.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		_, err = database.DB.Exec("UPDATE users SET username = ?, name = ?, password = ? WHERE id = ? AND role = 'doctor'", req.Username, req.Name, string(hash), id)
	} else {
		_, err = database.DB.Exec("UPDATE users SET username = ?, name = ? WHERE id = ? AND role = 'doctor'", req.Username, req.Name, id)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// 删除
func (dc *DoctorController) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, err := database.DB.Exec("DELETE FROM users WHERE id = ? AND role = 'doctor'", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 重置密码
func (dc *DoctorController) ResetPassword(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	_, err := database.DB.Exec("UPDATE users SET password = ? WHERE id = ? AND role = 'doctor'", string(hash), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功"})
}
