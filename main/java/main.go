package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	projectPath = ""
)

func init() {
	flag.StringVar(&projectPath, "d", projectPath, "project path")
	flag.Parse()
}

type MavenDependency struct {
	groupId    string
	artifactId string
	version    string
	packaging  string

	dirPath  string
	filePath string
}

func main() {
	if projectPath == "" {
		exit("place input project path param: -d <project path>", nil)
	}

	projectDir, err := os.ReadDir(projectPath)
	if err != nil {
		exit("read project dir error", err)
	}

	// 判断是否是 maven 项目
	isMavenProject := false
	for _, file := range projectDir {
		if !file.IsDir() && file.Name() == "pom.xml" {
			isMavenProject = true
		}
	}

	mavenDependencyList := make([]MavenDependency, 0)
	if isMavenProject {
		// 获取所有依赖
		cmd := &exec.Cmd{
			Path: fmt.Sprintf("%s%cbin%cmvn", os.Getenv("MAVEN_HOME"), os.PathSeparator, os.PathSeparator),
			Args: []string{"mvn", "dependency:list"},
			Dir:  projectPath,
		}
		pipe, err := cmd.StdoutPipe()
		if err != nil {
			exit("pipe maven dependency list error", err)
		}
		if err := cmd.Start(); err != nil {
			exit("start maven dependency list error", err)
		}
		mavenResult := make([]byte, 0)
		buf := make([]byte, 64*1024)
		for {

			readLen, err := pipe.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				exit("read maven dependency list error", err)
			}
			mavenResult = append(mavenResult, buf[:readLen]...)
		}
		rows := strings.Split(string(mavenResult), "\n")
		start := false
		for _, row := range rows {
			if start {
				row = strings.TrimSpace(row[strings.Index(row, "]")+1:])
				if len(row) == 0 {
					break
				}
				fmt.Println(row)
				split := strings.Split(row, ":")
				var mavenDependency MavenDependency = MavenDependency{}
				if len(split) == 5 {
					mavenDependency = MavenDependency{groupId: split[0], artifactId: split[1], version: split[3], packaging: split[2]}
				} else if len(split) == 6 {
					mavenDependency = MavenDependency{groupId: split[0], artifactId: split[1], version: split[4], packaging: split[2]}
				}
				dirs := strings.Split(mavenDependency.groupId, ".")
				for _, dir := range dirs {
					mavenDependency.dirPath += fmt.Sprintf("%s%c", dir, os.PathSeparator)
				}
				mavenDependency.dirPath += mavenDependency.artifactId
				mavenDependency.filePath = fmt.Sprintf("%s%c%s%c%s-%s.%s", mavenDependency.dirPath, os.PathSeparator, mavenDependency.version, os.PathSeparator, mavenDependency.artifactId, mavenDependency.version, mavenDependency.packaging)
				mavenDependencyList = append(mavenDependencyList, mavenDependency)
				continue
			}
			if strings.Contains(row, "The following files have been resolved:") {
				start = true
			}
		}
	}

	// 校验maven文件是否存在
	mavenRepository := fmt.Sprintf("%s%s%c.m2%crepository%c", os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"), os.PathSeparator, os.PathSeparator, os.PathSeparator)
	if len(mavenDependencyList) > 0 {
		for _, mavenDependency := range mavenDependencyList {
			jarPath := fmt.Sprintf("%s%s", mavenRepository, mavenDependency.filePath)
			_, err := os.Stat(jarPath)
			if err != nil {
				fmt.Println(mavenDependency.groupId, mavenDependency.artifactId, mavenDependency.version, mavenDependency.packaging, mavenDependency.dirPath, mavenDependency.filePath)
				fmt.Println("no find", jarPath)
			}
		}
	}
	fmt.Println("exit")
}

func exit(msg string, err error) {
	if err == nil {
		fmt.Println(msg)
	} else {
		fmt.Println(msg, err)
	}
	fmt.Println("exit")
	buf := make([]byte, 1)
	_, _ = os.Stdin.Read(buf)
	os.Exit(0)
}
