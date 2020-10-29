package models

import (
	"fmt"

	"github.com/ICKelin/cframe/pkg/database"
	"gopkg.in/mgo.v2/bson"
)

type AlarmType int32

var (
	C_ALARM = "alarm"
)

var (
	AlarmHappen   = AlarmType(1)
	AlarmHandling = AlarmType(2)
	AlarmHandled  = AlarmType(3)
)

type Alarm struct {
	database.Model `bson:",inline"`
	UserId         bson.ObjectId `json:"userId" bson:"userId"`
	EdgeName       string        `json:"edgeName" bson:"edgeName"`
	Content        string        `json:"content" bson:"content"`
	AlarmType      AlarmType     `json:"alarmType" bson:"alarmType"`
	Status         int           `json:"status" bson:"status"`
	Comment        string        `json:"comment" bson:"comment"`
}

type AlarmManager struct {
	database.ModelManager
}

func GetAlarmManager() *AlarmManager {
	return &AlarmManager{}
}

func (m AlarmManager) AddAlarm(alarm *AlarmManager) (*Alarm, error) {
	return nil, fmt.Errorf("TODO://")
}

func (m *AlarmManager) UpdateAlarm(userId, alarmId bson.ObjectId,
	nstatus int, comment string) error {
	return fmt.Errorf("TODO://")
}
