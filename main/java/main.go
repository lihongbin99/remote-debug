package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

type OsType int8

const (
	_ OsType = iota
	Linux
	Window
)

var (
	osType OsType

	java  = ""
	javac = ""
	mvn   = ""

	// 项目路径
	projectPath = ""
	// 模块路径
	modulePath = ""
	// 运行类
	runClass = ""

	// maven 仓库路径
	mavenRepositoryPath = ""

	// 发生异常后是否自动退出
	errAutoExit = 1
)

func init() {
	if os.PathSeparator == '/' {
		osType = Linux
	} else if os.PathSeparator == '\\' {
		osType = Window
	} else {
		exit(fmt.Sprintf("os type error: %c", os.PathSeparator), nil)
	}

	if osType == Linux {
		java = "/usr/bin/java"
		javac = "/usr/bin/javac"
		mvn = "/usr/bin/mvn"
		mavenRepositoryPath = fmt.Sprintf("%s/.m2/repository", os.Getenv("HOME"))
	} else if osType == Window {
		java = fmt.Sprintf("%s/bin/java.exe", os.Getenv("JAVA_HOME"))
		javac = fmt.Sprintf("%s/bin/javac.exe", os.Getenv("JAVA_HOME"))
		mvn = fmt.Sprintf("%s/bin/mvn", os.Getenv("MAVEN_HOME"))
		mavenRepositoryPath = fmt.Sprintf("%s%s/.m2/repository", os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"))
	} else {
		exit(fmt.Sprintf("os type error: %v", osType), nil)
	}

	flag.StringVar(&java, "java", java, "java.exe path")
	flag.StringVar(&javac, "javac", javac, "javac.exe path")
	flag.StringVar(&mvn, "mvn", mvn, "mvn path")

	flag.StringVar(&projectPath, "p", projectPath, "project path: C:\\Users\\Lee\\IdeaProjects\\remote-debug-test")
	flag.StringVar(&modulePath, "m", modulePath, "module path: service\\app")
	flag.StringVar(&runClass, "c", runClass, "run class path: io.lihongbin.remote.debug.test.RemoteDebugTestApplication")

	flag.StringVar(&mavenRepositoryPath, "r", mavenRepositoryPath, "repository path: C:\\Users\\Lee\\.m2\\repository")

	flag.IntVar(&errAutoExit, "e", errAutoExit, "error auto exit: 0")
	flag.Parse()
}

type MavenDependency struct {
	groupId    string
	artifactId string
	version    string
	packaging  string

	dirPath  string
	filePath []string
}

// 需要配置 MAVEN_HOME 环境变量, 否则需要在程序启动参数指定 -maven
// 需要配置 JAVA_HOME 环境变量, 否则需要在程序启动参数指定 -java
// 需要先对父工程 mvn compile install
func main() {
	// 校验参数
	verifyParam()

	// 解析 maven依赖
	mavenDependencyList := parseMavenDependency()

	// 解析所有jar包路径
	classpath := ""
	if len(mavenDependencyList) > 0 {
		for _, mavenDependency := range mavenDependencyList {
			for _, jarPath := range mavenDependency.filePath {
				classpath += fmt.Sprintf("%s/%s%c", mavenRepositoryPath, jarPath, os.PathListSeparator)
			}
		}
	}
	if len(classpath) > 0 {
		classpath = classpath[:len(classpath)-1]
	}

	classPath := fmt.Sprintf("%s/%s/target/classes", projectPath, modulePath)

	// 生成 运行命令
	runCommand := fmt.Sprintf("\"%s\" -Dfile.encoding=UTF-8 -classpath \"%s%c%s\" %s", java, classPath, os.PathListSeparator, classpath, runClass)

	fmt.Println(runCommand)
}

func verifyParam() {
	// 判断 java 和 javac 是否存在
	if java == "" {
		exit("place input JAVA_HOME path param: -java <java.exe path>", nil)
	}
	if _, err := os.Stat(java); err != nil {
		exit("no find java.exe", err)
	}
	if javac == "" {
		exit("place input JAVA_HOME path param: -java <javac.exe path>", nil)
	}
	if _, err := os.Stat(javac); err != nil {
		exit("no find javac.exe", err)
	}

	// 判断 mvn 是否存在
	if mvn == "" {
		exit("place input MAVEN_HOME path param: -maven <mvn path>", nil)
	}
	if _, err := os.Stat(mvn); err != nil {
		exit("no find mvn", err)
	}

	// 判断 project 和 module 和 启动类 是否存在
	if projectPath == "" {
		exit("place input project path param: -p <project path>", nil)
	}
	if _, err := os.Stat(projectPath); err != nil {
		exit("project path error", err)
	}
	if modulePath == "" {
		exit("place input module path param: -m <module path>", nil)
	}
	if _, err := os.Stat(fmt.Sprintf("%s/%s", projectPath, modulePath)); err != nil {
		exit("module path error", err)
	}
	if runClass == "" {
		exit("place input run class param: -c <run class>", nil)
	}
	if _, err := os.Stat(fmt.Sprintf("%s/%s/src/main/java/%s.java", projectPath, modulePath, strings.ReplaceAll(runClass, ".", "/"))); err != nil {
		exit("run class error", err)
	}

	// 判断 maven 仓库是否存在
	if mavenRepositoryPath == "" {
		exit("place input repository path param: -r <repository path>", nil)
	}
	if _, err := os.Stat(mavenRepositoryPath); err != nil {
		exit("repository path error", err)
	}
}

// 获取所有依赖
func parseMavenDependency() []MavenDependency {
	mavenDependencyList := make([]MavenDependency, 0)
	// 用来判断是否jar包是否已经添加了, 避免重复添加
	jarExistCache := make(map[string]int8, 0)

	// 编译项目的源文件 编译项目依赖的 jar包
	startTime := time.Now()
	cmdResult := runCmd(mvn, projectPath, []string{"mvn", "compile"})
	if !strings.Contains(cmdResult, "BUILD SUCCESS") {
		exit("mvn compile error", fmt.Errorf(cmdResult))
	}
	endTime := time.Now()
	fmt.Println("mvn exec compile time:", endTime.UnixMilli()-startTime.UnixMilli(), "ms")
	startTime = time.Now()
	cmdResult = runCmd(mvn, projectPath, []string{"mvn", "install"})
	if !strings.Contains(cmdResult, "BUILD SUCCESS") {
		exit("mvn install error", fmt.Errorf(cmdResult))
	}
	endTime = time.Now()
	fmt.Println("mvn exec install time:", endTime.UnixMilli()-startTime.UnixMilli(), "ms")

	// 获取 maven 返回的依赖列表
	cmdResult = runCmd(mvn, fmt.Sprintf("%s/%s", projectPath, modulePath), []string{"mvn", "dependency:list"})
	if !strings.Contains(cmdResult, "BUILD SUCCESS") {
		exit("mvn dependency:list error", fmt.Errorf(cmdResult))
	}

	// 分割每一行依赖项
	rows := strings.Split(cmdResult, "\n")
	start := false
	for _, row := range rows {
		// 通过这个条件判断获取依赖是否成功
		if !start && strings.Contains(row, "The following files have been resolved:") {
			start = true
			continue
		}
		if !start {
			continue
		}

		// maven 返回结果会以 [INFO] 开头, 所以需要先排除日志级别
		row = strings.TrimSpace(row[strings.Index(row, "]")+1:])
		// 如果有一行是空的则代表已经解析完所有依赖项了
		if len(row) == 0 {
			break
		}

		// 没有啥好办法, 只能见招拆招了
		dependencyInfoList := strings.Split(row, ":")
		var mavenDependency MavenDependency
		if len(dependencyInfoList) == 5 {
			mavenDependency = MavenDependency{groupId: dependencyInfoList[0], artifactId: dependencyInfoList[1], version: dependencyInfoList[3], packaging: dependencyInfoList[2]}
		} else if len(dependencyInfoList) == 6 {
			mavenDependency = MavenDependency{groupId: dependencyInfoList[0], artifactId: dependencyInfoList[1], version: dependencyInfoList[4], packaging: dependencyInfoList[2]}
		} else {
			exit("mvn dependency list dependencyInfoList len error: "+row, nil)
		}

		// 计算 jar包 路径, 需要把 groupId 转成文件路径
		mavenDependency.dirPath = fmt.Sprintf("%s/%s/%s", strings.ReplaceAll(mavenDependency.groupId, ".", "/"), mavenDependency.artifactId, mavenDependency.version)

		// 获取此依赖项下的所有 jar包
		mavenDependency.filePath = make([]string, 0)
		dependencyDir, err := os.ReadDir(fmt.Sprintf("%s/%s", mavenRepositoryPath, mavenDependency.dirPath))
		if err != nil {
			exit("no find dependency dir", err)
		}
		for _, jarFileInfo := range dependencyDir {
			if !jarFileInfo.IsDir() && strings.HasSuffix(jarFileInfo.Name(), ".jar") && !strings.HasSuffix(jarFileInfo.Name(), "-javadoc.jar") && !strings.HasSuffix(jarFileInfo.Name(), "-sources.jar") {
				jarPath := fmt.Sprintf("%s/%s", mavenDependency.dirPath, jarFileInfo.Name())
				// 去重
				if _, exist := jarExistCache[jarPath]; !exist {
					jarExistCache[jarPath] = 1
					mavenDependency.filePath = append(mavenDependency.filePath, jarPath)
				}
			}
		}

		// 添加依赖项
		mavenDependencyList = append(mavenDependencyList, mavenDependency)
	}

	// 判断 maven 获取依赖是否成功
	if !start {
		exit("mvn dependency list error", fmt.Errorf("%s", cmdResult))
	}

	return mavenDependencyList
}

func exit(msg string, err error) {
	if err == nil {
		fmt.Println(msg)
	} else {
		fmt.Println(msg, err)
	}
	if errAutoExit == 0 {
		fmt.Println("input char exit")
		buf := make([]byte, 1)
		_, _ = os.Stdin.Read(buf)
	}
	os.Exit(1)
}

func runCmd(path, dir string, args []string) string {
	cmd := &exec.Cmd{
		Path: path,
		Args: args,
		Dir:  dir,
	}
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		exit("cmd pipe error", err)
	}
	if err := cmd.Start(); err != nil {
		exit("cmd start error", err)
	}
	result := make([]byte, 0)
	buf := make([]byte, 64*1024)
	for {
		readLen, err := pipe.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			exit("cmd read error", err)
		}
		result = append(result, buf[:readLen]...)
	}
	return string(result)
}
