package service

import (
	"drafter/db"
	e "drafter/exception"
)

type tagService struct {
	// 显示在首页的tag云
	CounterCache map[string]int
}

var TagService = new(tagService)

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Count       int    `json:"count"`
}

func (this *tagService) Get(name string) (t Tag, err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("tag").Find(m{"name": name}).One(&t)
	return
}

func (this *tagService) List() (tags []Tag, err error) {
	tags = []Tag{}
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("tag").Find(nil).All(&tags)
	return
}

func (this *tagService) SetDescription(name string, description string) (err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("tag").Update(m{"name": name}, m{"$set": m{"description": description}})
	return
}

func (this *tagService) ModifyAmount(add []string, reduce []string) (err error) {
	// 我想未来还是要整理一下，新老数组里相同的项就不算好了
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	// var (
	// 	matchedAdd, matchedReduce []int
	// 	// i, j int
	// )

	// for i, ad := range add {
	// 	for j, rd := range reduce {
	// 		if ad == rd {
	// 			matchedAdd = append(matchedAdd, i)
	// 			matchedReduce = append(matchedReduce, j)
	// 		}
	// 	}
	// }

	C := ac.C("tag")
	_, err = C.UpdateAll(m{"name": m{"$in": reduce}}, m{"$inc": m{"count": -1}})
	if err != nil {
		return
	}

	// .update(name:{$in:[]}, update, true)无法为每条数组内的数据都新增一个项
	for _, t := range add {
		_, err = C.Upsert(m{"name": t}, m{"$inc": m{"count": 1}})
		if err != nil {
			return
		}
	}

	// 成功后更新tag云
	return this.UpdateCounterCache()
}

func (this *tagService) UpdateCounterCache() (err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	var tags []Tag
	err = ac.C("tag").Find(nil).All(&tags)
	if err != nil {
		return
	}

	counter := make(map[string]int)
	for _, t := range tags {
		counter[t.Name] = t.Count
	}
	this.CounterCache = counter
	return
}

// 删除一个标签的同时，删除其在博客中的引用
func (this *tagService) RemoveTag(name string) (err error) {
	count, ok := this.CounterCache[name]
	if !ok {
		err = e.TagDoNotExist(name)
		return
	}

	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	if count > 0 {
		_, err = BlogService.RemoveTag(ac, name)
		if err != nil {
			return
		}
	}

	err = ac.C("tag").Remove(m{"name": name})
	go this.UpdateCounterCache()
	return
}
