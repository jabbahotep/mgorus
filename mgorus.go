package mgorus

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type hooker struct {
	c *mgo.Collection
}

type M bson.M

func NewHooker(mgoUrl, db, collection string, cleanup bool, key []mgo.Index) (*hooker, error) {
	session, err := mgo.Dial(mgoUrl)
	if err != nil {
		return nil, err
	}
    c := session.DB(db).C(collection)

    if cleanup {
        c.DropCollection()
    }

	for x := range key {
		err = c.EnsureIndex(key[x])
		if err != nil {
			panic(err)
		}
	}

	return &hooker{c: c}, nil
}

func (h *hooker) Fire(entry *logrus.Entry) error {
	entry.Data["Level"] = entry.Level.String()
	entry.Data["Time"] = entry.Time
	entry.Data["Message"] = entry.Message
	if errData, ok := entry.Data[logrus.ErrorKey]; ok {
		if err, ok := errData.(error); ok && entry.Data[logrus.ErrorKey] != nil {
			entry.Data[logrus.ErrorKey] = err.Error()
		}
	}
	mgoErr := h.c.Insert(M(entry.Data))
	if mgoErr != nil {
		return fmt.Errorf("Failed to send log entry to mongodb: %s", mgoErr)
	}

	return nil
}

func (h *hooker) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}
