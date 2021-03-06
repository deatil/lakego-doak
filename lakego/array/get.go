package array

import (
    "strconv"
    "strings"

    "github.com/deatil/go-goch/goch"
)

// 构造函数
func NewArr() Arr {
    return Arr{
        keyDelim: ".",
    }
}

// 获取
func ArrGet(source map[string]any, key string, defVal ...any) any {
    return NewArr().Get(source, key, defVal...)
}

// 获取
func ArrGetWithGoch(source map[string]any, key string, defVal ...any) goch.Goch {
    data := NewArr().Get(source, key, defVal...)

    return goch.New(data)
}

type (
    // Goch 别名
    Goch = goch.Goch
)

/**
 * 数组
 *
 * @create 2022-5-3
 * @author deatil
 */
type Arr struct {
    // 分隔符
    keyDelim string
}

// 设置 keyDelim
func (this Arr) WithKeyDelim(data string) Arr {
    this.keyDelim = data

    return this
}

// 获取
func (this Arr) Get(source map[string]any, key string, defVal ...any) any {
    data := this.Find(source, key)
    if data != nil {
        return data
    }

    if len(defVal) > 0 {
        return defVal[0]
    }

    return nil
}

// 查找
func (this Arr) Find(source map[string]any, key string) any {
    lowerKey := strings.ToLower(key)

    var (
        val    any
        path   = strings.Split(lowerKey, this.keyDelim)
        nested = len(path) > 1
    )

    // 索引
    val = this.searchIndexableWithPathPrefixes(source, path)
    if val != nil {
        return val
    }

    if nested && this.isPathShadowedInDeepMap(path, source) != "" {
        return nil
    }

    // map
    val = this.searchMap(source, path)
    if val != nil {
        return val
    }

    return nil
}

// 数组
func (this Arr) searchMap(source map[string]any, path []string) any {
    if len(path) == 0 {
        return source
    }

    next, ok := source[path[0]]
    if !ok {
        return nil
    }

    if len(path) == 1 {
        return next
    }

    switch next.(type) {
        case map[any]any:
            return this.searchMap(goch.ToStringMap(next), path[1:])
        case map[string]any:
            return this.searchMap(next.(map[string]any), path[1:])
        default:
    }

    return nil
}

// 索引查询
func (this Arr) searchIndexableWithPathPrefixes(source any, path []string) any {
    if len(path) == 0 {
        return source
    }

    for i := len(path); i > 0; i-- {
        prefixKey := strings.ToLower(strings.Join(path[0:i], this.keyDelim))

        var val any
        switch sourceIndexable := source.(type) {
            case []any:
                val = this.searchSliceWithPathPrefixes(sourceIndexable, prefixKey, i, path)
            case map[string]any:
                val = this.searchMapWithPathPrefixes(sourceIndexable, prefixKey, i, path)
        }

        if val != nil {
            return val
        }
    }

    return nil
}

// 切片
func (this Arr) searchSliceWithPathPrefixes(
    sourceSlice []any,
    prefixKey string,
    pathIndex int,
    path []string,
) any {
    index, err := strconv.Atoi(prefixKey)
    if err != nil || len(sourceSlice) <= index {
        return nil
    }

    next := sourceSlice[index]

    if pathIndex == len(path) {
        return next
    }

    switch n := next.(type) {
        case map[any]any:
            return this.searchIndexableWithPathPrefixes(goch.ToStringMap(n), path[pathIndex:])
        case map[string]any, []any:
            return this.searchIndexableWithPathPrefixes(n, path[pathIndex:])
        default:
    }

    return nil
}

// map 数据
func (this Arr) searchMapWithPathPrefixes(
    sourceMap map[string]any,
    prefixKey string,
    pathIndex int,
    path []string,
) any {
    next, ok := sourceMap[prefixKey]
    if !ok {
        return nil
    }

    if pathIndex == len(path) {
        return next
    }

    switch n := next.(type) {
        case map[any]any:
            return this.searchIndexableWithPathPrefixes(goch.ToStringMap(n), path[pathIndex:])
        case map[string]any, []any:
            return this.searchIndexableWithPathPrefixes(n, path[pathIndex:])
        default:
    }

    return nil
}

// 是否合适
func (this Arr) isPathShadowedInDeepMap(path []string, m map[string]any) string {
    var parentVal any

    for i := 1; i < len(path); i++ {
        parentVal = this.searchMap(m, path[0:i])
        if parentVal == nil {
            return ""
        }

        switch parentVal.(type) {
            case map[any]any:
                continue
            case map[string]any:
                continue
            default:
                return strings.Join(path[0:i], this.keyDelim)
        }
    }

    return ""
}
