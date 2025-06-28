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

type MedicineController struct{}

func (mc *MedicineController) Create(c *gin.Context) {
	var medicine models.Medicine
	if err := c.ShouldBindJSON(&medicine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	now := time.Now()
	result, err := database.DB.Exec(`
		INSERT INTO medicines (name, specification, unit, price, stock, min_stock, category, manufacturer, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		medicine.Name, medicine.Specification, medicine.Unit, medicine.Price, medicine.Stock,
		medicine.MinStock, medicine.Category, medicine.Manufacturer, now, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建药品失败"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{
		"message": "药品创建成功",
		"id":      id,
	})
}

func (mc *MedicineController) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的药品ID"})
		return
	}

	var medicine models.Medicine
	err = database.DB.QueryRow(`
		SELECT id, name, specification, unit, price, stock, min_stock, category, manufacturer, created_at, updated_at
		FROM medicines WHERE id = ?`, id).Scan(
		&medicine.ID, &medicine.Name, &medicine.Specification, &medicine.Unit, &medicine.Price,
		&medicine.Stock, &medicine.MinStock, &medicine.Category, &medicine.Manufacturer, &medicine.CreatedAt, &medicine.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "药品不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"medicine": medicine})
}

func (mc *MedicineController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的药品ID"})
		return
	}

	var medicine models.Medicine
	if err := c.ShouldBindJSON(&medicine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	_, err = database.DB.Exec(`
		UPDATE medicines SET name = ?, specification = ?, unit = ?, price = ?, stock = ?, 
		min_stock = ?, category = ?, manufacturer = ?, updated_at = ? WHERE id = ?`,
		medicine.Name, medicine.Specification, medicine.Unit, medicine.Price, medicine.Stock,
		medicine.MinStock, medicine.Category, medicine.Manufacturer, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新药品失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "药品信息更新成功"})
}

func (mc *MedicineController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的药品ID"})
		return
	}

	_, err = database.DB.Exec("DELETE FROM medicines WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除药品失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "药品删除成功"})
}

func (mc *MedicineController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	category := c.Query("category")

	offset := (page - 1) * limit

	var query string
	var args []interface{}

	whereClause := "WHERE 1=1"
	if search != "" {
		whereClause += " AND (name LIKE ? OR specification LIKE ? OR manufacturer LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if category != "" {
		whereClause += " AND category = ?"
		args = append(args, category)
	}

	query = `
		SELECT id, name, specification, unit, price, stock, min_stock, category, manufacturer, created_at, updated_at
		FROM medicines ` + whereClause + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询药品列表失败"})
		return
	}
	defer rows.Close()

	var medicines []models.Medicine
	for rows.Next() {
		var medicine models.Medicine
		err := rows.Scan(
			&medicine.ID, &medicine.Name, &medicine.Specification, &medicine.Unit, &medicine.Price,
			&medicine.Stock, &medicine.MinStock, &medicine.Category, &medicine.Manufacturer, &medicine.CreatedAt, &medicine.UpdatedAt)
		if err != nil {
			continue
		}
		medicines = append(medicines, medicine)
	}

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM medicines " + whereClause
	countArgs := args[:len(args)-2] // 去掉 LIMIT 和 OFFSET 参数
	var total int
	database.DB.QueryRow(countQuery, countArgs...).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"medicines": medicines,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

func (mc *MedicineController) Search(c *gin.Context) {
	var search models.MedicineSearch
	if err := c.ShouldBindJSON(&search); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	query := `
		SELECT id, name, specification, unit, price, stock, min_stock, category, manufacturer, created_at, updated_at
		FROM medicines WHERE 1=1`
	var args []interface{}

	if search.Name != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+search.Name+"%")
	}
	if search.Category != "" {
		query += " AND category = ?"
		args = append(args, search.Category)
	}
	if search.Manufacturer != "" {
		query += " AND manufacturer LIKE ?"
		args = append(args, "%"+search.Manufacturer+"%")
	}

	query += " ORDER BY name ASC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索药品失败"})
		return
	}
	defer rows.Close()

	var medicines []models.Medicine
	for rows.Next() {
		var medicine models.Medicine
		err := rows.Scan(
			&medicine.ID, &medicine.Name, &medicine.Specification, &medicine.Unit, &medicine.Price,
			&medicine.Stock, &medicine.MinStock, &medicine.Category, &medicine.Manufacturer, &medicine.CreatedAt, &medicine.UpdatedAt)
		if err != nil {
			continue
		}
		medicines = append(medicines, medicine)
	}

	c.JSON(http.StatusOK, gin.H{"medicines": medicines})
}

func (mc *MedicineController) UpdateStock(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的药品ID"})
		return
	}

	var req struct {
		Stock int `json:"stock" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	_, err = database.DB.Exec("UPDATE medicines SET stock = ?, updated_at = ? WHERE id = ?",
		req.Stock, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新库存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "库存更新成功"})
}

func (mc *MedicineController) GetCategories(c *gin.Context) {
	rows, err := database.DB.Query("SELECT DISTINCT category FROM medicines WHERE category IS NOT NULL AND category != '' ORDER BY category")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取药品分类失败"})
		return
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			continue
		}
		categories = append(categories, category)
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func (mc *MedicineController) GetLowStock(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT id, name, specification, unit, price, stock, min_stock, category, manufacturer, created_at, updated_at
		FROM medicines WHERE stock <= min_stock ORDER BY stock ASC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询低库存药品失败"})
		return
	}
	defer rows.Close()

	var medicines []models.Medicine
	for rows.Next() {
		var medicine models.Medicine
		err := rows.Scan(
			&medicine.ID, &medicine.Name, &medicine.Specification, &medicine.Unit, &medicine.Price,
			&medicine.Stock, &medicine.MinStock, &medicine.Category, &medicine.Manufacturer, &medicine.CreatedAt, &medicine.UpdatedAt)
		if err != nil {
			continue
		}
		medicines = append(medicines, medicine)
	}

	c.JSON(http.StatusOK, gin.H{"medicines": medicines})
}

// 药品自动完成搜索
func (mc *MedicineController) AutoComplete(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusOK, gin.H{"medicines": []interface{}{}})
		return
	}

	// 获取所有药品进行本地搜索
	rows, err := database.DB.Query(`
		SELECT id, name, specification, unit, price, stock
		FROM medicines 
		ORDER BY name ASC`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索药品失败"})
		return
	}
	defer rows.Close()

	var allMedicines []map[string]interface{}
	for rows.Next() {
		var id int
		var name, specification, unit string
		var price float64
		var stock int

		err := rows.Scan(&id, &name, &specification, &unit, &price, &stock)
		if err != nil {
			continue
		}

		allMedicines = append(allMedicines, map[string]interface{}{
			"id":            id,
			"name":          name,
			"specification": specification,
			"unit":          unit,
			"price":         price,
			"stock":         stock,
		})
	}

	// 本地搜索，支持拼音和英文
	var matchedMedicines []map[string]interface{}
	queryLower := strings.ToLower(query)

	for _, medicine := range allMedicines {
		name := medicine["name"].(string)
		specification := medicine["specification"].(string)

		// 直接匹配（中文、英文）
		if strings.Contains(strings.ToLower(name), queryLower) ||
			strings.Contains(strings.ToLower(specification), queryLower) {
			matchedMedicines = append(matchedMedicines, medicine)
			continue
		}

		// 拼音匹配
		namePinyin := getPinyin(name)
		specPinyin := getPinyin(specification)

		if strings.Contains(strings.ToLower(namePinyin), queryLower) ||
			strings.Contains(strings.ToLower(specPinyin), queryLower) {
			matchedMedicines = append(matchedMedicines, medicine)
			continue
		}

		// 首字母匹配
		nameInitials := getInitials(name)
		specInitials := getInitials(specification)

		if strings.Contains(strings.ToLower(nameInitials), queryLower) ||
			strings.Contains(strings.ToLower(specInitials), queryLower) {
			matchedMedicines = append(matchedMedicines, medicine)
		}
	}

	// 限制结果数量
	if len(matchedMedicines) > 10 {
		matchedMedicines = matchedMedicines[:10]
	}

	c.JSON(http.StatusOK, gin.H{"medicines": matchedMedicines})
}

// 获取拼音
func getPinyin(text string) string {
	args := pinyin.NewArgs()
	args.Style = pinyin.Normal
	args.Separator = ""

	pinyinSlice := pinyin.Pinyin(text, args)
	var result strings.Builder

	for _, p := range pinyinSlice {
		if len(p) > 0 {
			result.WriteString(p[0])
		}
	}

	return result.String()
}

// 获取首字母
func getInitials(text string) string {
	args := pinyin.NewArgs()
	args.Style = pinyin.FirstLetter

	pinyinSlice := pinyin.Pinyin(text, args)
	var result strings.Builder

	for _, p := range pinyinSlice {
		if len(p) > 0 {
			result.WriteString(p[0])
		}
	}

	return result.String()
}
