package service

import (
	"tu-xun/config"
)

type BoxSvc struct{}

// AutoApprove 自动审批状态查询
func (b *BoxSvc) AutoApprove() (resp string, err error) {
	switch config.Config.AUTO_APPROVAL {
	case "all":
		resp = "自动审批已开启，在校园范围内提交的图片会通过，答题与答案距离少于50m的会通过，评论也会自动通过"
	case "attemptAndComment":
		resp = "自动审批已开启，答题与答案距离少于50m的会通过，评论也会自动通过"
	case "comment":
		resp = "自动审批已开启，评论也会自动通过"
	case "no":
		resp = "自动审批未开启"
	default:
		resp = "看看运维？？？？？"
	}
	return resp, nil
}
