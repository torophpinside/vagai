package main

import (
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/anomalyco/vagai-api/internal/services/analytics"
)

func main() {
	dsn := "vagai:vagai@tcp(127.0.0.1:3306)/vagai?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil { panic(err) }
	svc := analytics.New(db)
	_, err = svc.SkillGap(context.Background(), 17, 1)
	fmt.Println("SKILLGAP err:", err)
	_, err = svc.MatchingDistribution(context.Background(), 17, analytics.PageOptions{Page:1, PageSize:50})
	fmt.Println("MATCHING err:", err)
}
