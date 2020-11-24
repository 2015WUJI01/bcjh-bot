package service

import (
	"bcjh-bot/bot"
	"bcjh-bot/model/database"
	"bcjh-bot/model/onebot"
	"bcjh-bot/util"
	"bcjh-bot/util/logger"
	"fmt"
)

func ChefQuery(c *onebot.Context, args []string) {
	logger.Info("厨师查询，参数:", args)

	if len(args) == 0 {
		err := bot.SendMessage(c,
			fmt.Sprintf("指令示例:\n"+
				"%s厨师 羽十六", util.PrefixCharacter))
		if err != nil {
			logger.Error("发送信息失败!", err)
		}
		return
	}
	if args[0] == "%" {
		err := bot.SendMessage(c, "参数有误!")
		if err != nil {
			logger.Error("发送信息失败!", err)
		}
		return
	}

	chefs := make([]database.Chef, 0)
	err := database.DB.Where("gallery_id = ?", args[0]).Asc("gallery_id").Find(&chefs)
	if err != nil {
		logger.Error("查询数据库出错!", err)
		_ = bot.SendMessage(c, "查询数据失败!")
		return
	}
	if len(chefs) == 0 {
		err = database.DB.Where("name like ?", "%"+args[0]+"%").Asc("gallery_id").Find(&chefs)
		if err != nil {
			logger.Error("查询数据库出错!", err)
			_ = bot.SendMessage(c, "查询数据失败!")
			return
		}
	}

	var msg string
	if len(chefs) == 0 {
		msg = "哎呀，好像找不到呢!"
	} else if len(chefs) == 1 {
		chef := chefs[0]
		var gender string
		if chef.Gender == 1 {
			gender = "♂️"
		} else if chef.Gender == 2 {
			gender = "♀️"
		}
		rarity := ""
		for i := 0; i < chef.Rarity; i++ {
			rarity += "🔥"
		}
		skill := new(database.Skill)
		_, err = database.DB.Where("skill_id = ?", chef.SkillId).Get(skill)
		if err != nil {
			logger.Error("查询数据库出错!", err)
			_ = bot.SendMessage(c, "查询数据失败!")
			return
		}
		ultimate := new(database.Skill)
		_, err = database.DB.Where("skill_id = ?", chef.UltimateSkill).Get(ultimate)
		if err != nil {
			logger.Error("查询数据库出错!", err)
			_ = bot.SendMessage(c, "查询数据失败!")
			return
		}
		msg += fmt.Sprintf("%s %s %s\n", chef.GalleryId, chef.Name, gender)
		msg += fmt.Sprintf("%s\n", rarity)
		msg += fmt.Sprintf("来源: %s\n", chef.Origin)
		msg += fmt.Sprintf("炒:%d 烤:%d 煮:%d\n", chef.Stirfry, chef.Bake, chef.Boil)
		msg += fmt.Sprintf("蒸:%d 炸:%d 切:%d\n", chef.Steam, chef.Fry, chef.Cut)
		msg += fmt.Sprintf("🍖:%d 🍞:%d 🥕:%d 🐟:%d\n", chef.Meat, chef.Flour, chef.Vegetable, chef.Fish)
		msg += fmt.Sprintf("技能:%s\n", skill.Description)
		msg += fmt.Sprintf("修炼效果:%s", ultimate.Description)
	} else {
		msg = "查询到以下厨师:\n"
		for p, chef := range chefs {
			msg += fmt.Sprintf("%s %s", chef.GalleryId, chef.Name)
			if p != len(chefs)-1 {
				msg += "\n"
				if p == util.MaxSearchList-1 {
					msg += "......"
					break
				}
			}
		}
	}

	err = bot.SendMessage(c, msg)
	if err != nil {
		logger.Error("发送信息失败!", err)
	}
}
