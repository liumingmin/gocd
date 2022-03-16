package gocd

type CdService struct {
	name       string
	pkgUrl     string            // 程序包地址
	targetPath string            // 服务部署目标目录
	runCmd     string            // 启动文件或命令
	envVar     map[string]string // 动态参数-通过环境变量传递
	cdScript   *CdScript
}

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
