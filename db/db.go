package db

import (
	. "drafter/setting"
	"drafter/utils/pool"
	"log"

	"gopkg.in/mgo.v2"
)

var (
	sessionProto *mgo.Session
)

type Accessor struct {
	db      *mgo.Database
	session *mgo.Session
}

var p *pool.Pool

func Access() (*Accessor, error) {
	session, err := p.Acquire()
	if err != nil {
		return nil, err
	}
	
	return &Accessor{db: session.DB(Settings.DB.Database), session: session}, nil
}

func (ac *Accessor) Close() {
	p.Release(ac.session)
}

func (ac *Accessor) C(name string) *mgo.Collection {
	return ac.db.C(name)
}

func init() {

	var err error
	p, err = pool.CreatePool(Settings.DB.URL, 10, 1000)
	if err != nil {
		log.Fatal(err)
	}

}
