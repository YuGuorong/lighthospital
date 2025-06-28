package controllers

import (
	"lighthospital/database"
	"lighthospital/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type PrescriptionController struct{}

func (pc *PrescriptionController) Create(c *gin.Context) {
	var prescription models.Prescription
	if err := c.ShouldBindJSON(&prescription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	session := sessions.Default(c)
	doctorID := session.Get("user_id")

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建处方失败"})
		return
	}
	defer tx.Rollback()

	now := time.Now()

	// 创建处方
	result, err := tx.Exec(`
		INSERT INTO prescriptions (patient_id, doctor_id, diagnosis, doctor_advice, total_amount, status, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		prescription.PatientID, doctorID, prescription.Diagnosis, prescription.DoctorAdvice, prescription.TotalAmount,
		"draft", prescription.Notes, now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建处方失败"})
		return
	}

	prescriptionID, _ := result.LastInsertId()

	// 创建处方明细
	for _, item := range prescription.Items {
		_, err := tx.Exec(`
			INSERT INTO prescription_items (prescription_id, medicine_id, medicine_name, specification, 
			dosage, usage, frequency, days, quantity, unit_price, total_price)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			prescriptionID, item.MedicineID, item.MedicineName, item.Specification,
			item.Dosage, item.Usage, item.Frequency, item.Days, item.Quantity, item.UnitPrice, item.TotalPrice)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建处方明细失败"})
			return
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存处方失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "处方创建成功",
		"id":      prescriptionID,
	})
}

func (pc *PrescriptionController) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的处方ID"})
		return
	}

	// 查询处方基本信息
	var prescription models.Prescription
	var doctorName string
	err = database.DB.QueryRow(`
		SELECT p.id, p.patient_id, p.doctor_id, p.diagnosis, p.doctor_advice, p.total_amount, p.status, p.notes, p.created_at, p.updated_at,
		       u.name as doctor_name
		FROM prescriptions p
		LEFT JOIN users u ON p.doctor_id = u.id
		WHERE p.id = ?`, id).Scan(
		&prescription.ID, &prescription.PatientID, &prescription.DoctorID, &prescription.Diagnosis, &prescription.DoctorAdvice,
		&prescription.TotalAmount, &prescription.Status, &prescription.Notes, &prescription.CreatedAt, &prescription.UpdatedAt,
		&doctorName)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "处方不存在"})
		return
	}

	// 初始化医生信息
	if doctorName != "" {
		prescription.Doctor = &models.User{
			ID:   prescription.DoctorID,
			Name: doctorName,
		}
	} else {
		prescription.Doctor = &models.User{
			ID:   prescription.DoctorID,
			Name: "未知医生",
		}
	}

	// 查询患者信息
	var patient models.Patient
	err = database.DB.QueryRow(`
		SELECT id, name, gender, age, phone, address, id_card, medical_history, created_at, updated_at
		FROM patients WHERE id = ?`, prescription.PatientID).Scan(
		&patient.ID, &patient.Name, &patient.Gender, &patient.Age, &patient.Phone,
		&patient.Address, &patient.IDCard, &patient.MedicalHistory, &patient.CreatedAt, &patient.UpdatedAt)

	if err != nil {
		// 如果患者不存在，创建一个空的患者信息
		patient = models.Patient{
			ID:   prescription.PatientID,
			Name: "未知患者",
		}
	}
	prescription.Patient = &patient

	// 查询处方明细
	rows, err := database.DB.Query(`
		SELECT id, prescription_id, medicine_id, medicine_name, specification, dosage, usage, frequency, days, quantity, unit_price, total_price
		FROM prescription_items WHERE prescription_id = ? ORDER BY id`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询处方明细失败"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item models.PrescriptionItem
		err := rows.Scan(
			&item.ID, &item.PrescriptionID, &item.MedicineID, &item.MedicineName, &item.Specification,
			&item.Dosage, &item.Usage, &item.Frequency, &item.Days, &item.Quantity, &item.UnitPrice, &item.TotalPrice)
		if err != nil {
			continue
		}
		prescription.Items = append(prescription.Items, item)
	}

	c.JSON(http.StatusOK, gin.H{"prescription": prescription})
}

func (pc *PrescriptionController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的处方ID"})
		return
	}

	var prescription models.Prescription
	if err := c.ShouldBindJSON(&prescription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新处方失败"})
		return
	}
	defer tx.Rollback()

	// 更新处方基本信息
	_, err = tx.Exec(`
		UPDATE prescriptions SET diagnosis = ?, doctor_advice = ?, total_amount = ?, notes = ?, updated_at = ? WHERE id = ?`,
		prescription.Diagnosis, prescription.DoctorAdvice, prescription.TotalAmount, prescription.Notes, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新处方失败"})
		return
	}

	// 删除原有明细
	_, err = tx.Exec("DELETE FROM prescription_items WHERE prescription_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新处方明细失败"})
		return
	}

	// 重新创建明细
	for _, item := range prescription.Items {
		_, err := tx.Exec(`
			INSERT INTO prescription_items (prescription_id, medicine_id, medicine_name, specification, 
			dosage, usage, frequency, days, quantity, unit_price, total_price)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, item.MedicineID, item.MedicineName, item.Specification,
			item.Dosage, item.Usage, item.Frequency, item.Days, item.Quantity, item.UnitPrice, item.TotalPrice)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新处方明细失败"})
			return
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新处方失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "处方更新成功"})
}

func (pc *PrescriptionController) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的处方ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	_, err = database.DB.Exec("UPDATE prescriptions SET status = ?, updated_at = ? WHERE id = ?",
		req.Status, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新处方状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "处方状态更新成功"})
}

func (pc *PrescriptionController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的处方ID"})
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除处方失败"})
		return
	}
	defer tx.Rollback()

	// 删除处方明细
	_, err = tx.Exec("DELETE FROM prescription_items WHERE prescription_id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除处方明细失败"})
		return
	}

	// 删除处方
	_, err = tx.Exec("DELETE FROM prescriptions WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除处方失败"})
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除处方失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "处方删除成功"})
}

func (pc *PrescriptionController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	status := c.Query("status")

	offset := (page - 1) * limit

	var query string
	var args []interface{}

	whereClause := "WHERE 1=1"
	if search != "" {
		whereClause += " AND pt.name LIKE ?"
		args = append(args, "%"+search+"%")
	}
	if status != "" {
		whereClause += " AND p.status = ?"
		args = append(args, status)
	}

	query = `
		SELECT p.id, p.patient_id, p.doctor_id, p.diagnosis, p.total_amount, p.status, p.notes, p.created_at, p.updated_at,
		       pt.name as patient_name, u.name as doctor_name
		FROM prescriptions p
		LEFT JOIN patients pt ON p.patient_id = pt.id
		LEFT JOIN users u ON p.doctor_id = u.id
		` + whereClause + ` ORDER BY p.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询处方列表失败"})
		return
	}
	defer rows.Close()

	var prescriptions []models.Prescription
	for rows.Next() {
		var prescription models.Prescription
		var patientName, doctorName string
		err := rows.Scan(
			&prescription.ID, &prescription.PatientID, &prescription.DoctorID, &prescription.Diagnosis,
			&prescription.TotalAmount, &prescription.Status, &prescription.Notes, &prescription.CreatedAt, &prescription.UpdatedAt,
			&patientName, &doctorName)
		if err != nil {
			continue
		}

		prescription.Patient = &models.Patient{Name: patientName}
		prescription.Doctor = &models.User{Name: doctorName}
		prescriptions = append(prescriptions, prescription)
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM prescriptions p " + whereClause
	countArgs := args[:len(args)-2] // 去掉 LIMIT 和 OFFSET 参数
	var total int
	database.DB.QueryRow(countQuery, countArgs...).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"prescriptions": prescriptions,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

func (pc *PrescriptionController) Search(c *gin.Context) {
	var search models.PrescriptionSearch
	if err := c.ShouldBindJSON(&search); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	query := `
		SELECT p.id, p.patient_id, p.doctor_id, p.diagnosis, p.total_amount, p.status, p.notes, p.created_at, p.updated_at,
		       pt.name as patient_name, u.name as doctor_name
		FROM prescriptions p
		LEFT JOIN patients pt ON p.patient_id = pt.id
		LEFT JOIN users u ON p.doctor_id = u.id
		WHERE 1=1`
	var args []interface{}

	if search.PatientName != "" {
		query += " AND pt.name LIKE ?"
		args = append(args, "%"+search.PatientName+"%")
	}
	if search.DoctorName != "" {
		query += " AND u.name LIKE ?"
		args = append(args, "%"+search.DoctorName+"%")
	}
	if search.Status != "" {
		query += " AND p.status = ?"
		args = append(args, search.Status)
	}
	if !search.StartDate.IsZero() {
		query += " AND p.created_at >= ?"
		args = append(args, search.StartDate)
	}
	if !search.EndDate.IsZero() {
		query += " AND p.created_at <= ?"
		args = append(args, search.EndDate)
	}

	query += " ORDER BY p.created_at DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索处方失败"})
		return
	}
	defer rows.Close()

	var prescriptions []models.Prescription
	for rows.Next() {
		var prescription models.Prescription
		var patientName, doctorName string
		err := rows.Scan(
			&prescription.ID, &prescription.PatientID, &prescription.DoctorID, &prescription.Diagnosis,
			&prescription.TotalAmount, &prescription.Status, &prescription.Notes, &prescription.CreatedAt, &prescription.UpdatedAt,
			&patientName, &doctorName)
		if err != nil {
			continue
		}

		prescription.Patient = &models.Patient{Name: patientName}
		prescription.Doctor = &models.User{Name: doctorName}
		prescriptions = append(prescriptions, prescription)
	}

	c.JSON(http.StatusOK, gin.H{"prescriptions": prescriptions})
}
