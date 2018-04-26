package app

import (
	e "drafter/exception"
	"drafter/service"

	"gopkg.in/mgo.v2/bson"
)

type GetCommentBrefResponse struct {
	OK       bool                  `json:"ok"`
	Comments []service.CommentBref `json:"comments"`
	Users    []service.UserInfo    `json:"users"`
	User     service.UserInfo      `json:"user"`
}

func GetCommentBrefController(ctx *context) (err error) {
	err = getCommentBrefHandler(ctx)
	if err != nil {
		return ctx.SendJson(GetCommentBrefResponse{})
	}
	return
}

// /comments?blog=123 GET
func getCommentBrefHandler(ctx *context) (err error) {
	topicId := ctx.GetQuery().GetString("blog")

	if topicId == "" || !bson.IsObjectIdHex(topicId) {
		return e.InvalidId(topicId)
	}
	tid := bson.ObjectIdHex(topicId)
	err = UserVerifyController(ctx)
	var comments []service.CommentBref
	if err != nil {
		// 未登录的用户
		comments, err = commentS.ListCommentBref(tid, nil)
	} else {
		comments, err = commentS.ListCommentBref(tid, nil)
	}
	if err != nil {
		return
	}

	uids := []bson.ObjectId{}
	for _, c := range comments {
		uids = append(uids, c.Id)
	}
	users, err := userS.GetUserInfos(uids)
	if err != nil {
		return
	}

	return ctx.SendJson(GetCommentBrefResponse{
		OK:       true,
		Comments: comments,
		Users:    users,
	})

}

func getUser(ctx *context) (u *service.User, err error) {
	user, _ := ctx.context["user"]
	u, ok := user.(*service.User)
	if !ok {
		err = e.SessionNotFound()
	}
	return
}

type ComposeCommentRequest struct {
	Content string        `json:"content"`
	Blog    bson.ObjectId `json:"blog"`
	Ref     bson.ObjectId `json"ref"`
}

func ComposeCommentController(ctx *context) (err error) {
	user, err := getUser(ctx)
	if err != nil {
		return
	}

	var req ComposeCommentRequest

	err = ctx.GetReqStruct(&req)
	if err != nil {
		return
	}

	// check if blog exist
	_, err = readerS.GetBlogPresent(req.Blog)
	if err != nil {
		return
	}

	err = commentS.Compose(*user, service.Comment{Content: req.Content, Blog: req.Blog, Ref: req.Ref})
	return
}

type ListCommentResponse struct {
	Comments []service.Comment `json:"comments"`
	Users    []service.User    `json:"users"`
}

// 接下来是管理员接口
// /comments?type=blocked&user=123&blog=456
func ListCommentController(ctx *context) (err error) {
	query := ctx.GetQuery()
	// var req =  ListCommentResponse{}
	if user, ok := query.value["user"]; ok {
		// 请求用户相关评论
		if !bson.IsObjectIdHex(user) {
			err = e.InvalidId(user)
			return
		}
		uid := bson.ObjectIdHex(user)
		comments, err := commentS.ListCommentByUser(uid)
		if err != nil {
			return err
		}
		// commentS.GenerateUserId(comments)
		users, err := userS.GetUsers([]bson.ObjectId{uid})
		if err != nil {
			return err
		}

		return ctx.SendJson(ListCommentResponse{
			Comments: comments,
			Users:    users,
		})
	} else if blog, ok := query.value["blog"]; ok {
		if !bson.IsObjectIdHex(blog) {
			err = e.InvalidId(blog)
			return
		}
		comments, err := commentS.ListCommentByBlog(bson.ObjectIdHex(blog))
		if err != nil {
			return err
		}
		users, err := userS.GetUsers(commentS.GenerateUserId(comments))
		if err != nil {
			return err
		}

		return ctx.SendJson(ListCommentResponse{
			Comments: comments,
			Users:    users,
		})
	} else if t, ok := query.value["type"]; ok {
		switch t {
		case "reviewing":
			return responseCommentList(ctx, commentS.ListReviewingComments)
		case "removed":
			return responseCommentList(ctx, commentS.ListRemovedComments)
		case "blocked":
			return responseCommentList(ctx, commentS.ListBlockedComments)
		default:
			return e.NotFound()
		}
	} else {
		return e.NotFound()
	}
}

func responseCommentList(ctx *context, fn func() ([]service.Comment, error)) (err error) {
	comments, err := fn()
	if err != nil {
		return
	}
	users, err := userS.GetUsers(commentS.GenerateUserId(comments))
	if err != nil {
		return
	}

	return ctx.SendJson(ListCommentResponse{
		Comments: comments,
		Users:    users,
	})
}

// comment?id=  DELETE
func DeleteCommentController(ctx *context) (err error) {
	id := ctx.GetQuery().GetString("id")
	if !bson.IsObjectIdHex(id) {
		return e.InvalidId(id)
	}

	err = commentS.DeleteComment(bson.ObjectIdHex(id))
	if err != nil {
		return
	}
	return // 这里要不要返回审查结果呢？
}

// revertComment?id=
func RevertCommentController(ctx *context) (err error) {
	id := ctx.GetQuery().GetString("id")
	if !bson.IsObjectIdHex(id) {
		return e.InvalidId(id)
	}
	err = commentS.RevertComment(bson.ObjectIdHex(id))
	if err != nil {
		return
	}
	return ctx.SendMessage("revert comment success.")
}

type CensorUserResponse struct {
	SensitiveComment *service.Comment `json:"sensitiveComment"` // 受到质疑的评论
}

// passComment?id=
func PassCommentController(ctx *context) (err error) {
	id := ctx.GetQuery().GetString("id")
	if !bson.IsObjectIdHex(id) {
		return e.InvalidId(id)
	}

	sensitiveOne, err := commentS.PassComment(bson.ObjectIdHex(id))
	if err != nil {
		return
	}
	return ctx.SendJson(CensorUserResponse{SensitiveComment: sensitiveOne})
}

// censorUser?id=
func CensorUserController(ctx *context) (err error) {
	id := ctx.GetQuery().GetString("id")
	if !bson.IsObjectIdHex(id) {
		return e.InvalidId(id)
	}
	sensitiveOne, err := userS.TrustUser(bson.ObjectIdHex(id))
	if err != nil {
		return
	}
	return ctx.SendJson(CensorUserResponse{SensitiveComment: sensitiveOne})
}

// blockComment?id=
func BlockCommentController(ctx *context) (err error) {
	id := ctx.GetQuery().GetString("id")
	if !bson.IsObjectIdHex(id) {
		return e.InvalidId(id)
	}
	err = commentS.BlockComment(bson.ObjectIdHex(id))
	if err != nil {
		return
	}
	return ctx.SendMessage("block comment success.")
}
