package service

import (
	"drafter/db"
	e "drafter/exception"

	"drafter/utils/tasks"

	"gopkg.in/mgo.v2/bson"
)

type readerService struct{}

var ReaderService = new(readerService)

type BlogPresent struct {
	Id       bson.ObjectId `json:"id" bson:"_id"`
	Title    string        `json:"title"`
	Content  string        `json:"content" bson:"html"`
	Tags     []string      `json:"tags"`
	CreateAt int           `json:"createAt"`
	EditAt   int           `json:"editAt"`
}

func (this *readerService) GetBlogPresent(id bson.ObjectId) (b BlogPresent, err error) {

	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("blog").FindId(id).One(&b)
	return
}

func (this *readerService) ListBlogPresent(page int) (blogs []BlogPresent, err error) {
	pageSize := BlogPreference.PageSize
	blogs = []BlogPresent{}
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	if page < 1 {
		err = e.NotFound()
		return
	}

	skip := (page - 1) * pageSize
	err = ac.C("blog").Find(m{"alive": true}).Sort("-createat").Skip(skip).Limit(pageSize).All(&blogs)
	return
}

func (this *readerService) ListBlogPresentByTag(tag string, page int) (blogs []BlogPresent, err error) {
	blogs = []BlogPresent{}
	pageSize := BlogPreference.PageSize
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	if page < 1 {
		err = e.NotFound()
		return
	}

	skip := (page - 1) * pageSize
	err = ac.C("blog").Find(m{"alive": true, "tags": m{"$all": []string{tag}}}).Sort("-createat").Skip(skip).Limit(pageSize).All(&blogs)
	return
}

func (this *readerService) TagInfo(name string) (t Tag, err error) {
	return TagService.Get(name)
}

func (this *readerService) TagCounter() map[string]int {
	return TagService.CounterCache
}

// return {"blogs": , "tag": }
// 2017-12-26 add "count"
func (this *readerService) ListBlogPresents(page int, tag string, showTagInfo bool) (result map[string]interface{}, err error) {
	var blogs []BlogPresent

	result = make(map[string]interface{})

	if tag == "" {
		blogs, err = this.ListBlogPresent(page)
		result["blogs"] = blogs
		result["count"] = BlogService.ValidCount
		return
	}

	t := tasks.NewTasks()
	t.Add(1)
	go func() {
		blogs, err := this.ListBlogPresentByTag(tag, page)
		if err != nil {
			t.Panic(err)
		}
		result["blogs"] = blogs
		t.Done()
	}()

	if showTagInfo {
		t.Add(1)
		go func() {
			tag, err := this.TagInfo(tag)
			if err != nil {
				t.Panic(err)
			}
			result["tag"] = tag
			result["count"] = tag.Count
			t.Done()
		}()
	}

	err = t.Wait()

	return
}
