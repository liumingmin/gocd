package gocd

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type CdService interface {
	GetName() string              // service name 服务名
	GetParams() map[string]string // invoke script dynamic params,see cdScript 传入脚本的动态参数与cdScript定义必须一致
	GetCdScript() *CdScript       // deploy script  部署脚本
	IncDeployCounter() uint32     // deploy counter 部署计数器，用于并发部署和计数
}

type DefaultCdService struct {
	name          string            // service name 服务名
	params        map[string]string // invoke script dynamic params,see cdScript 传入脚本的动态参数与cdScript定义必须一致
	cdScript      *CdScript         // deploy script  部署脚本
	deployCounter uint32            // deploy counter 部署计数器，用于并发部署和计数
}

//程序运行配置中，抽提db信息放到环境变量中运行时传递
//不同环境的配置文件直接写入程序包,动态内容使用环境变量设置
//pkgUrl     string            // 程序包名，defaultScript仅支持tgz格式程序包
//targetPath string            // 服务部署目标目录
//runCmd     string            // 启动文件或命令
//envVar     map[string]string // 动态参数-通过环境变量传递

func NewDefaultCdService(name, pkgUrl, targetPath, runCmd string, envVar map[string]string) CdService {
	var envsStr strings.Builder
	if envVar != nil {
		for key, value := range envVar {
			envsStr.WriteString(fmt.Sprintf(" %v=%v", key, value))
		}
	}

	cdService := &DefaultCdService{
		name: name,
		params: map[string]string{
			"PKG_URL":     pkgUrl, //s3get download package
			"TARGET_PATH": targetPath,
			"RUN_CMD":     runCmd,
			"ENV_VAR":     envsStr.String(),
		},
		cdScript: NewDefaultCdScript(),
	}

	return cdService
}

func (t *DefaultCdService) GetName() string {
	return t.name
}

func (t *DefaultCdService) GetParams() map[string]string {
	return t.params
}

func (t *DefaultCdService) GetCdScript() *CdScript {
	return t.cdScript
}

func (t *DefaultCdService) IncDeployCounter() uint32 {
	return atomic.AddUint32(&t.deployCounter, 1)
}

//implements
func (t *DefaultCdService) UpdatePkgUrl(pkgUrl string) {
	t.params["PKG_URL"] = pkgUrl
}
