package typecho

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Plugin struct {
	Package     string
	Name        string
	Description string
	Author      string
	Version     string
	Link        string
	Require     string
	Source      string
	Readme      string
	Type        int
}

func Parse(path, packageName, repo string, retry bool) Plugin {

	plugin := Plugin{
		"",
		"",
		"",
		"",
		"",
		"",
		"*",
		"",
		"",
		0,
	}

	pluginContent, err := ioutil.ReadFile(path)
	if err != nil {
		if retry {
			os.Rename(strings.Replace(path, "Plugin.php", "plugin.php", 1), path)
			plugin = Parse(path, packageName, repo, false)
		}
		return plugin
	}

	reString := `/\*\*([\s\S]*?)\*/`
	re, _ := regexp.Compile(reString)
	matches := re.FindAllString(string(pluginContent), -1)
	wantedMatch := ""

	for _, match := range matches {
		if strings.Contains(match, "@package") {
			wantedMatch = match
			break
		}
	}

	lines := strings.Split(wantedMatch, "\n")

	for _, line := range lines {
		line = strings.Replace(line, "/**", "", -1)
		line = strings.Replace(line, "*/", "", 1)
		line = strings.Replace(line, "*", "", 1)
		line = strings.Trim(line, " ")
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "@package") {
			plugin.Name = strings.Trim(strings.Replace(line, "@package", "", 1), " ")
			continue
		}
		if strings.HasPrefix(line, "@author") {
			plugin.Author = strings.Trim(strings.Replace(line, "@author", "", 1), " ")
			continue
		}
		if strings.HasPrefix(line, "@version") {
			plugin.Version = strings.Trim(strings.Replace(line, "@version", "", 1), " ")
			continue
		}
		if strings.HasPrefix(line, "@link") {
			plugin.Link = strings.Trim(strings.Replace(line, "@link", "", 1), " ")
			continue
		}
		if strings.HasPrefix(line, "@dependence") {
			plugin.Require = strings.Trim(strings.Replace(line, "@dependence", "", 1), " ")
			continue
		}
		if !strings.HasPrefix(line, "@") {
			plugin.Description = plugin.Description + line
		}

	}
	plugin.Source = repo
	plugin.Package = packageName

	//读取readme.md

	pluginPath := filepath.Dir(path)
	readmeName := ""
	if _, err = os.Stat(pluginPath + "/README.md"); err == nil {
		readmeName = "README.md"
	} else if _, err = os.Stat(pluginPath + "/Readme.md"); err == nil {
		readmeName = "Readme.md"
	} else if _, err = os.Stat(pluginPath + "/readme.md"); err == nil {
		readmeName = "readme.md"
	} else if _, err = os.Stat(pluginPath + "/README"); err == nil {
		readmeName = "README"
	}

	if readme, err := ioutil.ReadFile(pluginPath + "/" + readmeName); err == nil {
		plugin.Readme = string(readme)
	}

	if !retry {
		os.Remove(path)
	}
	return plugin
}
