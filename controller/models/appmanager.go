package models

import (
	"encoding/json"
	"fmt"

	"github.com/ICKelin/cframe/pkg/etcdstorage"
	log "github.com/ICKelin/cframe/pkg/logs"
)

var (
	appPrefix = "/apps/"
)

type App struct {
	Secret string
}

type AppManager struct {
	storage *etcdstorage.Etcd
}

func NewAppManager(store *etcdstorage.Etcd) *AppManager {
	return &AppManager{
		storage: store,
	}
}

func (m *AppManager) AddApp(app *App) error {
	key := fmt.Sprintf("%s%s", appPrefix, app.Secret)
	return m.storage.Set(key, app)
}

func (m *AppManager) DelApp(secret string) error {
	key := fmt.Sprintf("%s%s", appPrefix, secret)
	m.storage.Del(key)
	return nil
}

func (m *AppManager) GetApp(secret string) (*App, error) {
	key := fmt.Sprintf("%s%s", appPrefix, secret)
	app := App{}
	err := m.storage.Get(key, &app)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (m *AppManager) GetApps() []*App {
	key := fmt.Sprintf("%s", appPrefix)
	res, err := m.storage.List(key)
	if err != nil {
		log.Error("list %s fail: %v", edgePrefix, err)
		return nil
	}

	apps := make([]*App, 0)
	for _, val := range res {
		r := App{}
		err := json.Unmarshal([]byte(val), &r)
		if err != nil {
			log.Error("unmarshal to edge fail: %v", err)
			continue
		}
		apps = append(apps, &r)
	}
	return apps
}
