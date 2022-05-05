package gocd

import (
	"context"
	"fmt"
	"strings"

	"github.com/liumingmin/gojenkins"
	"github.com/liumingmin/goutils/log"
)

type CdNodeBroker struct {
	jenkins        *gojenkins.Jenkins
	env            string
	defCdNodeParam *CdNodeParam

	nodesCache map[string]*gojenkins.Node
}

func NewCdNodeBroker(jenkins *gojenkins.Jenkins, env string, nodeParam *CdNodeParam) *CdNodeBroker {
	cdNodeBroker := &CdNodeBroker{
		jenkins:        jenkins,
		env:            env,
		defCdNodeParam: nodeParam,
		nodesCache:     make(map[string]*gojenkins.Node),
	}

	if cdNodeBroker.defCdNodeParam == nil {
		cdNodeBroker.defCdNodeParam = NewCdNodeParam()
	}

	cdNodeBroker.UpdateNodeCache(context.Background())
	return cdNodeBroker
}

func (t *CdNodeBroker) SetDefCdNodeParam(defCdNodeParam *CdNodeParam) {
	if defCdNodeParam != nil {
		t.defCdNodeParam = defCdNodeParam
	}
}

//最好使用内网IP
func (t *CdNodeBroker) CreateNode(ctx context.Context, ip, remark string, options ...CdNodeOption) error {
	cdNodeInfo := t.defCdNodeParam
	if len(options) > 0 {
		cdNodeInfo = &CdNodeParam{}
		*cdNodeInfo = *t.defCdNodeParam

		for _, option := range options {
			option(cdNodeInfo)
		}
	}

	desc := fmt.Sprintf("%v:(%v)%v", t.env, ip, remark)
	node, err := t.jenkins.CreateNode(ctx, ip, cdNodeInfo.numExecutors, desc, cdNodeInfo.remoteFs, ip,
		map[string]string{
			"method":        "SSHLauncher",
			"host":          ip,
			"port":          cdNodeInfo.sshPort,
			"credentialsId": cdNodeInfo.credentialsId,
			"jvmOptions":    cdNodeInfo.jvmOptions,
		})
	if err != nil {
		log.Error(ctx, "CreateNode failed, err: %v", err)
		return err
	}

	t.UpdateNodeCache(ctx)
	log.Info(ctx, "CreateNode: %v", node)
	return nil
}

func (t *CdNodeBroker) DeleteNode(ctx context.Context, ip string) (bool, error) {
	node, err := t.jenkins.GetNode(ctx, ip)
	if err != nil {
		return false, err
	}

	ok, err := node.Delete(ctx)
	if err != nil {
		log.Error(ctx, "DeleteNode failed, err: %v", err)
	}

	if err == nil && ok {
		t.UpdateNodeCache(ctx)
	}
	return ok, err
}

func (t *CdNodeBroker) UpdateNodeCache(ctx context.Context) error {
	cache, err := t.getAllNodesMap(ctx)
	if err != nil || cache == nil {
		log.Error(ctx, "UpdateNodeCache failed, err: %v", err)
		return err
	}
	t.nodesCache = cache
	return nil
}

func (t *CdNodeBroker) GetNodeByName(name string) *gojenkins.Node {
	node, ok := t.nodesCache[name]
	if ok {
		return node
	}
	return nil
}

func (t *CdNodeBroker) getAllNodes(ctx context.Context) ([]*gojenkins.Node, error) {
	nodes, err := t.jenkins.GetAllNodes(ctx)
	if err != nil {
		return nil, err
	}

	envNodes := make([]*gojenkins.Node, 0, len(nodes))
	for _, node := range nodes {
		if !strings.HasPrefix(node.Raw.Description, t.env) && node.Raw.DisplayName != "master" {
			continue
		}

		envNodes = append(envNodes, node)
	}

	return envNodes, nil
}

func (t *CdNodeBroker) getAllNodesMap(ctx context.Context) (map[string]*gojenkins.Node, error) {
	nodes, err := t.getAllNodes(ctx)
	if err != nil {
		return make(map[string]*gojenkins.Node), err
	}

	nodesMap := make(map[string]*gojenkins.Node)
	for _, node := range nodes {
		nodesMap[node.GetName()] = node
	}
	return nodesMap, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type CdNodeParam struct {
	credentialsId string
	jvmOptions    string
	numExecutors  int
	remoteFs      string
	sshPort       string
}

func NewCdNodeParam(options ...CdNodeOption) *CdNodeParam {
	cdNodeInfo := &CdNodeParam{
		numExecutors: 1,
		jvmOptions:   "-Xms16m -Xmx64m",
		remoteFs:     "/var/lib/jenkins",
		sshPort:      "22",
	}
	if len(options) > 0 {
		for _, option := range options {
			option(cdNodeInfo)
		}
	}
	return cdNodeInfo
}

type CdNodeOption func(*CdNodeParam)

func CdNodeCredIdOption(credId string) CdNodeOption {
	return func(nodeInfo *CdNodeParam) {
		nodeInfo.credentialsId = credId
	}
}

func CdNodeJvmOption(jvm string) CdNodeOption {
	return func(nodeInfo *CdNodeParam) {
		nodeInfo.jvmOptions = jvm
	}
}

func CdNodeNumExecutorsOption(numExecutors int) CdNodeOption {
	return func(nodeInfo *CdNodeParam) {
		nodeInfo.numExecutors = numExecutors
	}
}

func CdNodeRemoteFsOption(remoteFs string) CdNodeOption {
	return func(nodeInfo *CdNodeParam) {
		nodeInfo.remoteFs = remoteFs
	}
}

func CdNodeSshPortOption(sshPort string) CdNodeOption {
	return func(nodeInfo *CdNodeParam) {
		nodeInfo.sshPort = sshPort
	}
}
