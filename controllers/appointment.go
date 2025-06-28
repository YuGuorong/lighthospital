package controllers

import (
	"lighthospital/database"
	"lighthospital/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AppointmentController struct{}

func (ac *AppointmentController) Create(c *gin.Context) {
	var appointment models.Appointment
	if err := c.ShouldBindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	now := time.Now()
	result, err := database.DB.Exec(`
		INSERT INTO appointments (patient_id, doctor_id, appointment_time, duration, status, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		appointment.PatientID, appointment.DoctorID, appointment.AppointmentTime, appointment.Duration,
		"scheduled", appointment.Notes, now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建预约失败"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{
		"message": "预约创建成功",
		"id":      id,
	})
}

func (ac *AppointmentController) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的预约ID"})
		return
	}

	var appointment models.Appointment
	err = database.DB.QueryRow(`
		SELECT a.id, a.patient_id, a.doctor_id, a.appointment_time, a.duration, a.status, a.notes, a.created_at, a.updated_at,
		       p.name as patient_name, u.name as doctor_name
		FROM appointments a
		LEFT JOIN patients p ON a.patient_id = p.id
		LEFT JOIN users u ON a.doctor_id = u.id
		WHERE a.id = ?`, id).Scan(
		&appointment.ID, &appointment.PatientID, &appointment.DoctorID, &appointment.AppointmentTime,
		&appointment.Duration, &appointment.Status, &appointment.Notes, &appointment.CreatedAt, &appointment.UpdatedAt,
		&appointment.Patient.Name, &appointment.Doctor.Name)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "预约不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"appointment": appointment})
}

func (ac *AppointmentController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的预约ID"})
		return
	}

	var appointment models.Appointment
	if err := c.ShouldBindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	_, err = database.DB.Exec(`
		UPDATE appointments SET patient_id = ?, doctor_id = ?, appointment_time = ?, duration = ?, 
		status = ?, notes = ?, updated_at = ? WHERE id = ?`,
		appointment.PatientID, appointment.DoctorID, appointment.AppointmentTime, appointment.Duration,
		appointment.Status, appointment.Notes, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新预约失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "预约信息更新成功"})
}

func (ac *AppointmentController) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的预约ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	_, err = database.DB.Exec("UPDATE appointments SET status = ?, updated_at = ? WHERE id = ?",
		req.Status, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新预约状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "预约状态更新成功"})
}

func (ac *AppointmentController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的预约ID"})
		return
	}

	_, err = database.DB.Exec("DELETE FROM appointments WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除预约失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "预约删除成功"})
}

func (ac *AppointmentController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	patientID := c.Query("patient_id")
	doctorID := c.Query("doctor_id")
	status := c.Query("status")
	date := c.Query("date")

	offset := (page - 1) * limit

	var query string
	var args []interface{}

	whereClause := "WHERE 1=1"
	if patientID != "" {
		whereClause += " AND a.patient_id = ?"
		args = append(args, patientID)
	}
	if doctorID != "" {
		whereClause += " AND a.doctor_id = ?"
		args = append(args, doctorID)
	}
	if status != "" {
		whereClause += " AND a.status = ?"
		args = append(args, status)
	}
	if date != "" {
		whereClause += " AND DATE(a.appointment_time) = ?"
		args = append(args, date)
	}

	query = `
		SELECT a.id, a.patient_id, a.doctor_id, a.appointment_time, a.duration, a.status, a.notes, a.created_at, a.updated_at,
		       p.name as patient_name, u.name as doctor_name
		FROM appointments a
		LEFT JOIN patients p ON a.patient_id = p.id
		LEFT JOIN users u ON a.doctor_id = u.id
		` + whereClause + ` ORDER BY a.appointment_time ASC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询预约列表失败"})
		return
	}
	defer rows.Close()

	var appointments []models.Appointment
	for rows.Next() {
		var appointment models.Appointment
		var patientName, doctorName string
		err := rows.Scan(
			&appointment.ID, &appointment.PatientID, &appointment.DoctorID, &appointment.AppointmentTime,
			&appointment.Duration, &appointment.Status, &appointment.Notes, &appointment.CreatedAt, &appointment.UpdatedAt,
			&patientName, &doctorName)
		if err != nil {
			continue
		}

		appointment.Patient = &models.Patient{Name: patientName}
		appointment.Doctor = &models.User{Name: doctorName}
		appointments = append(appointments, appointment)
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM appointments a " + whereClause
	countArgs := args[:len(args)-2] // 去掉 LIMIT 和 OFFSET 参数
	var total int
	database.DB.QueryRow(countQuery, countArgs...).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"appointments": appointments,
		"total":        total,
		"page":         page,
		"limit":        limit,
	})
}

func (ac *AppointmentController) Search(c *gin.Context) {
	var search models.AppointmentSearch
	if err := c.ShouldBindJSON(&search); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	query := `
		SELECT a.id, a.patient_id, a.doctor_id, a.appointment_time, a.duration, a.status, a.notes, a.created_at, a.updated_at,
		       p.name as patient_name, u.name as doctor_name
		FROM appointments a
		LEFT JOIN patients p ON a.patient_id = p.id
		LEFT JOIN users u ON a.doctor_id = u.id
		WHERE 1=1`
	var args []interface{}

	if search.PatientName != "" {
		query += " AND p.name LIKE ?"
		args = append(args, "%"+search.PatientName+"%")
	}
	if search.DoctorName != "" {
		query += " AND u.name LIKE ?"
		args = append(args, "%"+search.DoctorName+"%")
	}
	if search.Status != "" {
		query += " AND a.status = ?"
		args = append(args, search.Status)
	}
	if !search.StartDate.IsZero() {
		query += " AND a.appointment_time >= ?"
		args = append(args, search.StartDate)
	}
	if !search.EndDate.IsZero() {
		query += " AND a.appointment_time <= ?"
		args = append(args, search.EndDate)
	}

	query += " ORDER BY a.appointment_time ASC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索预约失败"})
		return
	}
	defer rows.Close()

	var appointments []models.Appointment
	for rows.Next() {
		var appointment models.Appointment
		var patientName, doctorName string
		err := rows.Scan(
			&appointment.ID, &appointment.PatientID, &appointment.DoctorID, &appointment.AppointmentTime,
			&appointment.Duration, &appointment.Status, &appointment.Notes, &appointment.CreatedAt, &appointment.UpdatedAt,
			&patientName, &doctorName)
		if err != nil {
			continue
		}

		appointment.Patient = &models.Patient{Name: patientName}
		appointment.Doctor = &models.User{Name: doctorName}
		appointments = append(appointments, appointment)
	}

	c.JSON(http.StatusOK, gin.H{"appointments": appointments})
}

func (ac *AppointmentController) GetTodayAppointments(c *gin.Context) {
	doctorID := c.Query("doctor_id")
	today := time.Now().Format("2006-01-02")

	var query string
	var args []interface{}

	if doctorID != "" {
		query = `
			SELECT a.id, a.patient_id, a.doctor_id, a.appointment_time, a.duration, a.status, a.notes, a.created_at, a.updated_at,
			       p.name as patient_name, u.name as doctor_name
			FROM appointments a
			LEFT JOIN patients p ON a.patient_id = p.id
			LEFT JOIN users u ON a.doctor_id = u.id
			WHERE DATE(a.appointment_time) = ? AND a.doctor_id = ?
			ORDER BY a.appointment_time ASC`
		args = []interface{}{today, doctorID}
	} else {
		query = `
			SELECT a.id, a.patient_id, a.doctor_id, a.appointment_time, a.duration, a.status, a.notes, a.created_at, a.updated_at,
			       p.name as patient_name, u.name as doctor_name
			FROM appointments a
			LEFT JOIN patients p ON a.patient_id = p.id
			LEFT JOIN users u ON a.doctor_id = u.id
			WHERE DATE(a.appointment_time) = ?
			ORDER BY a.appointment_time ASC`
		args = []interface{}{today}
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询今日预约失败"})
		return
	}
	defer rows.Close()

	var appointments []models.Appointment
	for rows.Next() {
		var appointment models.Appointment
		var patientName, doctorName string
		err := rows.Scan(
			&appointment.ID, &appointment.PatientID, &appointment.DoctorID, &appointment.AppointmentTime,
			&appointment.Duration, &appointment.Status, &appointment.Notes, &appointment.CreatedAt, &appointment.UpdatedAt,
			&patientName, &doctorName)
		if err != nil {
			continue
		}

		appointment.Patient = &models.Patient{Name: patientName}
		appointment.Doctor = &models.User{Name: doctorName}
		appointments = append(appointments, appointment)
	}

	c.JSON(http.StatusOK, gin.H{"appointments": appointments})
}
