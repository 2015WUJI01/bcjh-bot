package service

import (
	"bcjh-bot/bot"
	"bcjh-bot/model/database"
	"bcjh-bot/model/onebot"
	"bcjh-bot/util"
	"bcjh-bot/util/logger"
	"fmt"
	"strings"
)

// 调料查询
func CondimentQuery(c *onebot.Context, args []string) {
	// res := make([]database.Condiment, 0)
	// session := database.DB.Select("*")
	//
	// for _, id := range []int{153, 608, 123} {
	// 	session.SQL("json_contains(skill->'$', ?) ", id)
	// }
	// session.Find(&res)
	// logger.Warnf("%v", res)
	// return

	prefix := util.PrefixCharacter
	split := util.ArgsSplitCharacter
	logger.Info("调料查询，参数:", args)

	// 无参数的情况
	if len(args) == 0 {
		msg := "指令示例:\n"
		msg += "1. 简单查询，接名称或ID:\n"
		msg += fmt.Sprintf("「%s调料 香菜」「%s调料 1」\n", prefix, prefix)
		msg += "2. 限制品质，和菜名写在一起:\n"
		msg += fmt.Sprintf("「%s调料 三火香菜」「%s调料 3星香菜」\n", prefix, prefix)
		msg += "3. 限制技能:\n"
		msg += fmt.Sprintf("「%s调料 三火%s炒技法+15」「%s调料 三火%s采集」「%s调料 三火%s售价」\n",
			prefix, split,
			prefix, split,
			prefix, split)
		msg += "4. 限制来源，或对应阁楼的技法:\n"
		msg += fmt.Sprintf("「%s调料 %%%s%%%s梵正」「%s调料 三火%s%%%s切」",
			prefix, split,
			split, prefix,
			split, split)
		if err := bot.SendMessage(c, msg); err != nil {
			logger.Error("发送信息失败!", err)
		}
		return
	}

	// 建立会话
	Session := database.DB.Select("*")

	// 1. 第一个参数
	// 过滤等级，在调料中查询是否包含等级相关的参数，查询后保留原有参数
	if substr, ok := StringContainsAny(args[0], []string{"1火", "一星", "1星", "一火"}); ok {
		args[0] = strings.ReplaceAll(args[0], substr, "")
		Session.Where("rarity = 1")
	} else if substr, ok := StringContainsAny(args[0], []string{"2火", "二星", "2星", "二火"}); ok {
		args[0] = strings.ReplaceAll(args[0], substr, "")
		Session.Where("rarity = 2")
	} else if substr, ok := StringContainsAny(args[0], []string{"3火", "三星", "3星", "三火"}); ok {
		args[0] = strings.ReplaceAll(args[0], substr, "")
		Session.Where("rarity = 3")
	}

	// 安照名称或 ID 进行查询，若名称或 ID 为空，则默认查找全部
	if args[0] != "" {
		Session.Where("name like ? or condiment_id = ?", "%"+args[0]+"%", args[0])
	}

	// 2. 当传入第二个参数时，默认认为是技法
	skills := make([]database.Skill, 0)
	if len(args) > 1 && args[1] != "" && args[1] != "%" {
		// TODO: 还没做的东西
		_ = bot.SendMessage(c, "限制技能还没做好，用%或空代替")
		return
		_ = database.DB.Select("skill_id").Where("description like ?", "%"+args[1]+"%").Find(&skills)
		logger.Debugf("%v", skills)
	}

	// 3. 当传入第三个参数时，默认认为是来源
	if len(args) > 2 && args[2] != "" && args[2] != "%" {
		if skill, ok := StringContainsAny(args[2], []string{"切", "蒸", "炸", "煮", "烤", "炒"}); ok {
			Session.Where("origin = ?", switchSkillAndOrigin(skill))
		} else if origin, ok := StringContainsAny(args[2], []string{"庖丁", "梵正", "膳祖", "彭铿", "易牙", "伊尹"}); ok {
			Session.Where("origin = ?", origin+"阁")
		}
	}

	// 查询得到结果
	condiments := make([]database.Condiment, 0)
	err := Session.Asc("condiment_id").Find(&condiments)

	if err != nil {
		logger.Error("查询数据库出错!", err)
		_ = bot.SendMessage(c, "查询数据失败!")
		return
	}

	var msg string
	switch {
	case len(condiments) == 0:
		msg = "哎呀，好像找不到呢!"
	case condimentsDistinct(condiments) == 1:
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprintf("%d.%s   %s | %s",
			condiments[0].CondimentId,
			condiments[0].Name,
			condiments[0].Origin,
			switchSkillAndOrigin(condiments[0].Origin),
		))
		for _, condiment := range condiments {
			sb.WriteString("\n")
			for i := 0; i < condiment.Rarity; i++ {
				sb.WriteString("🔥")
			}
			skills := make([]database.Skill, 0)
			_ = database.DB.Select("description").In("skill_id", condiment.Skill).Find(&skills)
			logger.Debugf("%v", skills)
			for _, skill := range skills {
				sb.WriteString(fmt.Sprintf("\n%v", skill.Description))
			}
		}
		msg = sb.String()
	default:
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprintf("查询到%d种调料，共%d个\n", condimentsDistinct(condiments), len(condiments)))
		for p, condiment := range condiments {
			sb.WriteString(fmt.Sprintf("%d.%s   %s | %s\n",
				condiment.CondimentId,
				condiment.Name,
				condiment.Origin,
				switchSkillAndOrigin(condiment.Origin),
			))
			if p != len(condiments)-1 && p == util.MaxSearchList {
				sb.WriteString("......")
				break
			}
		}
		msg = sb.String()
	}
	_ = bot.SendMessage(c, msg)
}

// 查询字符串中是否包含字符串切片中的任意一项
func StringContainsAny(needle string, haystack []string) (string, bool) {
	for _, substr := range haystack {
		if strings.Contains(needle, substr) {
			return substr, true
		}
	}
	return "", false
}

func condimentsDistinct(condiments []database.Condiment) int {
	result := make([]string, 0, len(condiments))
	temp := map[string]struct{}{}
	for _, condiment := range condiments {
		if _, ok := temp[condiment.Name]; !ok {
			temp[condiment.Name] = struct{}{}
			result = append(result, condiment.Name)
		}
	}
	// logger.Debugf("%v", result)
	return len(result)
}

// 把「技法」或「xx 阁」进行互转
func switchSkillAndOrigin(origin string) string {
	switch origin {
	case "切":
		return "庖丁阁"
	case "蒸":
		return "梵正阁"
	case "炸":
		return "膳祖阁"
	case "煮":
		return "彭铿阁"
	case "烤":
		return "易牙阁"
	case "炒":
		return "伊尹阁"
	case "庖丁", "庖丁阁":
		return "切"
	case "梵正", "梵正阁":
		return "蒸"
	case "膳祖", "膳祖阁":
		return "炸"
	case "彭铿", "彭铿阁":
		return "煮"
	case "易牙", "易牙阁":
		return "烤"
	case "伊尹", "伊尹阁":
		return "炒"
	}
	return ""
}
