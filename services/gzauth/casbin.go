package gzauth

import (
	"context"
	"errors"

	"github.com/casbin/casbin/v2"
	"github.com/spf13/cast"
)

type CasbinInfo struct {
	Path   string `json:"path"`   // 路径
	Method string `json:"method"` // 方法
}

func UpsertCasbin(ctx context.Context, casbin *casbin.SyncedEnforcer, roleId int64, casbinInfos []CasbinInfo) error {
	id := cast.ToString(roleId)
	_, _ = casbin.RemoveFilteredPolicy(0, id)
	rules := [][]string{}
	for _, v := range casbinInfos {
		rules = append(rules, []string{id, v.Path, v.Method})
	}

	success, err := casbin.AddPolicies(rules)
	if !success {
		return errors.New("存在相同api,添加失败,请联系管理员")
	}

	return err
}
