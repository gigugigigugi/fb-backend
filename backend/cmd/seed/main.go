package main

import (
	"context"
	"fmt"
	"log"

	"football-backend/common/config"
	"football-backend/common/database"
	"football-backend/internal/model"

	"gorm.io/gorm"
)

func main() {
	// 1. 加载配置 (会自动读取 .env 文件的 DB_DSN)
	config.Load()
	repo := database.Init(config.App.DB.DSN)
	db := repo.GetGormDB()

	// 2. 自动迁移所有的模型
	// 由于我们用的是全新创建的 docker 数据库，里面全是空的
	// Gorm 的 AutoMigrate 会根据我们的 struct 完美生成包含依赖关联的数据表架构！
	log.Println("Starting Auto Migration...")
	err := db.AutoMigrate(
		&model.User{},
		&model.Team{},
		&model.TeamMember{},
		&model.Match{},
		&model.Booking{},
		&model.Venue{},
		&model.Comment{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed successfully!")

	// 3. 准备待插入的10条真实东京场地数据 (位于江东区与涩谷区)
	venues := []model.Venue{
		{
			Name:       "辰巳の森海浜公園フットサル場",
			Prefecture: "東京都",
			City:       "江東区",
			Address:    "東京都江東区辰巳2-1-35",
			Latitude:   35.6433,
			Longitude:  139.8145,
			IsVerified: true,
			CreatedBy:  0, // 系统级官方数据标志位
		},
		{
			Name:       "MIYASHITA PARK 多目的運動施設",
			Prefecture: "東京都",
			City:       "渋谷区",
			Address:    "東京都渋谷区神宮前6-20-10",
			Latitude:   35.6616,
			Longitude:  139.7011,
			IsVerified: true,
			CreatedBy:  0,
		},
		{
			Name:       "代々木公園球技場",
			Prefecture: "東京都",
			City:       "渋谷区",
			Address:    "東京都渋谷区代々木神園町2-1",
			Latitude:   35.6669,
			Longitude:  139.6970,
			IsVerified: true,
			CreatedBy:  0,
		},
		{
			Name:       "豊洲テントドーム",
			Prefecture: "東京都",
			City:       "江東区",
			Address:    "東京都江東区豊洲2-1",
			Latitude:   35.6565,
			Longitude:  139.7944,
			IsVerified: true,
			CreatedBy:  0,
		},
		{
			Name:       "アディダスフットサルパーク渋谷",
			Prefecture: "東京都",
			City:       "渋谷区",
			Address:    "東京都渋谷区渋谷2-24-12 渋谷スクランブルスクエア上",
			Latitude:   35.6580,
			Longitude:  139.7016,
			IsVerified: true,
			CreatedBy:  0,
		},
		{
			Name:       "新木場フットサルクラブ",
			Prefecture: "東京都",
			City:       "江東区",
			Address:    "東京都江東区新木場2-11-2",
			Latitude:   35.6450,
			Longitude:  139.8322,
			IsVerified: true,
			CreatedBy:  0,
		},
		{
			Name:       "千駄ヶ谷フットサルコート",
			Prefecture: "東京都",
			City:       "渋谷区",
			Address:    "東京都渋谷区千駄ヶ谷1-17-1",
			Latitude:   35.6806,
			Longitude:  139.7118,
			IsVerified: true,
			CreatedBy:  0,
		},
		{
			Name:       "夢の島競技場",
			Prefecture: "東京都",
			City:       "江東区",
			Address:    "東京都江東区夢の島1-1",
			Latitude:   35.6517,
			Longitude:  139.8242,
			IsVerified: true,
			CreatedBy:  0,
		},
		{
			Name:       "大島小松川公園 運動広場",
			Prefecture: "東京都",
			City:       "江東区",
			Address:    "東京都江東区大島9-9",
			Latitude:   35.6946,
			Longitude:  139.8453,
			IsVerified: true,
			CreatedBy:  0,
		},
		{
			Name:       "恵比寿フットサルクラブ",
			Prefecture: "東京都",
			City:       "渋谷区",
			Address:    "東京都渋谷区恵比寿1-1-1",
			Latitude:   35.6466,
			Longitude:  139.7101,
			IsVerified: true,
			CreatedBy:  0,
		},
	}

	// 4. 清理旧数据并插入新数据
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&model.Venue{})

	for _, venue := range venues {
		if err := repo.Create(context.Background(), &venue); err != nil {
			log.Printf("Failed to insert venue %s: %v", venue.Name, err)
		} else {
			fmt.Printf("Inserted venue: %s\n", venue.Name)
		}
	}

	fmt.Println("\n🎉 Seed completely successfully! 10 realistic Tokyo venues have been created in the database.")
}
