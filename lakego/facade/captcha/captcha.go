package captcha

import (
    "strings"
    "image/color"

    "github.com/mojocn/base64Captcha"

    "github.com/deatil/lakego-doak/lakego/array"
    "github.com/deatil/lakego-doak/lakego/register"
    "github.com/deatil/lakego-doak/lakego/facade/config"
    "github.com/deatil/lakego-doak/lakego/facade/logger"
    "github.com/deatil/lakego-doak/lakego/captcha"
    "github.com/deatil/lakego-doak/lakego/captcha/interfaces"
    redis_store "github.com/deatil/lakego-doak/lakego/captcha/store/redis"
)

// 初始化
func init() {
    // 注册默认
    Register()
}

/**
 * 验证码
 *
 * @create 2021-10-12
 * @author deatil
 */
func New(once ...bool) captcha.Captcha {
    // 默认驱动
    driver := GetDefaultDriver()

    // 默认存储
    store := GetDefaultStore()

    return Captcha(driver, store, once...)
}

// 验证码
func Captcha(driverName string, storeName string, once ...bool) captcha.Captcha {
    // 验证码配置
    conf := config.New("captcha")

    // 存储列表
    stores := conf.GetStringMap("Stores")

    // 转为小写
    storeName = strings.ToLower(storeName)

    // 获取配置
    storeConfig, ok := stores[storeName]
    if !ok {
        panic("验证码存储驱动[" + storeName + "]配置不存在")
    }

    // 配置
    storeConf := storeConfig.(map[string]any)

    storeType := storeConf["type"].(string)
    store := register.
        NewManagerWithPrefix("captcha-store").
        GetRegister(storeType, storeConf, once...)
    if store == nil {
        panic("验证码存储驱动[" + storeType + "]没有被注册")
    }

    // 驱动列表
    drivers := conf.GetStringMap("Drivers")

    // 转为小写
    driverName = strings.ToLower(driverName)

    // 获取配置
    driverConfig, ok := drivers[driverName]
    if !ok {
        panic("验证码驱动[" + driverName + "]配置不存在")
    }

    // 驱动配置
    driverConf := driverConfig.(map[string]any)

    driverType := driverConf["type"].(string)
    driver := register.
        NewManagerWithPrefix("captcha-driver").
        GetRegister(driverType, driverConf, once...)
    if driver == nil {
        panic("验证码驱动[" + driverType + "]没有被注册")
    }

    return captcha.New(driver.(interfaces.Driver), store.(interfaces.Store))
}

// 默认驱动
func GetDefaultDriver() string {
    return config.New("captcha").GetString("default-driver")
}

// 默认存储
func GetDefaultStore() string {
    return config.New("captcha").GetString("default-store")
}

// 注册
func Register() {
    // 注册存储
    register.
        NewManagerWithPrefix("captcha-store").
        RegisterMany(map[string]func(map[string]any) any {
            "redis": func(conf map[string]any) any {
                cfg := array.ArrayFrom(conf)

                store := redis_store.New(redis_store.Config{
                    DB:       cfg.Value("db").ToInt(),
                    Addr:     cfg.Value("addr").ToString(),
                    Password: cfg.Value("password").ToString(),

                    MinIdleConn:  cfg.Value("minidle-conn").ToInt(),
                    DialTimeout:  cfg.Value("dial-timeout").ToDuration(),
                    ReadTimeout:  cfg.Value("read-timeout").ToDuration(),
                    WriteTimeout: cfg.Value("write-timeout").ToDuration(),

                    PoolSize:     cfg.Value("pool-size").ToInt(),
                    PoolTimeout:  cfg.Value("pool-timeout").ToDuration(),

                    EnableTrace:  cfg.Value("enabletrace").ToBool(),

                    KeyPrefix:    cfg.Value("key-prefix").ToString(),
                    TTL:          cfg.Value("ttl").ToInt(),

                    Logger: logger.New(),
                })

                return store
            },
            // 验证码包该驱动有问题
            "syncmap": func(conf map[string]any) any {
                livetime := array.ArrayGet(conf, "livetime").ToDuration()

                syncmap := base64Captcha.NewStoreSyncMap(livetime)

                return syncmap
            },
            "memory": func(conf map[string]any) any {
                cfg := array.ArrayFrom(conf)

                collectNum := cfg.Value("collect-num").ToInt()
                expiration := cfg.Value("expiration").ToDuration()

                memory := base64Captcha.NewMemoryStore(collectNum, expiration)

                return memory
            },
        })

    // 注册驱动
    register.
        NewManagerWithPrefix("captcha-driver").
        RegisterMany(map[string]func(map[string]any) any {
            // 字符
            "string": func(conf map[string]any) any {
                /*
                //go:embed fonts/*.ttf
                //go:embed fonts/*.ttc
                var embeddedFontsFS embed.FS

                // 验证码字体驱动,
                var fontsStorage *base64Captcha.EmbeddedFontsStorage = base64Captcha.NewEmbeddedFontsStorage(embeddedFontsFS)
                */

                cfg := array.ArrayFrom(conf)

                cd := base64Captcha.NewDriverString(
                    cfg.Value("height").ToInt(),
                    cfg.Value("width").ToInt(),
                    cfg.Value("noise-count").ToInt(),
                    cfg.Value("showline-options").ToInt(),
                    cfg.Value("length").ToInt(),
                    cfg.Value("source").ToString(),
                    &color.RGBA{
                        R: cfg.Value("bgcolor.r").ToUint8(),
                        G: cfg.Value("bgcolor.g").ToUint8(),
                        B: cfg.Value("bgcolor.b").ToUint8(),
                        A: cfg.Value("bgcolor.a").ToUint8(),
                    },
                    // 自定义字体目录，参考 fontsStorage 相关注释
                    nil,
                    cfg.Value("fonts").ToStringSlice(),
                )

                driver := cd.ConvertFonts()

                return driver
            },
            // 中文
            "chinese": func(conf map[string]any) any {
                cfg := array.ArrayFrom(conf)

                cd := base64Captcha.NewDriverChinese(
                    cfg.Value("height").ToInt(),
                    cfg.Value("width").ToInt(),
                    cfg.Value("noise-count").ToInt(),
                    cfg.Value("showline-options").ToInt(),
                    cfg.Value("length").ToInt(),
                    cfg.Value("source").ToString(),
                    &color.RGBA{
                        R: cfg.Value("bgcolor.r").ToUint8(),
                        G: cfg.Value("bgcolor.g").ToUint8(),
                        B: cfg.Value("bgcolor.b").ToUint8(),
                        A: cfg.Value("bgcolor.a").ToUint8(),
                    },
                    // 自定义字体目录
                    nil,
                    cfg.Value("fonts").ToStringSlice(),
                )

                driver := cd.ConvertFonts()

                return driver
            },
            // 数学公式
            "math": func(conf map[string]any) any {
                cfg := array.ArrayFrom(conf)

                cd := base64Captcha.NewDriverMath(
                    cfg.Value("height").ToInt(),
                    cfg.Value("width").ToInt(),
                    cfg.Value("noise-count").ToInt(),
                    cfg.Value("showline-options").ToInt(),
                    &color.RGBA{
                        R: cfg.Value("bgcolor.r").ToUint8(),
                        G: cfg.Value("bgcolor.g").ToUint8(),
                        B: cfg.Value("bgcolor.b").ToUint8(),
                        A: cfg.Value("bgcolor.a").ToUint8(),
                    },
                    // 自定义字体目录
                    nil,
                    cfg.Value("fonts").ToStringSlice(),
                )

                driver := cd.ConvertFonts()

                return driver
            },
            // 音频
            "audio": func(conf map[string]any) any {
                cfg := array.ArrayFrom(conf)

                driver := base64Captcha.NewDriverAudio(
                    cfg.Value("length").ToInt(),
                    cfg.Value("language").ToString(),
                )

                return driver
            },
            // digit
            "digit": func(conf map[string]any) any {
                cfg := array.ArrayFrom(conf)

                driver := base64Captcha.NewDriverDigit(
                    cfg.Value("height").ToInt(),
                    cfg.Value("width").ToInt(),
                    cfg.Value("length").ToInt(),
                    cfg.Value("max-skew").ToFloat64(),
                    cfg.Value("dot-count").ToInt(),
                )

                return driver
            },
        })
}


