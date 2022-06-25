package sredis

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/frame/gins"
	"github.com/gogf/gf/util/gconv"
	"github.com/gogf/gf/util/gutil"
)

const (
	// DEFAULT_NAME Default group name for instance usage.
	DEFAULT_NAME     = "default"
	gREDIS_NODE_NAME = "redis"
)

var (
	// Instances map containing configuration instances.
	instances = gmap.NewStrAnyMap(true)
)


func Client(name ...string) redis.UniversalClient {
	config := gins.Config()
	key := DEFAULT_NAME
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}
	var opts *redis.UniversalOptions
	return instances.GetOrSetFuncLock(key, func() interface{} {
		var m map[string]interface{}
		if _, v := gutil.MapPossibleItemByKey(gins.Config().GetMap("."), gREDIS_NODE_NAME); v != nil {
			m = gconv.Map(v)
		}
		if len(m) > 0 {
			if v, ok := m[key]; ok {
				err := gconv.Struct(v, &opts)
				if err != nil {
					panic(err)
				}
				return redis.NewUniversalClient(opts)
			} else {
				panic(fmt.Sprintf(`configuration for redis not found for group "%s"`, key))
			}
		} else {
			panic(fmt.Sprintf(`incomplete configuration for redis: "redis" node not found in config file "%s"`, config.GetFileName()))
		}
		return nil
	}).(redis.UniversalClient)
}