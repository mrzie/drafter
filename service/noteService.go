package service

import (
	"drafter/db"
	e "drafter/exception"
	"time"

	"encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type noteService struct{}

var NoteService = new(noteService)

type Note struct {
	Id         bson.ObjectId `json:"id" bson:"_id"`
	Title      string        `json:"title"`
	Content    string        `json:"content"`
	Abstract   string        `json:"abstract"`
	Tags       []string      `json:"tags"`
	Notebookid bson.ObjectId `json:"notebookid"`
	CreateAt   int64         `json:"createAt"`
	EditAt     int64         `json:"editAt"`
	Alive      bool          `json:"alive"`
}

type NoteBref struct {
	Id         bson.ObjectId `json:"id" bson:"_id"`
	Title      string        `json:"title"`
	Abstract   string        `json:"abstract"`
	Tags       []string      `json:"tags"`
	Notebookid bson.ObjectId `json:"notebookid"`
	CreateAt   int64         `json:"createAt"`
	EditAt     int64         `json:"editAt"`
	Alive      bool          `json:"alive"`
}

func (this *noteService) List(notebookid bson.ObjectId) (notes []NoteBref, err error) {
	notes = []NoteBref{}
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("note").Find(m{"notebookid": notebookid, "alive": true}).All(&notes)

	return
}

func (this *noteService) ListWaste() (notes []NoteBref, err error) {
	notes = []NoteBref{}
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("note").Find(m{"alive": false, "editat": m{"$gt": time.Now().Unix() * 1000 - 30 * 24 * 3600 * 1000}}).All(&notes)
	return
}

func (this *noteService) Get(ids []bson.ObjectId) (notes []Note, err error) {
	notes = []Note{}
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	err = ac.C("note").Find(m{"_id": m{"$in": ids}}).All(&notes)
	return
}

func (this *noteService) Remove(id bson.ObjectId) (err error) {
	// ac, err := db.Access()
	// if err != nil {
	// 	return
	// }
	// defer ac.Close()

	// err = ac.C("note").Remove(m{"_id": id, "alive": false})
	edition, _ := json.Marshal(map[string]interface{}{"alive": false})

	err = this.Edit(id, edition)
	return
}

func (this *noteService) Compose(n Note) (note Note, err error) {
	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	count, err := ac.C("notebook").FindId(n.Notebookid).Count()
	if count == 0 {
		err = e.NotebookDoNotExist(n.Notebookid)
		return
	}

	n.Id = bson.NewObjectId()
	abstractLen := 200
	if len([]rune(n.Content)) < abstractLen {
		abstractLen = len([]rune(n.Content))
	}
	n.Abstract = string([]rune(n.Content)[:abstractLen])
	now := time.Now().Unix() * 1000
	n.CreateAt = now
	n.EditAt = now
	n.Alive = true

	err = ac.C("note").Insert(n)
	if err != nil {
		return
	}

	go NotebookService.UpdateAmount(counterUpdator{n.Notebookid: 1})
	note = n
	return
}

type noteEditionTemplate struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Tags       []string `json:"tags"`
	Notebookid string   `json:"notebookid"`
	Alive      bool     `json:"alive"`
}

// 在controller层parseMap?
func (this *noteService) Edit(id bson.ObjectId, rawEdition json.RawMessage) (err error) {

	// if !mapValid(edition, noteEditionTemplate{}) {
	// 	err = e.InvalidEdition()
	// 	return
	// }
	edition := make(map[string]interface{})
	var ok bool
	if edition, ok = UnmarshalValidMap(rawEdition, new(noteEditionTemplate)); !ok {
		err = e.InvalidEdition()
		return
	}

	if content, ok := edition["content"]; ok {

		abstractLen := 200
		if len([]rune(content.(string))) < abstractLen {
			abstractLen = len([]rune(content.(string)))
		}
		edition["abstract"] = string([]rune(content.(string))[:abstractLen])
	}

	// 当title content 或 tags被修改时才修改editAt
	if mapIncludesOne(edition, []string{"title", "content", "tags", "alive"}) {
		edition["editat"] = time.Now().Unix() * 1000
	}

	ac, err := db.Access()
	if err != nil {
		return
	}
	defer ac.Close()

	notebook, triedToEditNotebook := edition["notebookid"]
	if triedToEditNotebook {
		// 如果设置了新的notebook（此处可能是新增、删除恢复、更换笔记本
		// 保证不会出现死笔记
		id := notebook.(string)
		if !bson.IsObjectIdHex(id) {
			return e.InvalidId(id)
		}
		_, err = NotebookService.Find(bson.ObjectIdHex(id))
		if err != nil {
			// 这个新的Notebook其实不存在
			return e.InvalidId(id)
		}
		edition["notebookid"] = bson.ObjectIdHex(notebook.(string))
	}

	var old Note
	_, err = ac.C("note").FindId(id).Apply(mgo.Change{
		Update:    m{"$set": edition},
		ReturnNew: false,
	}, &old)

	if err != nil {
		return
	}

	var newNoteBook bson.ObjectId

	// 逻辑陷阱：可能会将笔记本改为一个不存在的Object.Id，此时无法再检索出笔记
	if triedToEditNotebook {
		// if newNoteBook, ok = notebook.(bson.ObjectId); !ok {
		// 	err = e.InvalidEdition()
		// 	return
		// }
		// newNoteBook = notebook.(bson.ObjectId)
		newNoteBook = edition["notebookid"].(bson.ObjectId)
	} else {
		newNoteBook = old.Notebookid
	}

	// 当alive修改，或notebookid修改时，更新Notebook的计数

	notebookModify := make(counterUpdator)
	var newAlive bool
	if v, ok := edition["alive"]; ok {
		newAlive = v.(bool)
	} else {
		newAlive = old.Alive
	}

	if newAlive == old.Alive && newNoteBook == old.Notebookid {
		// 没有任何修改，不用更新笔记数量
	} else {
		if newAlive {
			notebookModify[newNoteBook] += 1
		}
		if old.Alive {
			notebookModify[old.Notebookid] -= 1
		}
		go NotebookService.UpdateAmount(notebookModify)
	}
	return
}
