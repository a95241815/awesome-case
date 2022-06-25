package smongodb

import (
	"context"
	"fmt"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/frame/gins"
	"github.com/gogf/gf/util/gconv"
	"github.com/gogf/gf/util/gutil"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	instances = gmap.NewStrAnyMap(true)
)

type mongoConnectPool struct {
	sessions         []*mongo.Client
	nextSessionIndex int
	poolNum          int
	options          *options.ClientOptions
	database         string
}

const (
	// Default group name for instance usage.
	sDEFAULT_NAME      = "default"
	sMONGODB_NODE_NAME = "mongodb"
)

// Conn Instance ����һ�� go-redis �ļ�Ⱥ����
func Conn(name ...string) *mongoConnectPool {
	config := gins.Config()
	key := sDEFAULT_NAME
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}
	return instances.GetOrSetFuncLock(key, func() interface{} {
		var m map[string]interface{}
		if _, v := gutil.MapPossibleItemByKey(gins.Config().GetMap("."), sMONGODB_NODE_NAME); v != nil {
			m = gconv.Map(v)
		}
		if len(m) > 0 {
			if v, ok := m[key]; ok {
				opts := gconv.Map(v)
				var mongoConPool = new(mongoConnectPool)
				mongoConPool.options = options.Client().ApplyURI(gconv.String(opts["uri"]))
				mongoConPool.poolNum = gconv.Int(opts["poolNum"])
				mongoConPool.database = gconv.String(opts["database"])
				mongoConPool.mongoConnect()
				return mongoConPool
			} else {
				panic(fmt.Sprintf(`configuration for mongodb not found for group "%s"`, key))
			}
		} else {
			panic(fmt.Sprintf(`incomplete configuration for mongodb: "mongodb" node not found in config file "%s"`, config.GetFileName()))
		}
		return nil
	}).(*mongoConnectPool)
}

func (mcp *mongoConnectPool) mongoConnect() {
	//�� poolNum Ĭ��ֵ����Ϊ 10
	if mcp.poolNum == 0 {
		mcp.poolNum = 10
	}
	for i := 0; i < mcp.poolNum; i++ {
		session, err := mcp.getClient()
		if err != nil {
			continue
		}
		mcp.sessions = append(mcp.sessions, session)
	}
}

func (mcp *mongoConnectPool) getClient() (*mongo.Client, error) {
	client, err := mongo.Connect(context.TODO(), mcp.options)
	if err != nil {
		g.Log().Error(err)
		return nil, err
	}
	return client, nil
}

// GetSession �ӳ������� mongodb ������
func (mcp *mongoConnectPool) GetSession() *mongo.Client {
	var session *mongo.Client
	thisSessionIndex := mcp.nextSessionIndex
	session = mcp.sessions[thisSessionIndex]
	mcp.nextSessionIndex = (mcp.nextSessionIndex + 1) % len(mcp.sessions)
	//ȷ��һ��������û���ж�
	if err := session.Ping(context.TODO(), readpref.Primary()); err != nil {
		//��������ж��ˣ���Ͽ�����һ��
		_ = session.Disconnect(context.TODO())
		err = session.Connect(context.TODO())
		if err != nil {
			g.Log().Error(err)
		} else {
			mcp.sessions[thisSessionIndex] = session
		}
		return mcp.GetSession()
	}
	return session
}

// Ping ���������
func (mcp *mongoConnectPool) Ping() error {
	var session *mongo.Client
	session = mcp.sessions[mcp.nextSessionIndex]
	//ȷ��һ��������û���ж�
	if err := session.Ping(context.TODO(), readpref.Primary()); err != nil {
		return err
	}
	return nil
}

// GetColl ��ȡ�����ļ����ݿ�����ļ���
func (mcp *mongoConnectPool) GetColl(collection string) *mongo.Collection {
	return mcp.GetSession().Database(mcp.database).Collection(collection)
}

func (mcp *mongoConnectPool) MongoClear() {
	if mcp != nil {
		mcp.MongoDisconnect()
		mcp = nil
	}
}

func (mcp *mongoConnectPool) MongoDisconnect() {
	if mcp.sessions == nil || len(mcp.sessions) <= 0 {
		return
	}
	for _, session := range mcp.sessions {
		_ = session.Disconnect(context.TODO())
	}
	mcp.sessions = []*mongo.Client{}
}