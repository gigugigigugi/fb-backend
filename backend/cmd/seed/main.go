package main

import (
	"fmt"
	"log"

	"football-backend/common/config"
	"football-backend/common/database"
	"football-backend/internal/model"

	"time"
)

func main() {
	// 1. 加载配置 (会自动读取 .env 文件的 DB_DSN)
	config.Load()
	db := database.Init(config.App.DB.DSN)

	// 2. 在执行 AutoMigrate 之前，由于结构发生过破坏性变更（比如由 mobile 变成了必填的 email）
	// AutoMigrate 无法为存量的 null email 记录添加 NOT NULL 约束。
	// 所以我们在这里采用暴力重置：直接 DROP TABLE 然后再建。
	// [注意] 仅限于 cmd/seed (开发测试阶段) 这么玩！
	log.Println("Dropping old tables to avoid migration conflicts...")
	db.Migrator().DropTable(&model.Match{}, &model.TeamMember{}, &model.Team{}, &model.User{}, &model.Booking{}, &model.Venue{}, &model.Comment{})

	// 3. 自动迁移所有的模型
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

	// （第 5 步：由于上面的 DropTable 已经清空了数据，这段清理旧数据的代码就可以删除了）

	for i := range venues {
		if err := db.Create(&venues[i]).Error; err != nil {
			log.Printf("Failed to insert venue %s: %v", venues[i].Name, err)
		} else {
			fmt.Printf("Inserted venue: %s\n", venues[i].Name)
		}
	}

	// 5. 生成一个队长用户
	captain := model.User{
		Email:        "admin@football.com",
		PasswordHash: "mock_hash",
		Nickname:     "测试队长",
		Avatar:       "https://google.com/avatar.png",
	}
	db.Create(&captain)

	// 6. 生成一支队伍
	team := model.Team{
		Name:      "FC 东京测试队",
		Slogan:    "We are the champions!",
		CaptainID: captain.ID,
	}
	db.Create(&team)

	// 7. 批量铺设 20 场未来的比赛数据，打散到不同的场地和时间
	now := time.Now()
	for i := 1; i <= 20; i++ {
		// 让比赛时间在一星期内的不同天随机发散
		matchTime := now.Add(time.Duration(i*12) * time.Hour)

		venue := venues[i%len(venues)] // 轮流使用创建出的10个场地

		matchStatus := "RECRUITING"
		if i%4 == 0 {
			matchStatus = "FULL" // 注入一些假满员数据
		} else if i%7 == 0 {
			matchStatus = "CANCELED" // 注入一些已取消数据
		}

		match := model.Match{
			TeamID:     team.ID,
			VenueID:    venue.ID,
			StartTime:  matchTime,
			EndTime:    matchTime.Add(2 * time.Hour),
			Price:      1500.0,
			MaxPlayers: 14,
			Format:     []int{5, 7, 11}[i%3], // 循环分配赛制
			Note:       fmt.Sprintf("这是第%d场友谊赛，大家开心踢球！", i),
			Status:     matchStatus,
		}
		db.Create(&match)
	}

	fmt.Println("\n🎉 Seed completely successfully! Venues, Teams, Users, and Matches have been initialized.")
}
