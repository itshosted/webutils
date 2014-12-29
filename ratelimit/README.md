Ratelimit
Limit the amount of HTTP-requests per X seconds using
Redis.

```
var (
Redis  *redis.Pool
)

func RedisPool(protocol string, server string) *redis.Pool {
  return &redis.Pool{
    MaxIdle:     3,
    IdleTimeout: 240 * time.Second,
    Dial: func() (redis.Conn, error) {
      c, err := redis.Dial(protocol, server)
      if err != nil {
        return nil, err
      }
      return c, err
    },
    TestOnBorrow: func(c redis.Conn, t time.Time) error {
      _, err := c.Do("PING")
      return err
    },
  }
}

func loadRedis() error {
  Redis = RedisPool("tcp", ":6379")
  return nil
}

ratelimit.SetRedis(Redis)
http.Handle("/", middleware.Use(mux))
```
