package models

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chekun/echoed/typecho"
	"github.com/chekun/echoed/typecho/ziputil"
)

type Plugin struct {
	Id        int
	Package   string
	Type      int
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   []*Version
}

func (p *Plugin) TableName() string {
	return "plugins"
}

type Version struct {
	Id          int
	Name        string
	PluginId    int
	Plugin      *Plugin `gorm:"ForeignKey:PluginId"`
	Author      string
	Version     string
	Description string
	Link        string
	Require     string
	Readme      string
	Downloads   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (v *Version) TableName() string {
	return "plugin_versions"
}

func UpdatePlugin(p *typecho.Plugin) {
	if IsPluginExisted(p.Package) {
		AppendVersion(p)
	} else {
		AddNewPlugin(p)
	}
}

func GetAllPlugins() []*Plugin {
	var plugins []*Plugin
	db.Preload("Version").Find(&plugins)
	return plugins
}

func AddNewPlugin(p *typecho.Plugin) {
	plugin := new(Plugin)
	plugin.Package = p.Package
	plugin.Type = p.Type
	if db.Create(&plugin); !db.NewRecord(plugin) {
		AppendVersion(p)
	}
}

func cleanString(str string) string {
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\n", "", -1)
	str = strings.Replace(str, "\r", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	return str
}

func AppendVersion(p *typecho.Plugin) {
	plugin := FindPlugin(p.Package)
	if !IsVersionExisted(plugin.Id, p.Version) {
		version := new(Version)
		version.Plugin = plugin
		version.Author = cleanString(p.Author)
		version.Name = cleanString(p.Name)
		version.Description = cleanString(p.Description)
		version.Link = cleanString(p.Link)
		version.Require = cleanString(p.Require)
		version.Version = cleanString(p.Version)
		version.Readme = p.Readme
		if db.Create(&version); !db.NewRecord(version) {
			zipPlugin(cleanString(p.Package), cleanString(p.Version), p.Source, p.Type)
			db.Model(&plugin).Update("updated_at", time.Now())
		}
	}
}

func FindPlugin(name string) *Plugin {
	plugin := new(Plugin)
	db.Where("package = ?", name).First(&plugin)
	return plugin
}

func IsPluginExisted(name string) bool {
	plugin := new(Plugin)
	var count int
	db.Model(plugin).Where("package = ?", name).Count(&count)
	return count > 0
}

func IsVersionExisted(pluginId int, v string) bool {
	version := new(Version)
	var count int
	db.Model(version).Where("plugin_id = ?", pluginId).Count(&count)
	return count > 0
}

func CountVersionDownload(name, version string) {
	versionObj := new(Version)
	if err := db.Model(versionObj).Where("`name` = ? AND `version` = ?", name, version).First(&versionObj).Error; err == nil {
		versionObj.Downloads++
		db.Save(&versionObj)
	}
}

func zipPlugin(name, version, repo string, pType int) {

	storagePath := os.Getenv("PACKAGES_DIRECTORY")
	repoPath := os.Getenv("WORKING_DIRECTORY")

	zipFile := ""
	if pType == 0 {
		zipFile = storagePath + "/plugins/" + name + "/" + name + "-" + version + ".zip"
	} else {
		zipFile = storagePath + "/themes/" + name + "/" + name + "-" + version + ".zip"
	}
	directory := repoPath + "/" + repo + "/" + name + "/"

	err := ziputil.Zip(zipFile, directory)
	if err != nil {
		fmt.Println(name, version, err)
	}
}

type VersionJson struct {
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Require     string    `json:"require"`
	CreatedAt   time.Time `json:"created_at"`
}

type PluginJson struct {
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	Versions []VersionJson `json:"versions"`
}

type PluginsJson struct {
	Packages []PluginJson `json:"packages"`
}

func ToJson() ([]byte, error) {
	var pluginsJson PluginsJson
	plugins := GetAllPlugins()
	for _, plugin := range plugins {
		var pluginJson PluginJson
		pluginJson.Name = plugin.Package
		if plugin.Type == 0 {
			pluginJson.Type = "plugin"
		} else {
			pluginJson.Type = "theme"
		}

		for _, version := range plugin.Version {
			versionJson := VersionJson{}
			versionJson.Version = version.Version
			versionJson.Author = version.Author
			versionJson.CreatedAt = version.CreatedAt
			versionJson.Description = version.Description
			versionJson.Link = version.Link
			versionJson.Require = version.Require

			pluginJson.Versions = append(pluginJson.Versions, versionJson)
		}

		pluginsJson.Packages = append(pluginsJson.Packages, pluginJson)
	}

	bytes, err := json.Marshal(pluginsJson)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}
