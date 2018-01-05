package service

import (
	"drafter/db"
	e "drafter/exception"

	"gopkg.in/mgo.v2/bson"
)

// 考虑到用户可能会给笔记本改名的狗血情况，我们需要在note里记录笔记本信息的时候传一个ID而不是笔记本的名字
// 操

type Notebook struct {
	Id    bson.ObjectId `json:"id" bson:"_id"`
	Name  string        `json:"name"`
	Count int           `json:"count"`
}

type notebookService struct{}

var NotebookService = new(notebookService)

// 由于是作者操作，数据访问量并不高，所以不做缓存。每次获取数据都从数据库直接拿
func (this *notebookService) List() (nbs []Notebook, err error) {
	nbs = []Notebook{}
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("notebook").Find(nil).All(&nbs)
	return
}

func (this *notebookService) Find(id bson.ObjectId) (nb Notebook, err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("notebook").FindId(id).One(&nb)
	return
}

func (this *notebookService) Add(name string) (n Notebook, err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	C := ac.C("notebook")
	count, err := C.Find(m{"name": name}).Count()
	if err != nil {
		return
	}
	if count != 0 {
		err = e.NotebookAlreadyExist(name)
		return
	}

	n = Notebook{Id: bson.NewObjectId(), Name: name}
	err = C.Insert(n)
	return
}

func (this *notebookService) Rename(id bson.ObjectId, name string) (err error) {
	// if !bson.IsObjectIdHex(id) {
	// 	err = e.InvalidId(id)
	// 	return
	// }

	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("notebook").Update(m{"_id": id}, m{"$set": m{"name": name}})

	return
}

type counterUpdator map[bson.ObjectId]int

func (this *notebookService) UpdateAmount(update counterUpdator) (err error) {

	// for id, _ := range update {
	// 	if !bson.IsObjectIdHex(id) {
	// 		err = e.InvalidId(id)
	// 		return
	// 	}
	// }

	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	C := ac.C("notebook")
	for id, count := range update {
		_, err = C.Upsert(m{"_id": id}, m{"$inc": m{"count": count}})
		if err != nil {
			return
		}
	}

	return
}

func (this *notebookService) Remove(id bson.ObjectId) (err error) {
	// if !bson.IsObjectIdHex(notebookid) {
	// 	err = e.InvalidId(notebookid)
	// 	return
	// }

	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	// id := bson.ObjectIdHex(notebookid)
	//  将原本该笔记本的内容移除到回收站
	_, err = ac.C("note").UpdateAll(m{"notebookid": id}, m{"$set": m{"alive": false}})

	if err != nil {
		return
	}

	err = ac.C("notebook").Remove(m{"_id": id})
	return
}

// 用户在删除一个笔记本时，删除到 1.另一个笔记本 2.垃圾箱 3.彻底删除
// 灵活的体验交给前端，接口要负责功能的单一。所以直接放到垃圾箱。并且，过期15天的笔记会自动删除
// func (this *notebookService) Remove(name string) (err error) {
// 	ac, err := db.Access()
// 	if err != nil {
// 		return
// 	}
// 	defer ac.Close()

// 	err = ac.C("notebook").Remove(m{"name": name})
// 	if err != nil {
// 		return
// 	}

// 	_, err = ac.C("note").UpdateAll(m{"notebook": name}, m{"$set": m{"notebook": "", "alive": false}})
// 	// if err != nil {
// 	// 	return
// 	// }
// 	return
// }

// func (this *notebookService) UpdateAmount(modifition map[string]int) (err error) {
// 	ac, err := db.Access()
// 	if err != nil {
// 		return
// 	}
// 	defer ac.Close()

// 	C := ac.C("notebook")

// 	for name, count := range modifition {
// 		err = C.Update(m{"name": name}, m{"$inc": m{"count": count}})
// 		if err != nil {
// 			return
// 		}
// 	}
// 	return
// }

// func (this *notebookService) Rename(from string, to string) {

// }

// func (this *notebookService) Exec(selector interface{}, update interface{}) (info *mgo.ChangeInfo, err error) {
// 	ac, err := db.Access()
// 	if err != nil {
// 		return
// 	}
// 	defer ac.Close()

// 	return ac.C("notebook").UpdateAll(selector, update)
// }
