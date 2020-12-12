package service

import (
	"bcjh-bot/bot"
	"bcjh-bot/model/database"
	"bcjh-bot/model/onebot"
	"bcjh-bot/util"
	"bcjh-bot/util/logger"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func QuestQuery(c *onebot.Context, args []string) {
	str := strings.Join(args, util.ArgsSplitCharacter)
	pattern := regexp.MustCompile(`^(任务)?\s*(主线|支线)?\s*(任务)?\s*-?\s*([0-9]+(\.[0-9]+)?)([-\s]+([0-9]+))?`)
	allIndexes := pattern.FindAllSubmatchIndex([]byte(str), -1)
	// logger.Debugf("%v", allIndexes)

	// 不满足匹配条件
	if len(allIndexes) == 0 {
		_ = bot.SendMessage(c, "输入格式有误")
		return
	}
	pos := allIndexes[0]

	// 指定参数下标
	const (
		allPos = iota * 2
		mission1Pos
		typePos
		mission2Pos
		idPos
		isSubQuestPos
		hasLenPos
		lenPos
	)

	var idStr string
	var questType string
	length := 1

	// 初步确定查询的任务类型
	if pos[typePos] != -1 && pos[typePos+1] != -1 {
		questType = str[pos[typePos]:pos[typePos+1]]
	}

	// 初步确定任务 ID
	if pos[idPos] != -1 && pos[idPos+1] != -1 {
		idStr = str[pos[idPos]:pos[idPos+1]]
	} else {
		logger.Errorf("任务 ID 通过了正则但查询不到")
		return
	}

	// 确定主线还是支线
	if pos[isSubQuestPos] != -1 && pos[isSubQuestPos+1] != -1 &&
		strings.HasSuffix(str[pos[isSubQuestPos]:pos[isSubQuestPos+1]], ".") {
		if questType == "主线" {
			_ = bot.SendMessage(c, "您要找的是「支线任务 "+idStr+"」吗")
			return
		}
		questType = "支线"
	} else {
		if questType == "支线" {
			_ = bot.SendMessage(c, "您要找的是「主线任务 "+idStr+"」吗")
			return
		}
		questType = "主线"
	}

	// 查询条目数
	if pos[lenPos] != -1 && pos[lenPos+1] != -1 {
		length, _ = strconv.Atoi(str[pos[lenPos]:pos[lenPos+1]])
	}

	// logger.Debugf("查询结果：[%v %v] 查询%v条", questType, idStr, length)

	// 准备查询得到结果集
	quests := make([]database.Quest, 0)
	// 开始查询
	var err error
	if questType == "主线" {
		id, _ := strconv.Atoi(idStr)
		if id > 700 {
			_ = bot.SendMessage(c, "主线任务目前只有 700 个哦")
			return
		}
		if length == 1 {
			quests, err = findMainQuest(id)
		} else if length > util.MaxQueryListLength {
			length = util.MaxQueryListLength
			quests, err = findMainQuests(id, length)
		} else {
			quests, err = findMainQuests(id, length)
		}
	} else if questType == "支线" {
		quests, err = findSubQuest(idStr)
	}
	// 处理查询失败的错误
	if err != nil {
		logger.Errorf("查找出错：%v", err)
		_ = bot.SendMessage(c, util.SystemErrorNote)
		return
	}
	// 构造返回语句
	_ = bot.SendMessage(c, makeQuestsString(quests))
}

// 主线查询（单条）
func findMainQuest(id int) ([]database.Quest, error) {
	Session := database.DB.Where("type = ? and quest_id = ?", "主线任务", id).
		Limit(util.MaxQueryListLength)
	quests := make([]database.Quest, 0)
	if err := Session.Find(&quests); err != nil {
		return quests, err
	}
	return quests, nil
}

// 主线查询（多条）
func findMainQuests(id int, length int) ([]database.Quest, error) {
	Session := database.DB.Where("type = ? and quest_id >= ? and quest_id <= ?", "主线任务", id, id+length-1).
		Limit(util.MaxQueryListLength)
	quests := make([]database.Quest, 0)
	if err := Session.Find(&quests); err != nil {
		return quests, err
	}
	return quests, nil
}

// 支线查询（单条）
func findSubQuest(subId string) ([]database.Quest, error) {
	Session := database.DB.Where("type = ? and quest_id_disp = ?", "支线任务", subId).
		Limit(util.MaxQueryListLength)
	quests := make([]database.Quest, 0)
	if err := Session.Find(&quests); err != nil {
		return quests, err
	}
	return quests, nil
}

// 构造返回信息及格式
func makeQuestsString(quests []database.Quest) string {
	if len(quests) == 0 {
		return "哎呀，好像找不到呢!"
	}
	sb := strings.Builder{}

	for count, quest := range quests {
		sb.WriteString(fmt.Sprintf("[%s] ", quest.Type))
		if quest.Type == "主线任务" {
			sb.WriteString(fmt.Sprintf("%v", quest.QuestId))
		} else {
			sb.WriteString(fmt.Sprintf("%v", quest.QuestIdDisp))
		}
		sb.WriteString(fmt.Sprintf("\n要求：%s", quest.Goal))
		sb.WriteString("\n奖励：")
		if len(quest.Rewards) == 0 {
			sb.WriteString("无")
		} else {
			for i, reward := range quest.Rewards {
				if reward.Quantity == "" {
					sb.WriteString(fmt.Sprintf("%s", reward.Name))
				} else {
					sb.WriteString(fmt.Sprintf("%s*%v", reward.Name, reward.Quantity))
				}
				if i != len(quest.Rewards)-1 {
					sb.WriteString(", ")
				}
			}
		}
		if count != len(quests)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
