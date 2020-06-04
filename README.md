# xxcache
sync redis data to memory through rdb file,readonly command load data from local cache first,if key is not found, then load from redis.

# usage
```
import 	"github.com/liujiangwei/xxcache"

client , err := xxcache.New("localhost:6379")

client.Get("key")
```

# command Support

## string

- Set(key,value string) (string, error)
- SetNX(key, value string)(string, error)
- SetEX(key, value string, expires uint64)(string, error)
- PSetEX(key, value string, expires uint64)(string, error)
- GET(key string)(string, error)
- GetSet(key, value string)(string, error)
- StrLen(key string)(int, error)
- Append(key string)(int, error)
- SetRange(key string, pos int, replace string)(int, error)
- GetRange(key string, start, end int)(string, error)
- Incr(key string)(int, error)
- IncrBy(key string)(int, error)
- IncrByFloat(key string, increment float64)(float64, error)
- Decr(key string)(int, error)
- DecrBy(key string)(int, error)
- MSet(kv map[string]string)(string, error)
- MSetNX(kv map[string]string)(string, error)
- MGet(keys ...string)([]string, error)
