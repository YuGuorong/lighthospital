package controllers

import (
	"lighthospital/database"
	"lighthospital/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mozillazg/go-pinyin"
)

type PatientController struct{}

func (pc *PatientController) Create(c *gin.Context) {
	var patient models.Patient
	if err := c.ShouldBindJSON(&patient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 生成拼音
	pinyin := generatePinyin(patient.Name)

	now := time.Now()
	result, err := database.DB.Exec(`
		INSERT INTO patients (name, pinyin, gender, age, phone, address, id_card, medical_history, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		patient.Name, pinyin, patient.Gender, patient.Age, patient.Phone, patient.Address,
		patient.IDCard, patient.MedicalHistory, now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建患者失败"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{
		"message": "患者创建成功",
		"id":      id,
	})
}

func (pc *PatientController) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的患者ID"})
		return
	}

	var patient models.Patient
	err = database.DB.QueryRow(`
		SELECT id, name, pinyin, gender, age, phone, address, id_card, medical_history, created_at, updated_at
		FROM patients WHERE id = ?`, id).Scan(
		&patient.ID, &patient.Name, &patient.Pinyin, &patient.Gender, &patient.Age, &patient.Phone,
		&patient.Address, &patient.IDCard, &patient.MedicalHistory, &patient.CreatedAt, &patient.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "患者不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"patient": patient})
}

func (pc *PatientController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的患者ID"})
		return
	}

	var patient models.Patient
	if err := c.ShouldBindJSON(&patient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 生成拼音
	pinyin := generatePinyin(patient.Name)

	_, err = database.DB.Exec(`
		UPDATE patients SET name = ?, pinyin = ?, gender = ?, age = ?, phone = ?, address = ?, 
		id_card = ?, medical_history = ?, updated_at = ? WHERE id = ?`,
		patient.Name, pinyin, patient.Gender, patient.Age, patient.Phone, patient.Address,
		patient.IDCard, patient.MedicalHistory, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新患者失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "患者信息更新成功"})
}

func (pc *PatientController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的患者ID"})
		return
	}

	_, err = database.DB.Exec("DELETE FROM patients WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除患者失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "患者删除成功"})
}

func (pc *PatientController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	offset := (page - 1) * limit

	var query string
	var args []interface{}

	if search != "" {
		query = `
			SELECT id, name, pinyin, gender, age, phone, address, id_card, medical_history, created_at, updated_at
			FROM patients 
			WHERE name LIKE ? OR pinyin LIKE ? OR phone LIKE ? OR id_card LIKE ?
			ORDER BY created_at DESC LIMIT ? OFFSET ?`
		args = []interface{}{"%" + search + "%", "%" + search + "%", "%" + search + "%", "%" + search + "%", limit, offset}
	} else {
		query = `
			SELECT id, name, pinyin, gender, age, phone, address, id_card, medical_history, created_at, updated_at
			FROM patients 
			ORDER BY created_at DESC LIMIT ? OFFSET ?`
		args = []interface{}{limit, offset}
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询患者列表失败"})
		return
	}
	defer rows.Close()

	var patients []models.Patient
	for rows.Next() {
		var patient models.Patient
		err := rows.Scan(
			&patient.ID, &patient.Name, &patient.Pinyin, &patient.Gender, &patient.Age, &patient.Phone,
			&patient.Address, &patient.IDCard, &patient.MedicalHistory, &patient.CreatedAt, &patient.UpdatedAt)
		if err != nil {
			continue
		}
		patients = append(patients, patient)
	}

	// 获取总数
	var total int
	if search != "" {
		database.DB.QueryRow("SELECT COUNT(*) FROM patients WHERE name LIKE ? OR pinyin LIKE ? OR phone LIKE ? OR id_card LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%").Scan(&total)
	} else {
		database.DB.QueryRow("SELECT COUNT(*) FROM patients").Scan(&total)
	}

	c.JSON(http.StatusOK, gin.H{
		"patients": patients,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (pc *PatientController) Search(c *gin.Context) {
	var search models.PatientSearch
	if err := c.ShouldBindJSON(&search); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	query := `
		SELECT id, name, pinyin, gender, age, phone, address, id_card, medical_history, created_at, updated_at
		FROM patients WHERE 1=1`
	var args []interface{}

	if search.Name != "" {
		query += " AND (name LIKE ? OR pinyin LIKE ?)"
		args = append(args, "%"+search.Name+"%", "%"+search.Name+"%")
	}
	if search.Phone != "" {
		query += " AND phone LIKE ?"
		args = append(args, "%"+search.Phone+"%")
	}
	if search.IDCard != "" {
		query += " AND id_card LIKE ?"
		args = append(args, "%"+search.IDCard+"%")
	}

	query += " ORDER BY created_at DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索患者失败"})
		return
	}
	defer rows.Close()

	var patients []models.Patient
	for rows.Next() {
		var patient models.Patient
		err := rows.Scan(
			&patient.ID, &patient.Name, &patient.Pinyin, &patient.Gender, &patient.Age, &patient.Phone,
			&patient.Address, &patient.IDCard, &patient.MedicalHistory, &patient.CreatedAt, &patient.UpdatedAt)
		if err != nil {
			continue
		}
		patients = append(patients, patient)
	}

	c.JSON(http.StatusOK, gin.H{"patients": patients})
}

// FindOrCreateByName 根据姓名查找患者，如果不存在则创建
func (pc *PatientController) FindOrCreateByName(c *gin.Context) {
	var request struct {
		Name   string `json:"name" binding:"required"`
		Gender string `json:"gender"`
		Age    int    `json:"age"`
		Phone  string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 先查找是否存在同名患者
	var patient models.Patient
	err := database.DB.QueryRow(`
		SELECT id, name, pinyin, gender, age, phone, address, id_card, medical_history, created_at, updated_at
		FROM patients WHERE name = ?`, request.Name).Scan(
		&patient.ID, &patient.Name, &patient.Pinyin, &patient.Gender, &patient.Age, &patient.Phone,
		&patient.Address, &patient.IDCard, &patient.MedicalHistory, &patient.CreatedAt, &patient.UpdatedAt)

	if err == nil {
		// 患者已存在，返回患者信息
		c.JSON(http.StatusOK, gin.H{
			"patient": patient,
			"created": false,
			"message": "找到现有患者",
		})
		return
	}

	// 患者不存在，创建新患者
	pinyin := generatePinyin(request.Name)
	now := time.Now()
	result, err := database.DB.Exec(`
		INSERT INTO patients (name, pinyin, gender, age, phone, address, id_card, medical_history, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		request.Name, pinyin, request.Gender, request.Age, request.Phone, "", "", "", now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建患者失败"})
		return
	}

	id, _ := result.LastInsertId()

	// 返回新创建的患者信息
	newPatient := models.Patient{
		ID:        int(id),
		Name:      request.Name,
		Pinyin:    pinyin,
		Gender:    request.Gender,
		Age:       request.Age,
		Phone:     request.Phone,
		CreatedAt: now,
		UpdatedAt: now,
	}

	c.JSON(http.StatusOK, gin.H{
		"patient": newPatient,
		"created": true,
		"message": "患者创建成功",
	})
}

func generatePinyin(name string) string {
	p := pinyin.NewArgs()
	p.Style = pinyin.Normal
	pinyinArray := pinyin.Pinyin(name, p)

	var result []string
	for _, chars := range pinyinArray {
		if len(chars) > 0 {
			result = append(result, chars[0])
		}
	}
	return strings.Join(result, "")
}
