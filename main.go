package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"lighthospital/controllers"
	"lighthospital/database"
	"lighthospital/middleware"
)

func main() {
	// 初始化数据库
	database.InitDB()
	log.Println("数据库初始化完成")

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin实例
	r := gin.Default()

	// 设置CORS中间件
	r.Use(middleware.CORSMiddleware())

	// 设置session中间件
	store := cookie.NewStore([]byte("lighthospital_secret_key"))
	store.Options(sessions.Options{
		MaxAge:   3600 * 24, // 24小时
		Path:     "/",
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("lighthospital_session", store))

	// 静态文件服务
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// 首页
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "轻量级诊所管理系统",
		})
	})

	// 处方详情页面
	r.GET("/prescription/:id", func(c *gin.Context) {
		c.HTML(http.StatusOK, "prescription_detail.html", gin.H{
			"title": "处方详情",
		})
	})

	// API路由组
	api := r.Group("/api")
	{
		// 认证相关路由
		auth := api.Group("/auth")
		{
			authController := &controllers.AuthController{}
			auth.POST("/login", authController.Login)
			auth.POST("/logout", middleware.AuthRequired(), authController.Logout)
			auth.GET("/current", middleware.AuthRequired(), authController.GetCurrentUser)
			auth.POST("/change-password", middleware.AuthRequired(), authController.ChangePassword)
		}

		// 需要认证的路由
		authorized := api.Group("/")
		authorized.Use(middleware.AuthRequired())
		{
			// 患者管理
			patients := authorized.Group("/patients")
			{
				patientController := &controllers.PatientController{}
				patients.POST("", middleware.OperationLogger("创建", "患者"), patientController.Create)
				patients.GET("/:id", patientController.Get)
				patients.PUT("/:id", middleware.OperationLogger("更新", "患者"), patientController.Update)
				patients.DELETE("/:id", middleware.OperationLogger("删除", "患者"), patientController.Delete)
				patients.GET("", patientController.List)
				patients.POST("/search", patientController.Search)
				patients.POST("/find-or-create", middleware.OperationLogger("快速查找或创建", "患者"), patientController.FindOrCreateByName)
			}

			// 药品管理
			medicines := authorized.Group("/medicines")
			{
				medicineController := &controllers.MedicineController{}
				medicines.POST("", middleware.OperationLogger("创建", "药品"), medicineController.Create)
				medicines.GET("/:id", medicineController.Get)
				medicines.PUT("/:id", middleware.OperationLogger("更新", "药品"), medicineController.Update)
				medicines.DELETE("/:id", middleware.OperationLogger("删除", "药品"), medicineController.Delete)
				medicines.GET("", medicineController.List)
				medicines.POST("/search", medicineController.Search)
				medicines.PUT("/:id/stock", middleware.OperationLogger("更新库存", "药品"), medicineController.UpdateStock)
				medicines.GET("/categories", medicineController.GetCategories)
				medicines.GET("/low-stock", medicineController.GetLowStock)
			}

			// 处方管理
			prescriptions := authorized.Group("/prescriptions")
			{
				prescriptionController := &controllers.PrescriptionController{}
				prescriptions.POST("", middleware.OperationLogger("创建", "处方"), prescriptionController.Create)
				prescriptions.GET("/:id", prescriptionController.Get)
				prescriptions.PUT("/:id", middleware.OperationLogger("更新", "处方"), prescriptionController.Update)
				prescriptions.DELETE("/:id", middleware.OperationLogger("删除", "处方"), prescriptionController.Delete)
				prescriptions.GET("", prescriptionController.List)
				prescriptions.POST("/search", prescriptionController.Search)
				prescriptions.PUT("/:id/status", middleware.OperationLogger("更新状态", "处方"), prescriptionController.UpdateStatus)
			}

			// 预约管理
			appointments := authorized.Group("/appointments")
			{
				appointmentController := &controllers.AppointmentController{}
				appointments.POST("", middleware.OperationLogger("创建", "预约"), appointmentController.Create)
				appointments.GET("/:id", appointmentController.Get)
				appointments.PUT("/:id", middleware.OperationLogger("更新", "预约"), appointmentController.Update)
				appointments.DELETE("/:id", middleware.OperationLogger("删除", "预约"), appointmentController.Delete)
				appointments.GET("", appointmentController.List)
				appointments.POST("/search", appointmentController.Search)
				appointments.PUT("/:id/status", middleware.OperationLogger("更新状态", "预约"), appointmentController.UpdateStatus)
				appointments.GET("/today", appointmentController.GetTodayAppointments)
			}

			// 打印服务
			print := authorized.Group("/print")
			{
				printController := &controllers.PrintController{}
				print.GET("/prescription/:id", printController.PrintPrescription)
				print.GET("/appointment/:id", printController.PrintAppointment)
			}

			// 医生管理（仅管理员）
			doctors := authorized.Group("/doctors")
			doctors.Use(middleware.RoleRequired("admin"))
			{
				doctorController := &controllers.DoctorController{}
				doctors.GET("", doctorController.List)
				doctors.GET("/:id", doctorController.Get)
				doctors.POST("", doctorController.Create)
				doctors.PUT("/:id", doctorController.Update)
				doctors.DELETE("/:id", doctorController.Delete)
				doctors.POST("/:id/reset-password", doctorController.ResetPassword)
			}
		}
	}

	// 启动服务器
	port := ":8080"
	log.Printf("服务器启动在端口 %s", port)
	log.Println("访问地址: http://localhost:8080")
	log.Println("局域网访问: http://[本机IP]:8080")

	// 优雅关闭
	go func() {
		if err := r.Run(port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")
}
