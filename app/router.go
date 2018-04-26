package app

import (
	e "drafter/exception"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Router struct {
	r      *mux.Router
	prefix string
}

var contexts = make(map[*http.Request]*context)

type controller func(*context) error

func (f controller) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := GetContext(w, req)
	err := f(ctx)
	if err != nil {
		errorHandler(ctx, err)
	} else if ctx.Session != nil {
		ctx.Session.Save(req, w)
	}
	delete(contexts, req)
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

// 这里存疑
// 我自己都忘了自己当时是怎么处理这个router的了
// 果然自己现在写不出更好的了，哎
func (r *Router) WithMiddleware(f controller) *Router {
	sub := NewRouter(mux.NewRouter())
	r.r.NewRoute().Handler(controller(func(ctx *context) (err error) {
		err = f(ctx)
		if err != nil {
			return
		}
		sub.r.ServeHTTP(ctx.Res, ctx.Req)
		return
	}))
	return sub
}

// func (r *Router) HandlePrefix(prefix string, methods string, f controller) {
// 	mtds := strings.Split(methods, ",")

// 	R := r.r.PathPrefix(prefix).HandlerFunc(f.ServeHTTP)
// 	if methods != "" {
// 		R = R.Methods(mtds...)
// 	}
// }

func GetContext(w http.ResponseWriter, req *http.Request) *context {
	ctx, ok := contexts[req]
	if !ok {
		ctx = &context{Req: req, Res: w}
		contexts[req] = ctx
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
