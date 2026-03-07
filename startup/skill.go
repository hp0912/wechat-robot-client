package startup

import (
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

func InitSkillService() error {
	vars.SkillService = service.NewSkillService(vars.SkillsDir, vars.DB)
	return vars.SkillService.Initialize()
}
