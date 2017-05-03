package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/chekun/echoed/typecho"
	"github.com/chekun/echoed/typecho/models"
	"github.com/joho/godotenv"
)

func runCommand(cmd *exec.Cmd) {
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("%s\n", string(stdout))
}

func gitWork(repositoryPath, repository string) string {
	os.Chdir(repositoryPath)

	repoFolder := strings.Replace(repository, "/", "-", -1)

	if _, err := os.Stat(repoFolder); os.IsNotExist(err) {
		//文件夹不存在，全新clone
		gitPath := "https://github.com/" + repository + ".git"
		log.Println("cloning", gitPath)
		cmd := exec.Command("git", "clone", gitPath, repoFolder)
		runCommand(cmd)
		//处理submodule
		os.Chdir(repoFolder)
		log.Println("git submodule update --init")
		cmd = exec.Command("git", "submodule", "update", "--init")
		runCommand(cmd)
		os.Chdir("../")
	}

	os.Chdir(repoFolder)

	log.Println("git stash")
	cmd := exec.Command("git", "stash")
	runCommand(cmd)
	log.Println("git fetch origin")
	cmd = exec.Command("git", "fetch", "origin")
	runCommand(cmd)
	log.Println("git merge origin/master")
	cmd = exec.Command("git", "merge", "origin/master")
	runCommand(cmd)
	log.Println("git submodule udpate --init")
	cmd = exec.Command("git", "submodule", "update", "--init")
	runCommand(cmd)
	log.Println("git submodule foreach git pull origin master")
	cmd = exec.Command("git", "submodule", "foreach", "git", "pull", "origin", "master")
	runCommand(cmd)
	os.Chdir("../")

	return repoFolder
}

func main() {

	binPath, _ := os.Getwd()

	os.Setenv("ROOT_PATH", binPath+"/../")

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	models.NewDb()

	repositoryPath := os.Getenv("WORKING_DIRECTORY")

	//插件处理
	pluginRepository := os.Getenv("PLUGIN_REPOSITORY")

	repoFolder := gitWork(repositoryPath, pluginRepository)

	files, _ := ioutil.ReadDir(repoFolder)
	for _, file := range files {
		fileName := file.Name()
		if fileName == ".gitignore" || fileName == ".git" || fileName == ".gitattributes" || fileName == ".gitmodules" {
			continue
		}
		if file.IsDir() {
			//插件解析Plugin.php中的信息，然后打包待下载
			plugin := typecho.Parse(repoFolder+"/"+fileName+"/Plugin.php", fileName, repoFolder, true)
			if plugin.Package != "" {
				models.UpdatePlugin(&plugin)

			}
			continue
		}
	}

	//主题处理
	themeRepository := os.Getenv("THEME_REPOSITORY")
	repoFolder = gitWork(repositoryPath, themeRepository)

	files, _ = ioutil.ReadDir(repoFolder)
	for _, file := range files {
		fileName := file.Name()
		if fileName == ".gitignore" || fileName == ".git" || fileName == ".gitattributes" || fileName == ".gitmodules" {
			continue
		}
		if file.IsDir() {
			//主题解析index中的信息，然后打包待下载
			plugin := typecho.Parse(repoFolder+"/"+fileName+"/index.php", fileName, repoFolder, true)
			plugin.Type = 1
			if plugin.Name != "" {
				models.UpdatePlugin(&plugin)
			}
			continue
		}
	}

	json, err := models.ToJson()
	if err != nil {
		log.Println("write json failed")
	}
	ioutil.WriteFile(binPath+"/../storage/packages.json", json, 0755)

	log.Println("Sync Succeeded!")
}
