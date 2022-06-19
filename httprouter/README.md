# httpRouter

> https://github.com/julienschmidt/httprouter/



## 标准库的 `http.ServeMux` 缺陷

- 当 URL 存在参数的时候，没办法很好的使用。
- `http.ServeMux` 的 URL 通常只能是固定的，当 URL 里面存在可变化的参数时，将很难发挥作用。
- 例如: `/PATH/topic/123/post/456`，URL 中有 topic id，post id 这样的非固定参数时，就很不方便
  - 标准库也可以做到，注册 `/PATH/topic`, 然后自己解析后面的 URL



## Router struct

- 实现了 `http.Handler` interface

```go
// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	/* route 登记树 */
	// key --> value = HTTP-Method ---> Route-Match-Tree
	// 1. HTTP-Method 找到相应 method 的 route-tree
	// 2. 在通过 URL Path 找到相应的 handle callback function
	trees map[string]*node
	paramsPool sync.Pool
	maxParams  uint16
    
	/* 功能标志位 */
	// 除了 SaveMatchedRoutePath，其他功能标志位都是默认开启的
	SaveMatchedRoutePath bool
	RedirectTrailingSlash bool
	RedirectFixedPath bool
	HandleMethodNotAllowed bool
	HandleOPTIONS bool
	GlobalOPTIONS http.Handler
	globalAllowed string
    
    /* 错误 URL 回复定制 */
	NotFound http.Handler
	MethodNotAllowed http.Handler
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})
}
```



### 路由的安装

**针对不同的 HTTP method 注册 callback**

```go
// HEAD is a shortcut for router.Handle(http.MethodHead, path, handle)
func (r *Router) HEAD(path string, handle Handle) {
	r.Handle(http.MethodHead, path, handle)
}

// OPTIONS is a shortcut for router.Handle(http.MethodOptions, path, handle)
func (r *Router) OPTIONS(path string, handle Handle) {
	r.Handle(http.MethodOptions, path, handle)
}

// POST is a shortcut for router.Handle(http.MethodPost, path, handle)
func (r *Router) POST(path string, handle Handle) {
	r.Handle(http.MethodPost, path, handle)
}

// PUT is a shortcut for router.Handle(http.MethodPut, path, handle)
func (r *Router) PUT(path string, handle Handle) {
	r.Handle(http.MethodPut, path, handle)
}

// PATCH is a shortcut for router.Handle(http.MethodPatch, path, handle)
func (r *Router) PATCH(path string, handle Handle) {
	r.Handle(http.MethodPatch, path, handle)
}

// DELETE is a shortcut for router.Handle(http.MethodDelete, path, handle)
func (r *Router) DELETE(path string, handle Handle) {
	r.Handle(http.MethodDelete, path, handle)
}
```

**实际实现**

```go
// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
// HTTP-Method + URL Path ---> handle callback function
// 1. HTTP-Method 找到相应 method 的 route-tree
// 2. 在通过 URL Path 找到相应的 handle callback function
func (r *Router) Handle(method, path string, handle Handle) {
	varsCount := uint16(0)
	.......
	// 相应 HTTP Method 的 route tree
	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root

		r.globalAllowed = r.allowed("*", "")
	}

	// 加入 route tree 里面
	root.addRoute(path, handle)

	.......
}
```





### 路由匹配、转发的实现

```go
// ServeHTTP makes the router implement the http.Handler interface.
// net/http 框架调用路径：
// 1. net/http/server.go:
// 2. func (sh serverHandler) ServeHTTP(rw ResponseWriter, req *Request)
// 3. 通过 handler.ServeHTTP(rw, req) 转发过来的
// TCP、http、TLS 全部由 net/http 框架完成
// HTTP Response 也是由 net/http 框架完成，httprouter 的工作是：更好更快的完成 URL 路由
/* 处理流程
 * 1. 从 http.Request 中获取基本信息（URL、HTTP Method 之类的）
 * 2. route-tree 匹配
 * 3. HTTP Method 检查
 * 4. 都不行，返回 404 Not Found
 */
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	........
	path := req.URL.Path

	if root := r.trees[req.Method]; root != nil {
		/* route tree match */
		// r.getParams 是不用传递的，直接把 r 闭包进 getParams 函数
		// 然后等 root.getValue() 又需要的时候才取出 Params 对象
		if handle, ps, tsr := root.getValue(path, r.getParams); handle != nil {
			/* 成功匹配，那就转发到相应的 handle function 去 */
			if ps != nil {
				handle(w, req, *ps)
				// Params 仅仅是这么临时用一用的而已
				// 因为 root.getValue() 只能通过内存逃逸的方式，将生成的 Params 放在堆上
				// 所以我们也只好通过 sync.Pool 的方式来减轻 GC 的压力
				r.putParams(ps)
			} else {
				handle(w, req, nil)
			}
			return
		}
		........
	}
	
	/* 严格匹配失败，返回错误信息 */
	/* 检查相应的功能标志位，并执行相应的函数 */
	if req.Method == http.MethodOptions && r.HandleOPTIONS {
		........
	} else if r.HandleMethodNotAllowed { // Handle 405
		if allow := r.allowed(path, req.Method); allow != "" {
			w.Header().Set("Allow", allow)
			if r.MethodNotAllowed != nil {
				r.MethodNotAllowed.ServeHTTP(w, req)
			} else {
				http.Error(w,
					http.StatusText(http.StatusMethodNotAllowed),
					http.StatusMethodNotAllowed,
				)
			}
			return
		}
	}

	// Handle 404
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}
```







## Param struct

- `httprouter.Params` 实际上是一个有序的 slice，里面装的是一对对的 key-value pair
- 通常我们会顺序进行获取解析，而不是通过 `Params.ByName()` 遍历查找
- 因为 `Params` struct 会经常使用，所以通过 `sync.Pool` 来做缓冲优化
- 号称极少 heap allocation 的 httprouter，这个也是位数不多要用到 heap 内存的地方

```go
// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) ByName(name string) string {
	for _, p := range ps {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}
```



