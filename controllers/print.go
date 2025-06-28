package controllers

import (
	"lighthospital/database"
	"lighthospital/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

type PrintController struct{}

func (pc *PrintController) PrintPrescription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的处方ID"})
		return
	}

	// 查询处方信息
	var prescription models.Prescription
	err = database.DB.QueryRow(`
		SELECT p.id, p.patient_id, p.doctor_id, p.diagnosis, p.total_amount, p.status, p.notes, p.created_at, p.updated_at,
		       u.name as doctor_name
		FROM prescriptions p
		LEFT JOIN users u ON p.doctor_id = u.id
		WHERE p.id = ?`, id).Scan(
		&prescription.ID, &prescription.PatientID, &prescription.DoctorID, &prescription.Diagnosis,
		&prescription.TotalAmount, &prescription.Status, &prescription.Notes, &prescription.CreatedAt, &prescription.UpdatedAt,
		&prescription.Doctor.Name)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "处方不存在"})
		return
	}

	// 查询患者信息
	var patient models.Patient
	err = database.DB.QueryRow(`
		SELECT id, name, gender, age, phone, address, id_card, medical_history, created_at, updated_at
		FROM patients WHERE id = ?`, prescription.PatientID).Scan(
		&patient.ID, &patient.Name, &patient.Gender, &patient.Age, &patient.Phone,
		&patient.Address, &patient.IDCard, &patient.MedicalHistory, &patient.CreatedAt, &patient.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "患者不存在"})
		return
	}

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

	// 生成PDF
	pdf := generatePrescriptionPDF(prescription, patient)

	// 设置响应头
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=prescription_"+strconv.Itoa(id)+".pdf")

	// 输出PDF
	err = pdf.Output(c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成PDF失败"})
		return
	}
}

func generatePrescriptionPDF(prescription models.Prescription, patient models.Patient) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	// 标题
	pdf.Cell(0, 10, "处方笺")
	pdf.Ln(15)

	// 患者信息
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "患者信息")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, "姓名: "+patient.Name)
	pdf.Cell(40, 6, "性别: "+patient.Gender)
	pdf.Cell(40, 6, "年龄: "+strconv.Itoa(patient.Age))
	pdf.Ln(8)

	pdf.Cell(40, 6, "电话: "+patient.Phone)
	pdf.Cell(40, 6, "身份证: "+patient.IDCard)
	pdf.Ln(8)

	pdf.Cell(0, 6, "地址: "+patient.Address)
	pdf.Ln(10)

	// 诊断信息
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "诊断")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, prescription.Diagnosis)
	pdf.Ln(10)

	// 药品明细
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "药品明细")
	pdf.Ln(10)

	// 表头
	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(40, 6, "药品名称")
	pdf.Cell(30, 6, "规格")
	pdf.Cell(20, 6, "用法")
	pdf.Cell(20, 6, "频次")
	pdf.Cell(15, 6, "天数")
	pdf.Cell(15, 6, "数量")
	pdf.Cell(20, 6, "单价")
	pdf.Cell(20, 6, "金额")
	pdf.Ln(6)

	// 药品明细
	pdf.SetFont("Arial", "", 9)
	for _, item := range prescription.Items {
		pdf.Cell(40, 6, item.MedicineName)
		pdf.Cell(30, 6, item.Specification)
		pdf.Cell(20, 6, item.Usage)
		pdf.Cell(20, 6, item.Frequency)
		pdf.Cell(15, 6, strconv.Itoa(item.Days))
		pdf.Cell(15, 6, strconv.Itoa(item.Quantity))
		pdf.Cell(20, 6, "¥"+strconv.FormatFloat(item.UnitPrice, 'f', 2, 64))
		pdf.Cell(20, 6, "¥"+strconv.FormatFloat(item.TotalPrice, 'f', 2, 64))
		pdf.Ln(6)
	}

	pdf.Ln(5)

	// 总计
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(0, 8, "总计: ¥"+strconv.FormatFloat(prescription.TotalAmount, 'f', 2, 64))
	pdf.Ln(15)

	// 医生信息
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, "医生: "+prescription.Doctor.Name)
	pdf.Cell(40, 6, "日期: "+prescription.CreatedAt.Format("2006-01-02"))
	pdf.Ln(10)

	// 备注
	if prescription.Notes != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 8, "备注")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 6, prescription.Notes)
		pdf.Ln(10)
	}

	return pdf
}

func (pc *PrintController) PrintAppointment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的预约ID"})
		return
	}

	// 查询预约信息
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

	// 生成PDF
	pdf := generateAppointmentPDF(appointment)

	// 设置响应头
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=appointment_"+strconv.Itoa(id)+".pdf")

	// 输出PDF
	err = pdf.Output(c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成PDF失败"})
		return
	}
}

func generateAppointmentPDF(appointment models.Appointment) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	// 标题
	pdf.Cell(0, 10, "预约单")
	pdf.Ln(15)

	// 预约信息
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "预约信息")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, "患者: "+appointment.Patient.Name)
	pdf.Cell(40, 6, "医生: "+appointment.Doctor.Name)
	pdf.Ln(8)

	pdf.Cell(40, 6, "预约时间: "+appointment.AppointmentTime.Format("2006-01-02 15:04"))
	pdf.Cell(40, 6, "时长: "+strconv.Itoa(appointment.Duration)+"分钟")
	pdf.Ln(8)

	pdf.Cell(40, 6, "状态: "+appointment.Status)
	pdf.Ln(10)

	// 备注
	if appointment.Notes != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 8, "备注")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 6, appointment.Notes)
		pdf.Ln(10)
	}

	return pdf
}
