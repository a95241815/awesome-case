package sapollo

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
	"github.com/philchia/agollo/v4"
	"strconv"
	"strings"
)

type Config struct {
	Appid      string
	Cluster    string
	Namespaces []string
	Address    string
}

// Start ����apollo��ȡ�����ļ�
func Start(ac Config) {
	//���û�����û���������Ĭ���Ǳ��ػ���������apollo
	env := gstr.ToLower(genv.Get("ENV"))
	if env == "" {
		_ = g.Cfg().SetPath("etc")
		g.Cfg().SetFileName("config-local.yaml")
		g.Log().Info("��������δ����,��ȡ�����ļ�: config-local.yaml")
		return
	}
	err := agollo.Start(&agollo.Conf{
		AppID:          ac.Appid,
		Cluster:        ac.Cluster,
		NameSpaceNames: ac.Namespaces,
		MetaAddr:       ac.Address,
	}, agollo.SkipLocalCache())
	if err != nil {
		panic(err)
	}
	agollo.OnUpdate(func(event *agollo.ChangeEvent) {
		setConfig(ac.Namespaces)
	})
	setConfig(ac.Namespaces)

}

//��������
func setConfig(namespaces []string) {
	var yamlContents string
	//��ȡ����yaml�����ռ���������ݣ�д���ļ�����
	for _, namespace := range namespaces {
		if strings.HasSuffix(namespace, ".yaml") {
			content := agollo.GetContent(agollo.WithNamespace(namespace))
			yamlContents += "\n\n" + content
		}
	}
	writeYamlFile(yamlContents)
	//��ȡproperties�����ռ�����ã�д�ڴ�
	for _, namespace := range namespaces {
		if strings.HasSuffix(namespace, ".properties") {
			//��ȡ����������д���ڴ�
			allKeys := agollo.GetAllKeys(agollo.WithNamespace(namespace))
			for _, key := range allKeys {
				if key == "yaml" {
					continue
				}
				setKeyValue(key, agollo.GetString(key))
			}
		}
	}
}

// д��������
func setKeyValue(key string, value string) {
	if value == "true" {
		err := g.Cfg().Set(key, true)
		if err != nil {
			g.Log().Error(err)
		}
		return
	}
	if value == "false" {
		err := g.Cfg().Set(key, false)
		if err != nil {
			g.Log().Error(err)
		}
		return
	}
	// ���͵����
	if i, err := strconv.Atoi(value); err == nil {
		err := g.Cfg().Set(key, i)
		if err != nil {
			g.Log().Error(err)
		}
		return
	}
	// �����������
	if f, err := strconv.ParseFloat(value, 10); err == nil {
		err := g.Cfg().Set(key, f)
		if err != nil {
			g.Log().Error(err)
		}
		return
	}
	err := g.Cfg().Set(key, value)
	if err != nil {
		g.Log().Error(err)
	}
	return
}

// ��apollo������д���ļ�
func writeYamlFile(contents string) {
	dir := "etc"
	err := gfile.Mkdir(dir)
	if err != nil {
		g.Log().Error(err)
		panic(err)
	}
	filename := "config-" + gstr.ToLower(genv.Get("ENV")) + ".yaml"
	path := dir + "/" + filename
	err = gfile.PutContents(path, contents)
	if err != nil {
		g.Log().Error(err)
		panic(err)
	}
	//����Ĭ�ϵ������ļ�Ŀ¼�������ļ�
	err = g.Cfg().SetPath(dir)
	if err != nil {
		g.Log().Error(err)
		panic(err)
	}
	g.Cfg().SetFileName(filename)
	g.Log().Info("��ȡ�����ļ�: %v", path)
	return
}