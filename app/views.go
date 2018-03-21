package app

import (
	"drafter/service"
	. "drafter/setting"
	"html/template"
	"net/http"

	"gopkg.in/mgo.v2/bson"
)

type AdminViewModel struct {
	Config    AdminViewModelConfig
	StaticDir string
}

type AdminViewModelConfig struct {
	Preference    *service.BlogPreferenceValue `json:"preference"`
	Authenticated bool                         `json:"authenticated"`
}

func adminViews(w http.ResponseWriter, req *http.Request) {
	var (
		ctx   = GetContext(w, req)
		err   = VerifyController(ctx)
		model = new(AdminViewModel)
	)

	if err == nil {
		// 已登录
		// var p service.BlogPreferenceValue = service.BlogPreference
		// p, err = configS.GetBlogPreferences()
		model.Config.Preference = &service.BlogPreference
		model.Config.Authenticated = true
	}
	model.StaticDir = Settings.StaticDir

	t, err := template.ParseFiles("./templates/admin.html")
	if err == nil {
		err = t.Execute(w, model)
	}
	if err != nil {
		// 编译模板失败了、渲染失败了
		ctx.SendMessage("system error: template parsing/rendering.")
	}
}

type MainViewModel struct {
	Config    MainViewModelConfig
	Basic     MainViewModelBasic
	StaticDir string
}

type MainViewModelBasic struct {
	SiteName string `json:"sitename"`
	Domain   string `json:"domain"`
	Intro    string `json:"intro"`
	Author   string `json:"author"`
}

type MainViewModelConfig struct {
	Lists []MainViewListModel   `json:"lists",omitempty`
	Blogs []service.BlogPresent `json:"blogs",omitempty`
	// Loadings
	Tags []service.Tag `json:"tags",omitempty`
	// 第一个版本我们先不搞tagCloud吧
	// TagsSyncAt uint
}

type MainViewListModel struct {
	Query string            `json:"query"`
	Blogs [][]bson.ObjectId `json:"blogs"`
	Count int               `json:"count"`
}

func initMainViewModel() *MainViewModel {
	p := service.BlogPreference
	return &MainViewModel{
		Config: MainViewModelConfig{
			Lists: []MainViewListModel{},
			Blogs: []service.BlogPresent{},
			Tags:  []service.Tag{},
		},
		Basic: MainViewModelBasic{
			SiteName: p.SiteName,
			Domain:   p.Domain,
			Intro:    p.Intro,
			Author:   p.Author,
		},
		StaticDir: Settings.StaticDir,
	}
}

func mainView(w http.ResponseWriter, req *http.Request) {
	var (
		ctx         = GetContext(w, req)
		tag, hasTag = ctx.GetVar()["tag"]
		count       int
	)

	// var p int
	// if len(query.GetString("p")) > 0 {
	// 	p = query.GetInt("p")
	// 	if p <= 0 {
	// 		if hasTag {
	// 			ctx.Redirect("/tag/" + tag)
	// 		} else {
	// 			ctx.Redirect("/")
	// 		}
	// 		return
	// 	}
	// } else {
	// 	p = 1
	// }
	// p:=1
	// todo

	result, err := readerS.ListBlogPresents(1, tag, true)
	if err != nil {
		ctx.Redirect("/error")
		return
	}

	model := initMainViewModel()
	model.Config.Blogs = result["blogs"].([]service.BlogPresent)
	if hasTag {
		t := result["tag"].(service.Tag)
		model.Config.Tags = append([]service.Tag{}, t)
		count = t.Count
	} else {
		count = blogS.ValidCount
	}

	blogIds := []bson.ObjectId{}
	// 整理blog的id数组
	for _, b := range model.Config.Blogs {
		blogIds = append(blogIds, b.Id)
	}
	model.Config.Lists = append([]MainViewListModel{}, MainViewListModel{
		Query: tag,
		Blogs: [][]bson.ObjectId{blogIds},
		Count: count,
	})

	t, err := template.ParseFiles("./templates/index.html")

	if err == nil {
		err = t.Execute(w, model)
	}
	if err != nil {
		// 编译模板失败了、渲染失败了
		ctx.SendMessage("system error: template parsing/rendering.")
	}
}

func blogView(w http.ResponseWriter, req *http.Request) {
	var (
		ctx = GetContext(w, req)
		id  = ctx.GetVar()["id"]
	)

	if !bson.IsObjectIdHex(id) {
		ctx.Redirect("/error")
	}

	b, err := readerS.GetBlogPresent(bson.ObjectIdHex(id))
	if err != nil {
		ctx.Redirect("/error")
	}

	model := initMainViewModel()
	model.Config.Blogs = append([]service.BlogPresent{}, b)

	t, err := template.ParseFiles("./templates/index.html")

	if err == nil {
		err = t.Execute(w, model)
	}
	if err != nil {
		// 编译模板失败了、渲染失败了
		ctx.SendMessage("system error: template parsing/rendering.")
	}
}

func errorView(w http.ResponseWriter, req *http.Request) {
	var (
		ctx = GetContext(w, req)

		t, err = template.ParseFiles("./templates/index.html")
	)

	if err == nil {
		err = t.Execute(w, initMainViewModel())
	}
	if err != nil {
		// 编译模板失败了、渲染失败了
		ctx.SendMessage("system error: template parsing/rendering.")
	}
}
