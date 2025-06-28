package database

import (
	"database/sql"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite", "./clinic.db")
	if err != nil {
		log.Fatal(err)
	}

	// 创建表
	createTables()

	// 执行数据库迁移
	migrateDatabase()

	// 初始化默认数据
	initDefaultData()
}

func createTables() {
	// 用户表
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		name TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'doctor',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 患者表
	createPatientsTable := `
	CREATE TABLE IF NOT EXISTS patients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		gender TEXT NOT NULL,
		age INTEGER NOT NULL,
		phone TEXT,
		address TEXT,
		id_card TEXT,
		medical_history TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 药品表
	createMedicinesTable := `
	CREATE TABLE IF NOT EXISTS medicines (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		specification TEXT NOT NULL,
		unit TEXT NOT NULL,
		price REAL NOT NULL DEFAULT 0,
		stock INTEGER NOT NULL DEFAULT 0,
		min_stock INTEGER NOT NULL DEFAULT 0,
		category TEXT,
		manufacturer TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 处方表
	createPrescriptionsTable := `
	CREATE TABLE IF NOT EXISTS prescriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER NOT NULL,
		doctor_id INTEGER NOT NULL,
		diagnosis TEXT,
		doctor_advice TEXT,
		total_amount REAL NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'draft',
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (patient_id) REFERENCES patients (id),
		FOREIGN KEY (doctor_id) REFERENCES users (id)
	);`

	// 处方明细表
	createPrescriptionItemsTable := `
	CREATE TABLE IF NOT EXISTS prescription_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		prescription_id INTEGER NOT NULL,
		medicine_id INTEGER,
		medicine_name TEXT NOT NULL,
		specification TEXT NOT NULL,
		dosage TEXT NOT NULL,
		usage TEXT NOT NULL,
		frequency TEXT NOT NULL,
		days INTEGER NOT NULL DEFAULT 1,
		quantity INTEGER NOT NULL DEFAULT 1,
		unit_price REAL NOT NULL DEFAULT 0,
		total_price REAL NOT NULL DEFAULT 0,
		FOREIGN KEY (prescription_id) REFERENCES prescriptions (id),
		FOREIGN KEY (medicine_id) REFERENCES medicines (id)
	);`

	// 预约表
	createAppointmentsTable := `
	CREATE TABLE IF NOT EXISTS appointments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER NOT NULL,
		doctor_id INTEGER NOT NULL,
		appointment_time DATETIME NOT NULL,
		duration INTEGER NOT NULL DEFAULT 30,
		status TEXT NOT NULL DEFAULT 'scheduled',
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (patient_id) REFERENCES patients (id),
		FOREIGN KEY (doctor_id) REFERENCES users (id)
	);`

	// 操作日志表
	createOperationLogsTable := `
	CREATE TABLE IF NOT EXISTS operation_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		username TEXT NOT NULL,
		action TEXT NOT NULL,
		module TEXT NOT NULL,
		description TEXT,
		ip TEXT,
		user_agent TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id)
	);`

	tables := []string{
		createUsersTable,
		createPatientsTable,
		createMedicinesTable,
		createPrescriptionsTable,
		createPrescriptionItemsTable,
		createAppointmentsTable,
		createOperationLogsTable,
	}

	for _, table := range tables {
		_, err := DB.Exec(table)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func initDefaultData() {
	// 检查是否已有管理员用户
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		// 创建默认管理员用户
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		_, err := DB.Exec(`
			INSERT INTO users (username, password, name, role, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			"admin", string(hashedPassword), "系统管理员", "admin", time.Now(), time.Now())
		if err != nil {
			log.Fatal(err)
		}
		log.Println("默认管理员用户已创建: admin/admin123")
	}

	// 检查是否已有默认医生用户
	err = DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'doctor'").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		// 创建默认医生用户
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("doctor123"), bcrypt.DefaultCost)
		_, err := DB.Exec(`
			INSERT INTO users (username, password, name, role, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			"doctor", string(hashedPassword), "张医生", "doctor", time.Now(), time.Now())
		if err != nil {
			log.Fatal(err)
		}
		log.Println("默认医生用户已创建: doctor/doctor123")
	}

	// 添加一些示例药品
	var medicineCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM medicines").Scan(&medicineCount)
	if err != nil {
		log.Fatal(err)
	}

	if medicineCount == 0 {
		sampleMedicines := []struct {
			name, spec, unit, category, manufacturer string
			price                                    float64
			stock                                    int
		}{
			{"阿莫西林胶囊", "0.25g*24粒", "盒", "抗生素", "华北制药", 15.50, 100},
			{"布洛芬片", "0.1g*20片", "盒", "解热镇痛", "中美史克", 8.80, 80},
			{"感冒灵颗粒", "10g*10袋", "盒", "感冒药", "999药业", 12.00, 60},
			{"维生素C片", "0.1g*100片", "瓶", "维生素", "东北制药", 5.50, 120},
			{"板蓝根颗粒", "10g*20袋", "盒", "清热解毒", "白云山", 18.00, 50},
		}

		for _, med := range sampleMedicines {
			_, err := DB.Exec(`
				INSERT INTO medicines (name, specification, unit, price, stock, min_stock, category, manufacturer, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				med.name, med.spec, med.unit, med.price, med.stock, 10, med.category, med.manufacturer, time.Now(), time.Now())
			if err != nil {
				log.Printf("添加示例药品失败: %v", err)
			}
		}
		log.Println("示例药品数据已添加")
	}
}

func migrateDatabase() {
	// 检查prescriptions表是否有doctor_advice字段
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('prescriptions') 
		WHERE name = 'doctor_advice'`).Scan(&count)

	if err != nil {
		log.Printf("检查字段失败: %v", err)
		return
	}

	// 如果字段不存在，则添加
	if count == 0 {
		_, err := DB.Exec(`ALTER TABLE prescriptions ADD COLUMN doctor_advice TEXT;`)
		if err != nil {
			log.Printf("添加doctor_advice字段失败: %v", err)
		} else {
			log.Println("prescriptions表已添加doctor_advice字段")
		}
	} else {
		log.Println("doctor_advice字段已存在")
	}
}
