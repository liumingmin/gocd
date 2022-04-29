package gocd

type CdService struct {
	name       string
	pkgUrl     string            // 程序包名，defaultScript仅支持tgz格式程序包
	targetPath string            // 服务部署目标目录
	runCmd     string            // 启动文件或命令
	envVar     map[string]string // 动态参数-通过环境变量传递
	cdScript   *CdScript

	deployCounter uint32
}

//程序运行配置中，抽提db信息放到环境变量中运行时传递
//不同环境的配置文件直接写入程序包,动态内容使用环境变量设置
func NewDefaultCdService(name, pkgUrl, targetPath, runCmd string, envVar map[string]string) *CdService {
	if envVar == nil {
		envVar = make(map[string]string)
	}
	return &CdService{
		name:       name,
		pkgUrl:     pkgUrl,
		targetPath: targetPath,
		runCmd:     runCmd,
		envVar:     envVar,
		cdScript:   NewDefaultCdScript(),
	}
}

func (t *CdService) Name() string {
	return t.name
}

func (t *CdService) PkgUrl() string {
	return t.pkgUrl
}

func (t *CdService) TargetPath() string {
	return t.targetPath
}

func (t *CdService) RunCmd() string {
	return t.runCmd
}

func (t *CdService) EnvVar() map[string]string {
	return t.envVar
}

func (t *CdService) BindScript(cdScript *CdScript) {
	t.cdScript = cdScript
}

func (t *CdService) GetCdTaskScriptConfig(hostIp string) (string, error) {
	return t.cdScript.GetCdTaskScriptConfig(hostIp)
}
