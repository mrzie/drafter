package app

import (
	e "drafter/exception"
	"drafter/service"

	"strings"

	"gopkg.in/mgo.v2/bson"
)

type notebookRequest struct {
	Id   bson.ObjectId `json:"id",omitempty bson:"_id"`
	Name string        `json:"name",omitempty`
}

// /notebook POST
func NewNotebookController(ctx *context) (err error) {
	var req notebookRequest
	err = ctx.GetReqStruct(&req)
	if err != nil {
		return
	}
	n, err := notebookS.Add(req.Name)
	if err != nil {
		return
	}
	return ctx.SendJson(n)
}

// /notebok DELETE
func DeleteNotebookController(ctx *context) (err error) {
	var req notebookRequest
	err = ctx.GetReqStruct(&req)
	if err != nil {
		return
	}

	err = notebookS.Remove(req.Id)

	if err != nil {
		return
	}
	return ctx.SendMessage("Delete notebook success.")
}

// /notebook PATCH
func RenameNotebookController(ctx *context) (err error) {
	var req notebookRequest
	err = ctx.GetReqStruct(&req)
	if err != nil {
		return
	}

	err = notebookS.Rename(req.Id, req.Name)

	if err != nil {
		return
	}
	return ctx.SendMessage("Rename notebook success.")
}

//  /notebooks GET
func ListNotebookController(ctx *context) (err error) {
	nbs, err := notebookS.List()
	if err != nil {
		return
	}
	return ctx.SendJson(nbs)
}

// /notes?notebookid= GET  返回一个笔记本下的所有笔记
func ListNoteController(ctx *context) (err error) {
	notebookid := ctx.GetQuery().GetString("notebookid")
	if !bson.IsObjectIdHex(notebookid) {
		err = e.InvalidId(notebookid)
		return
	}

	notes, err := noteS.List(bson.ObjectIdHex(notebookid))
	if err != nil {
		return
	}
	return ctx.SendJson(notes)
}

// /wastenote GET
func ListWasteNoteController(ctx *context) (err error) {
	notes, err := noteS.ListWaste()
	if err != nil {
		return
	}
	return ctx.SendJson(notes)
}

// note?id=1,2,3,4
func GetNotesController(ctx *context) (err error) {
	ids := strings.Split(ctx.GetQuery().GetString("id"), ",")
	oids := []bson.ObjectId{}
	for _, v := range ids {
		if !bson.IsObjectIdHex(v) {
			err = e.InvalidId(v)
			return
		}
		oids = append(oids, bson.ObjectIdHex(v))
	}

	notes, err := noteS.Get(oids)
	if err != nil {
		return
	}
	return ctx.SendJson(notes)
}

// /note POST
func NewNoteController(ctx *context) (err error) {
	var n service.Note
	err = ctx.GetReqStruct(&n)

	if err != nil {
		return
	}
	note, err := noteS.Compose(n)
	if err != nil {
		return
	}
	return ctx.SendJson(note)
}

// /note/{id} PATCH
func EditNoteController(ctx *context) (err error) {
	id := ctx.GetVar()["id"]
	if !bson.IsObjectIdHex(id) {
		err = e.InvalidId(id)
		return
	}
	reqSrc, err := ctx.GetReqBody()
	if err != nil {
		return
	}
	// var edition map[string]interface{}

	// err = json.Unmarshal(reqSrc, &edition)
	// if err != nil {
	// 	return
	// }
	err = noteS.Edit(bson.ObjectIdHex(id), reqSrc)
	if err != nil {
		return
	}

	return ctx.SendMessage("Edit note success.")
}

// /note/{id} DELETE
func RemoveNoteController(ctx *context) (err error) {
	id := ctx.GetVar()["id"]
	if !bson.IsObjectIdHex(id) {
		err = e.InvalidId(id)
		return
	}

	err = noteS.Remove(bson.ObjectIdHex(id))
	if err != nil {
		return
	}
	return ctx.SendMessage("Remove note success.")
}

//-----------------

// /blog?id=a,b,c GET
func GetBlogController(ctx *context) (err error) {
	// id := ctx.GetVar()["id"]
	// if !bson.IsObjectIdHex(id) {
	// 	err = e.InvalidId(id)
	// 	return
	// }
	ids := strings.Split(ctx.GetQuery().GetString("id"), ",")
	oids := []bson.ObjectId{}
	for _, v := range ids {
		if !bson.IsObjectIdHex(v) {
			err = e.InvalidId(v)
			return
		}
		oids = append(oids, bson.ObjectIdHex(v))
	}

	b, err := blogS.Get(oids)
	if err != nil {
		return
	}
	return ctx.SendJson(b)
}

// /blog POST
func ComposeBlogController(ctx *context) (err error) {
	var b service.Blog
	err = ctx.GetReqStruct(&b)
	if err != nil {
		return
	}
	err = blogS.Compose(&b)

	if err != nil {
		return
	}

	return ctx.SendJson(b)
}

// type editBlogTemplate struct {
// 	Noteid bson.ObjectId `json:"noteid"`
// }

// /blog/{id} PUT
func EditBlogController(ctx *context) (err error) {
	id := ctx.GetVar()["id"]
	if !bson.IsObjectIdHex(id) {
		err = e.InvalidId(id)
		return
	}

	// edition := make(map[string]interface{})

	// var req editBlogTemplate
	// err = ctx.GetReqStruct(&req)
	// if err != nil {
	// 	return
	// }
	// n, err := noteS.Get([]bson.ObjectId{req.Noteid})
	// if err != nil {
	// 	return
	// }
	// if len(n) == 0 {
	// 	err = e.NoteDoNotExist(req.Noteid)
	// }

	// edition, err := json.Marshal(map[string]interface{}{})

	reqSrc, err := ctx.GetReqBody()
	if err != nil {
		return
	}
	// edition := struct2map(n)
	nid := bson.ObjectIdHex(id)

	err = blogS.Edit(nid, reqSrc)
	if err != nil {
		return
	}
	return ctx.SendMessage("Edit blog success")
}

// /blogs?skip=&limit= GET
func ListBlogController(ctx *context) (err error) {
	query := ctx.GetQuery()
	skip := query.GetInt("skip")
	limit := query.GetInt("limit")
	blogs, err := blogS.List(skip, limit)
	if err != nil {
		return
	}
	return ctx.SendJson(blogs)
}

// /blog/{id} DELETE
func RemoveBlogController(ctx *context) (err error) {
	id := ctx.GetVar()["id"]
	if !bson.IsObjectIdHex(id) {
		err = e.InvalidId(id)
		return
	}
	err = blogS.Remove(bson.ObjectIdHex(id))
	if err != nil {
		return
	}

	return ctx.SendMessage("Delete blog success")
}

// /user-preference GET
func UserPreferenceController(ctx *context) (err error) {
	// p, err := configS.GetBlogPreferences()
	// if err != nil {
	// 	return
	// }
	// p := service.BlogPreference

	return ctx.SendJson(service.BlogPreference)
}

// /user-preference PUT
func SetUserPreferenceController(ctx *context) (err error) {
	var p service.BlogPreferenceValue
	err = ctx.GetReqStruct(&p)
	if err != nil {
		return
	}

	err = configS.SetBlogPreferences(p)
	if err != nil {
		return
	}

	return ctx.SendMessage("Success.")
}

type describeTagRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// /describe-tag PUT
func DescribeTagController(ctx *context) (err error) {
	var req describeTagRequest
	err = ctx.GetReqStruct(&req)
	if err != nil {
		return
	}

	err = tagS.SetDescription(req.Name, req.Description)
	if err != nil {
		return
	}

	return ctx.SendMessage("Describe tag success.")
}

func ListTagsController(ctx *context) (err error) {
	tags, err := tagS.List()
	if err != nil {
		return
	}

	return ctx.SendJson(tags)
}

func DeleteTagController(ctx *context) (err error) {
	name := ctx.GetVar()["name"]
	err = tagS.RemoveTag(name)
	if err != nil {
		return
	}
	return ctx.SendMessage("Delete tag success.")
}

// reader apis

// /blogs?p=
func ListBlogHandler(ctx *context) (err error) {
	q := ctx.GetQuery()
	page := q.GetInt("p")
	tag := q.GetString("tag")
	showTag := q.GetString("showTag") != ""
	if len(q.GetString("p")) == 0 {
		// 未传p参数
		page = 1
	}
	result, err := readerS.ListBlogPresents(page, tag, showTag)
	// todo 你写了sendJson和sendStruct了吗？
	return ctx.SendJson(result)
}

// /blog/{id}
func GetBlogHandler(ctx *context) (err error) {
	id := ctx.GetVar()["id"]
	if !bson.IsObjectIdHex(id) {
		err = e.InvalidId(id)
		return
	}

	blog, err := readerS.GetBlogPresent(bson.ObjectIdHex(id))
	if err != nil {
		return
	}

	return ctx.SendJson(blog)
}

// /tags
func GetTagsHandler(ctx *context) (err error) {
	tags := readerS.TagCounter()
	return ctx.SendJson(tags)
}

// verify apis

type loginRequest struct {
	Password  string `json:"password"`
	Exclusive bool   `json:"exclusive"`
}

var authInfoCatch service.AuthValue

func LoginHandler(ctx *context) (err error) {
	var req loginRequest
	err = ctx.GetReqStruct(&req)
	if err != nil {
		return
	}

	authInfo, err := authS.Login(req.Password, req.Exclusive)
	if err != nil {
		return
	}

	session := ctx.GetSession()
	session.Values["auth"] = authInfo
	err = ctx.SaveSession()
	if err != nil {
		return
	}

	return ctx.SendMessage("Login success.")
}

func VerifyController(ctx *context) (err error) {
	v, ok := ctx.GetSession().Values["auth"]
	if !ok {
		err = e.SessionNotFound()
		return
	}
	authInfo, ok := v.(service.AuthValue)
	if !ok {
		err = e.SessionNotFound()
		return
	}

	err = authS.Verify(authInfo)
	if err != nil {
		return
	}

	return nil
}

type editPasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func EditPasswordController(ctx *context) (err error) {
	var req editPasswordRequest
	err = ctx.GetReqStruct(&req)
	if err != nil {
		return
	}

	authInfo, err := authS.EditPassword(req.NewPassword, req.OldPassword)

	if err != nil {
		return
	}

	ctx.GetSession().Values["auth"] = authInfo
	ctx.SaveSession()

	return ctx.SendMessage("Reset password success.")
}
