package app

import (
	e "drafter/exception"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

type Router struct {
	r      *mux.Router
	prefix string
}

var contexts = make(map[*http.Request]*context)
var contextsMux sync.Mutex

type controller func(*context) error

func (f controller) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := GetContext(w, req)
	err := f(ctx)
	if err != nil {
		errorHandler(ctx, err)
	} else if ctx.Session != nil {
		ctx.Session.Save(req, w)
	}
	contextsMux.Lock()
	delete(contexts, req)
	contextsMux.Unlock()
}

func (r *Router) Handle(route string, methods string, f controller) {
	mtds := strings.Split(methods, ",")

	r.r.HandleFunc(r.prefix+route, f.ServeHTTP).Methods(mtds...)
}

func (r *Router) HandleMidware(prefix string, methods string, f controller) *Router {
	mtds := strings.Split(methods, ",")
	sub := NewRouter(mux.NewRouter())
	prefixedRoute := r.r.PathPrefix(prefix)
	pathTemplate, _ := prefixedRoute.GetPathTemplate()
	sub.prefix = pathTemplate
	// if pathTemplate != "" {

	// }
	R := prefixedRoute.Handler(controller(func(ctx *context) (err error) {

		err = f(ctx)
		if err != nil {
			return
		}

		sub.r.ServeHTTP(ctx.Res, ctx.Req)
		return
	}))
	if methods != "" {
		R = R.Methods(mtds...)
	}

	return sub
}

func (r *Router) WithMiddleware(f controller) *RouterWithMiddleware {
	return GetRouterWithMiddleWare(r, f)
}

// func (r *Router) HandlePrefix(prefix string, methods string, f controller) {
// 	mtds := strings.Split(methods, ",")

// 	R := r.r.PathPrefix(prefix).HandlerFunc(f.ServeHTTP)
// 	if methods != "" {
// 		R = R.Methods(mtds...)
// 	}
// }

type RouterWithMiddleware struct {
	r         *Router
	decorator func(controller) controller
}

func (this *RouterWithMiddleware) Handle(route string, methods string, f controller) {
	this.r.Handle(route, methods, this.decorator(f))
}

func GetRouterWithMiddleWare(r *Router, middleware controller) *RouterWithMiddleware {
	return &RouterWithMiddleware{
		r: r,
		decorator: func(f controller) controller {
			return func(ctx *context) (err error) {
				err = middleware(ctx)
				if err != nil {
					return
				}
				return f(ctx)
			}
		},
	}
}

func GetContext(w http.ResponseWriter, req *http.Request) *context {
	ctx, ok := contexts[req]
	if !ok {
		ctx = &context{Req: req, Res: w}
		contextsMux.Lock()
		contexts[req] = ctx
		contextsMux.Unlock()
	}
	return ctx
}

// func (r *Router) Prefix(prefix string, methods string) *Router {
// 	R := r.r.PathPrefix(prefix)
// 	mtds := strings.Split(methods, ",")

// 	if len(mtds) != 0 {
// 		R = R.Methods(mtds...)
// 	}

// 	return NewRouter(R.Subrouter())
// }

func errorHandler(ctx *context, err error) {
	exc := e.Aver(err).(e.Exception)
	var statusCode int = 500
	switch {
	case exc.Code == 0:
		statusCode = 500
	case exc.Code == 900:
		statusCode = 500
	case exc.Code >= 100 && exc.Code < 200:
		// 	请求参数错误
		statusCode = 400
	case exc.Code >= 200 && exc.Code < 300:
		// 数据库错误
		statusCode = 500

	case exc.Code >= 300 && exc.Code < 400:
		// 代码错误
		statusCode = 500

	case exc.Code >= 400 && exc.Code < 500:
		statusCode = 401
	}
	ctx.Res.WriteHeader(statusCode)
	data, _ := json.Marshal(exc)
	ctx.Res.Write(data)
}

func NewRouter(router *mux.Router) *Router {
	r := new(Router)
	r.r = router
	return r
}
