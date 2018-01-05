package service

import (
	"drafter/db"
	e "drafter/exception"
	"encoding/json"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type blogService struct {
	ValidCount int
}

var BlogService = new(blogService)

type Blog struct {
	Id       bson.ObjectId `json:"id" bson:"_id"`
	Title    string        `json:"title"`
	Content  string        `json:"content"`
	Html     string        `json:"html"`
	Tags     []string      `json:"tags"`
	CreateAt int           `json:"createAt"`
	EditAt   int           `json:"editAt"`
	Noteid   bson.ObjectId `json:"noteid"`
	Alive    bool          `json:"alive"`
}

// func (this *blogService) Compose(n Note) {
// 比较尴尬的是，目前还没写出在golang里实现markdown和mathJax的程序，所以我们只能让node中间层帮我们做这件事
func (this *blogService) Compose(b *Blog) (err error) {
	// func (this *blogService) Compose(title string, content string, tags []string, html string, noteid bson.ObjectId) (err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	b.Id = bson.NewObjectId()
	now := int(time.Now().Unix() * 1000)
	b.CreateAt = now
	b.EditAt = now
	b.Alive = true
	err = ac.C("blog").Insert(*b)
	go TagService.ModifyAmount(b.Tags, []string{})
	go this.ResetCount()
	return err
}

type blogEditionTemplate struct {
	Title   string        `json:"title"`
	Content string        `json:"content"`
	Html    string        `json:"html"`
	Tags    []string      `json:"tags"`
	Noteid  bson.ObjectId `json:"noteid"`
	Alive   bool          `json:"alive"`
}

func (this *blogService) Edit(id bson.ObjectId, rawEdition json.RawMessage) (err error) {
	edition := make(map[string]interface{})
	var ok bool
	if edition, ok = UnmarshalValidMap(rawEdition, new(blogEditionTemplate)); !ok {
		err = e.InvalidEdition()
		return
	}
	return this.edit(id, edition)
}

// func (this *blogService) EditFromNote(id bson.ObjectId, noteid bson.ObjectId) (err error) {
// 	n, err := NoteService.Get([]bson.ObjectId{noteid})
// 	if err != nil {
// 		return
// 	}
// 	if len(n) == 0 {
// 		err = e.NoteDoNotExist(noteid)
// 		return
// 	}
// 	note := n[0]
// 	return this.edit(id, map[string]interface{}{
// 		"title":   note.Title,
// 		"content": note.Content,
// 		"tags":    note.Tags,
// 		"noteid":  note.Id,
// 		"html":    marked(note.Content),
// 		// keep alive as also
// 	})
// }

func (this *blogService) edit(id bson.ObjectId, edition map[string]interface{}) (err error) {
	var newTags []string
	if tags, ok := edition["tags"]; ok {
		newTags = tags.([]string)
	}

	if mapIncludesOne(edition, []string{"title", "tags", "content", "html"}) {
		edition["editat"] = time.Now().Unix() * 1000
	}

	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	var old Blog
	_, err = ac.C("blog").FindId(id).Apply(mgo.Change{
		Update:    m{"$set": edition},
		Upsert:    false,
		ReturnNew: false,
	}, &old)

	if err != nil {
		return
	}

	// 修改tags数量前判断alive情况
	if alive, ok := edition["alive"]; ok {
		aliveNow := alive.(bool)
		if old.Alive && !aliveNow {
			go TagService.ModifyAmount([]string{}, old.Tags)
		} else if aliveNow && !old.Alive {
			if newTags != nil {
				go TagService.ModifyAmount(newTags, []string{})
			} else {
				go TagService.ModifyAmount(old.Tags, []string{})
			}
		} else if old.Alive && aliveNow {
			if newTags != nil {
				go TagService.ModifyAmount(newTags, old.Tags)
			}
		}
		go this.ResetCount()
	} else {
		// 没有修改alive
		// 其实是不可能的吧
		// 每次提交数据的时候啥都一起提交过来了
		// 除非我在前端再加控制过滤完全匹配的tags
		if newTags != nil && old.Alive {
			go TagService.ModifyAmount(newTags, old.Tags)
		}
	}

	return
}

func (this *blogService) Remove(id bson.ObjectId) (err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("blog").Remove(m{"_id": id, "alive": false})
	return
}

type BlogBref struct {
	Id       bson.ObjectId `json:"id" bson:"_id"`
	Title    string        `json:"title"`
	Tags     []string      `json:"tags"`
	CreateAt int           `json:"createAt"`
	EditAt   int           `json:"editAt"`
	Noteid   bson.ObjectId `json:"noteid"`
	Alive    bool          `json:"alive"`
}

func (this *blogService) List(skip int, limit int) (blogs []BlogBref, err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	query := ac.C("blog").Find(nil).Sort("-createat").Skip(skip)
	if limit != 0 {
		query = query.Limit(limit)
	}
	blogs = []BlogBref{}

	err = query.All(&blogs)
	return
}

func (this *blogService) Get(ids []bson.ObjectId) (blogs []Blog, err error) {
	blogs = []Blog{}
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("blog").Find(m{"_id": m{"$in": ids}}).All(&blogs)
	return
}

func (this *blogService) RemoveTag(ac *db.Accessor, tag string) (updated int, err error) {
	info, err := ac.C("blog").UpdateAll(m{"tags": m{"$all": []string{tag}}}, m{"$pull": m{"tags": tag}})
	if err == nil {
		updated = info.Updated
	}
	return
}

func (this *blogService) ResetCount() (err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	count, err := ac.C("blog").Find(m{"alive": true}).Count()
	if err != nil {
		return
	}
	this.ValidCount = count
	return
}
