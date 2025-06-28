package controllers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StatsController 统计控制器
type StatsController struct {
	DB *sql.DB
}

// NewStatsController 创建统计控制器
func NewStatsController(db *sql.DB) *StatsController {
	return &StatsController{DB: db}
}

// GetStats 获取统计数据
func (sc *StatsController) GetStats(c *gin.Context) {
	stats := make(map[string]interface{})

	// 获取总患者数
	var totalPatients int
	err := sc.DB.QueryRow("SELECT COUNT(*) FROM patients").Scan(&totalPatients)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取患者统计失败"})
		return
	}
	stats["total_patients"] = totalPatients

	// 获取总药品数
	var totalMedicines int
	err = sc.DB.QueryRow("SELECT COUNT(*) FROM medicines").Scan(&totalMedicines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取药品统计失败"})
		return
	}
	stats["total_medicines"] = totalMedicines

	// 获取总处方数
	var totalPrescriptions int
	err = sc.DB.QueryRow("SELECT COUNT(*) FROM prescriptions").Scan(&totalPrescriptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取处方统计失败"})
		return
	}
	stats["total_prescriptions"] = totalPrescriptions

	// 获取本周处方统计
	prescriptionWeekly, err := sc.getPrescriptionWeeklyStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取处方周统计失败"})
		return
	}
	stats["prescription_weekly"] = prescriptionWeekly

	// 获取药品分类统计
	medicineCategories, err := sc.getMedicineCategoryStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取药品分类统计失败"})
		return
	}
	stats["medicine_categories"] = medicineCategories

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// getPrescriptionWeeklyStats 获取本周处方统计
func (sc *StatsController) getPrescriptionWeeklyStats() ([]int, error) {
	// 获取本周的开始日期（周一）
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // 周日改为7
	}
	weekStart := now.AddDate(0, 0, -weekday+1)
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

	// 查询本周每天的处方数量
	weeklyStats := make([]int, 7)
	for i := 0; i < 7; i++ {
		dayStart := weekStart.AddDate(0, 0, i)
		dayEnd := dayStart.AddDate(0, 0, 1)

		var count int
		err := sc.DB.QueryRow(`
			SELECT COUNT(*) FROM prescriptions 
			WHERE created_at >= ? AND created_at < ?
		`, dayStart, dayEnd).Scan(&count)
		if err != nil {
			return nil, err
		}
		weeklyStats[i] = count
	}

	return weeklyStats, nil
}

// getMedicineCategoryStats 获取药品分类统计
func (sc *StatsController) getMedicineCategoryStats() ([]map[string]interface{}, error) {
	rows, err := sc.DB.Query(`
		SELECT category, COUNT(*) as count 
		FROM medicines 
		WHERE category IS NOT NULL AND category != '' 
		GROUP BY category 
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []map[string]interface{}
	for rows.Next() {
		var category string
		var count int
		err := rows.Scan(&category, &count)
		if err != nil {
			return nil, err
		}
		categories = append(categories, map[string]interface{}{
			"name":  category,
			"count": count,
		})
	}

	// 如果没有分类数据，返回默认值
	if len(categories) == 0 {
		categories = []map[string]interface{}{
			{"name": "其他", "count": 0},
		}
	}

	return categories, nil
}
