package global

import (
	"bcjh-bot/dao"
	"bcjh-bot/model/database"
	"bcjh-bot/util/logger"
)

func init() {
	initPluginAliasComparison()
}

func IsSuperAdmin(qq int64) bool {
	has, err := dao.DB.Exist(&database.Admin{
		QQ: qq,
	})
	if err != nil {
		logger.Error("查询数据库出错:", err)
		return false
	}
	return has
}
