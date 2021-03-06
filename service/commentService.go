package service

import (
	"drafter/db"
	e "drafter/exception"
	"drafter/utils"
	"io/ioutil"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type commentService struct {
}

type Comment struct {
	Id      bson.ObjectId `json:"id" bson:"_id",omitempty`
	Ref     string        `json:"ref"`
	Blog    bson.ObjectId `json:"blog"` // blog id
	User    bson.ObjectId `json:"user"`
	State   CommentState  `json:"state"` // 0 - reviewing 1 - pass 2 - implicated 3 - block
	Alive   bool          `json:"alive"`
	Time    int64         `json:"time"`
	Content string        `json:"content"`
}

type CommentState uint8

const (
	COMMENT_REVIEWING  CommentState = 0
	COMMENT_PASS       CommentState = 1
	COMMENT_IMPLICATED CommentState = 2
	COMMENT_BLOCKED    CommentState = 3
)

type CommentBref struct {
	Id      bson.ObjectId `json:"id" bson:"_id",omitempty`
	Ref     string        `json:"ref"`
	User    bson.ObjectId `json:"user"`
	Time    int64         `json:"time"`
	Content string        `json:"content"`
	State   CommentState  `json:"state"`
}

var CommentService = new(commentService)

func (this *commentService) Compose(user User, comment Comment) (commentResult CommentBref, err error) {
	now := time.Now().Unix()

	if FrequencyLimiter.Check(user.Id, now) {
		err = e.FrequentRequest()
		return
	}

	if user.Level == USER_BLOCKED {
		err = e.UserBlocked()
		return
	}

	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	var isSensitive bool
	if user.Level == USER_DOUBTED || user.Level == USER_REVIEWING {
		// doubted user or reviewing
		comment.State = COMMENT_IMPLICATED // implicated
	} else if user.Level == USER_TRUSTED {
		// trust user (actually normal user)
		if isSensitive = this.CheckSensitiveWords(comment.Content); isSensitive {
			comment.State = COMMENT_REVIEWING // reviewing
		} else {
			comment.State = COMMENT_PASS // pass
		}
	} else {
		err = e.ValueError()
		return
	}
	comment.Time = now
	comment.User = user.Id
	comment.Id = bson.NewObjectId()
	comment.Alive = true

	FrequencyLimiter.Mark(user.Id, now)
	err = ac.C("comments").Insert(comment)

	if err != nil {
		return
	}

	commentResult = CommentBref{
		Id:      comment.Id,
		Time:    comment.Time,
		Content: comment.Content,
		State:   comment.State,
		Ref:     comment.Ref,
		User:    comment.User,
	}
	if isSensitive {
		err = UserService.reviewUser(user.Id)
	}
	return
}

func (this *commentService) ListCommentBref(blog bson.ObjectId, user *User) (comments []CommentBref, err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()
	comments = []CommentBref{}
	var condition m
	if user == nil {
		condition = m{"blog": blog, "state": COMMENT_PASS, "alive": true}
	} else {
		condition = m{
			"$or": []m{
				m{"blog": blog, "state": COMMENT_PASS, "alive": true},
				m{
					"blog":  blog,
					"alive": true,
					"user":  user.Id,
					"$or": []m{
						m{"state": COMMENT_REVIEWING},
						m{"state": COMMENT_IMPLICATED},
					},
				},
			},
		}
	}

	// 对于某个用户，他自己的、未通过审核的评论也可以看到
	err = ac.C("comments").Find(condition).All(&comments)
	return
}

func (this *commentService) ListCommentByBlog(blog bson.ObjectId) (comments []Comment, err error) {
	return this.listComment(m{"blog": blog})
}

func (this *commentService) ListCommentByUser(user bson.ObjectId) (comments []Comment, err error) {
	return this.listComment(m{"user": user})
}

func (this *commentService) ListCommentByUsers(users []bson.ObjectId) (comments []Comment, err error) {
	return this.listComment(m{"user": m{"$in": users}})
}

func (this *commentService) ListReviewingComments() (comments []Comment, err error) {
	return this.listComment(m{"state": COMMENT_REVIEWING})
}

func (this *commentService) ListRecentComments() (comments []Comment, err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	comments = []Comment{}
	err = ac.C("comments").Find(m{
		"alive": true,
		"state": m{
			"$ne": COMMENT_BLOCKED,
		},
	}).Sort("-time").Limit(50).All(&comments)
	return
}

func (this *commentService) ListBlockedComments() (comments []Comment, err error) {
	return this.listComment(m{"state": COMMENT_BLOCKED})
}

func (this *commentService) ListRemovedComments() (comments []Comment, err error) {
	return this.listComment(m{"alive": false})
}

func (this *commentService) listComment(query m) (comments []Comment, err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	comments = []Comment{}
	err = ac.C("comments").Find(query).Sort("-time").All(&comments)
	return
}

func (this *commentService) GenerateUserId(comments []Comment) (users []bson.ObjectId) {
	users = []bson.ObjectId{}
	for _, c := range comments {
		// 对不起忘了做去重
		duplicate := false
		for _, u := range users {
			if u == c.User {
				duplicate = true
				break
			}
		}
		if !duplicate {
			users = append(users, c.User)
		}
	}
	return
}

func (this *commentService) modify(selector m, modify m) (err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	_, err = ac.C("comments").UpdateAll(selector, m{"$set": modify})
	return
}

func (this *commentService) CheckSensitiveWords(content string) (isSensitive bool) {
	words := this.getSensitiveWords()
	return this.checkSensitiveWords(content, words)
}

func (this *commentService) getSensitiveWords() (words []string) {
	bytes, err := ioutil.ReadFile("./sensitive.txt")
	if err != nil {
		return
	}

	return strings.Split(string(bytes), "\n")
}

func (this *commentService) checkSensitiveWords(content string, words []string) (isSensitive bool) {
	for _, w := range words {
		if utils.RunesIndexOf(content, w) > -1 {
			return true
		}
	}
	return false
}

func (this *commentService) findAndModify(query m, modify m) (matched bool, comment Comment, err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	changeInfo, err := ac.C("comments").Find(query).Apply(mgo.Change{
		Update: m{"$set": modify},
		// ReturnNew: false,
	}, &comment)
	return changeInfo.Matched > 0, comment, err
}

func (this *commentService) deleteComment(id bson.ObjectId) (matched bool, comment Comment, err error) {
	return this.findAndModify(m{"_id": id}, m{"alive": false})
}

func (this *commentService) confirmComment(id bson.ObjectId) (matched bool, comment Comment, err error) {
	return this.findAndModify(m{"_id": id, "state": COMMENT_REVIEWING}, m{"state": COMMENT_PASS})
}

func (this *commentService) blockComment(id bson.ObjectId) (matched bool, comment Comment, err error) {
	return this.findAndModify(m{"_id": id, "state": COMMENT_REVIEWING}, m{"state": COMMENT_BLOCKED})
}

func (this *commentService) revertComment(id bson.ObjectId) (matched bool, comment Comment, err error) {
	return this.findAndModify(m{"_id": id}, m{"state": COMMENT_PASS, "alive": true})
}
func (this *commentService) reviewComment(id bson.ObjectId) (comment Comment, err error) {
	_, comment, err = this.findAndModify(m{"_id": id}, m{"state": COMMENT_REVIEWING})
	return
}

func (this *commentService) CheckUserComments(id bson.ObjectId) (sensitiveOne *Comment, err error) {
	comments, err := this.listComment(m{"user": id, "alive": true, "state": m{"$ne": COMMENT_PASS}})
	if err != nil {
		return
	}
	commentsLength := len(comments)
	if commentsLength == 0 {
		return
	}

	// 这里这个comments顺序是从新到老，检测应该从老到新
	i := 0
	li := commentsLength - i - 1
	for ; i < li; i, li = i+1, li-1 {
		comments[i], comments[li] = comments[li], comments[i]
	}

	var (
		passList                  = []bson.ObjectId{}
		sensitiveId bson.ObjectId = ""
	)

	sensitiveWords := this.getSensitiveWords()
	for _, comment := range comments {
		if comment.State != COMMENT_IMPLICATED {
			// 触发checkusercomments的用户无非两种，新用户被检测，或敏感评论审核通过之后其余评论被检测。
			// 不应该出现状态alive=true而state不是pass或implicated的评论
			err = e.ValueError()
			return
		}
		if this.checkSensitiveWords(comment.Content, sensitiveWords) {
			// 发现敏感词
			sensitiveId = comment.Id
			sensitiveOne = &comment
			sensitiveOne.State = COMMENT_REVIEWING
			break
		} else {
			// 安全
			passList = append(passList, comment.Id)
		}
	}
	if sensitiveId != "" {
		_, err = this.reviewComment(sensitiveId)
		if err != nil {
			return
		}
		err = UserService.reviewUser(id)
		if err != nil {
			return
		}
	} else {
		err = UserService.trustUser(id)
	}
	if len(passList) > 0 {

		err = this.modify(m{"_id": m{"$in": passList}}, m{"state": COMMENT_PASS})
		if err != nil {

			return
		}
	}
	return
}

func (this *commentService) DeleteComment(id bson.ObjectId) (setUserDoubted bool, err error) {
	matched, comment, err := this.deleteComment(id)
	if err != nil {
		return
	}
	if !matched {
		err = e.NotFound()
		return
	}
	if comment.State == COMMENT_REVIEWING {
		// 若删除的是审核中的评论，设置用户为未置信状态。
		err = UserService.uncertainForUser(comment.User)
		if err != nil {
			return
		}
		setUserDoubted = true
	}
	// 若删除的是已屏蔽的评论，并且用户已经没有其他被屏蔽的言论，设置用户为未置信状态
	if comment.State == COMMENT_BLOCKED {
		var comments []Comment
		comments, err = this.ListCommentByUser(comment.User)
		if err != nil {
			return
		}
		sensitiveYet := false
		for _, c := range comments {
			if c.State == COMMENT_BLOCKED && c.Alive {
				sensitiveYet = true
				break
			}
		}
		if !sensitiveYet {
			err = UserService.uncertainForUser(comment.User)
			if err != nil {
				return
			}
			setUserDoubted = true
		}
	}
	return
}

func (this *commentService) RevertComment(id bson.ObjectId) (err error) {
	matched, _, err := this.revertComment(id)
	if err != nil {
		return
	}
	if !matched {
		return e.NotFound()
	}
	return
}

func (this *commentService) BlockComment(id bson.ObjectId) (err error) {
	matched, comment, err := this.blockComment(id)
	if err != nil {
		return
	}
	if !matched {
		return e.NotFound()
	}

	// 一篇评论的严重程度已经到了屏蔽的程度，此时用户也被拉黑
	err = UserService.blockUser(comment.User)
	if err != nil {
		return
	}

	return
}

func (this *commentService) PassComment(id bson.ObjectId) (sensitiveOne *Comment, err error) {
	matched, comment, err := this.confirmComment(id)
	if err != nil {
		return
	}
	if !matched {
		err = e.NotFound()
		return
	}

	return this.CheckUserComments(comment.User)
}

// 发言频率限制
type frequencyLimiter struct {
	cache map[bson.ObjectId]*UserFrequency
}

type UserFrequency struct {
	Record  []int64
	blocker chan bool
	uid     bson.ObjectId
}

func initUserFrequency(id bson.ObjectId) *UserFrequency {
	f := new(UserFrequency)
	f.Record = make([]int64, FREQUENCY_CAP)
	f.blocker = make(chan bool)
	f.uid = id
	return f
}

func (this *UserFrequency) waitingClose() {
	select {
	case <-time.After(FREQUENCY_LIMIT):
		if _, ok := FrequencyLimiter.cache[this.uid]; ok {
			delete(FrequencyLimiter.cache, this.uid)
			close(this.blocker)
		}
		return
	case <-this.blocker:
		return
	}
}

func (this *UserFrequency) append(time int64) *UserFrequency {
	if len(this.Record) >= 5 {
		this.Record = append(this.Record[1:4], time)
	} else {
		this.Record = append(this.Record, time)
	}
	return this
}

const (
	FREQUENCY_CAP   = 5   // 每个用户的频率记录长度
	FREQUENCY_LIMIT = 300 // 五分钟内最多发五条
)

// 我们这里是为了限制发言的频率不过高
// 假设一下，先假设一下频率好了
// 假设连发五条的情况下，
var FrequencyLimiter frequencyLimiter

func (this *frequencyLimiter) Mark(id bson.ObjectId, now int64) {
	cache, ok := this.cache[id]
	if !ok || cache == nil {
		this.cache[id] = initUserFrequency(id).append(now)
		return
	}

	this.cache[id].append(now)
	go this.cache[id].waitingClose()
}

func (this *frequencyLimiter) Check(id bson.ObjectId, now int64) (isFrequent bool) {
	cache, ok := this.cache[id]
	if !ok || cache == nil {
		return false
	}
	if len(cache.Record) < 5 {
		return false
	}
	if cache.Record[0]+FREQUENCY_LIMIT <= now {
		return true
	} else {
		return false
	}
}

func (this *frequencyLimiter) init() {
	this.cache = map[bson.ObjectId]*UserFrequency{}
}
