package http

// 读接口
type Reader interface {
    BindJSON(i any) error
    ShouldBind(i any) error
    PostForm(key string) string
}

// 查询读
type QueryReader interface {
    Query(key string) string
    DefaultQuery(key string, def string) string
}

// 路径数据读
type PathParamReader interface {
    Param(key string) string
}

// =============

// json 输出
type JSONWriter interface {
    JSON(code int, data any)
}

// 写接口
type Writer interface {
    JSONWriter
}
