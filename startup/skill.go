package startup

import (
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

// DefaultSkillsDir 默认的 Skills 存储目录
const DefaultSkillsDir = "/data/skills"

func InitSkillService() error {
	vars.SkillService = service.NewSkillService(DefaultSkillsDir, vars.DB)
	return vars.SkillService.Initialize()
}
