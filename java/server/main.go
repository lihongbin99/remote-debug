package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"remote-debug/common/io"
	"remote-debug/common/utils"
	"remote-debug/java/common"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type MavenDependency struct {
	groupId    string
	artifactId string
	version    string
	packaging  string

	dirPath  string
	filePath []string
}

var (
	listenPort = 50005

	projectInfo = common.ProjectInfo{}

	processMap = make(map[string]int)

	projectPath string
	projectName string
)

func init() {
	flag.IntVar(&listenPort, "port", listenPort, "listen port: 50005")

	flag.Parse()

	// 校验环境参数是否正确
	common.VerifyParam()
}

func main() {
	listenAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		common.Exit("resolve tcp addr error", err)
	}
	listener, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		common.Exit("listen tcp error", err)
	}

	fmt.Println("start server success")
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			common.Exit("accept tcp error", err)
		}
		go startProcess(conn)
	}
}

func startProcess(conn *net.TCPConn) {
	defer func() { _ = conn.Close() }()
	fmt.Println("new request")

	// 接收参数
	readParam(conn)

	// 重新设置项目路径
	projectName = getProjectName(projectInfo.ProjectPath)
	projectPath = fmt.Sprintf("%s/remote-debug/%s", common.HomePath, projectName)
	projectInfo.ProjectPath = projectPath

	// 如果有旧项目则需要先暂停
	stopOldProject(projectName)

	// 删除旧文件
	_ = os.RemoveAll(projectInfo.ProjectPath)

	// 下载文件
	zipData, err := downloadZip(conn)
	if err != nil {
		_ = io.SendMessage(conn, common.Result{Code: 500, Msg: err.Error()})
		return
	}

	// 解压 zip
	if err = UnZip(zipData); err != nil {
		_ = io.SendMessage(conn, common.Result{Code: 500, Msg: err.Error()})
		return
	}

	// 解析 maven依赖
	mavenDependencyList, err := parseMavenDependency()
	if err != nil {
		common.PrintError("parse maven dependency error", err)
		_ = io.SendMessage(conn, common.Result{Code: 500, Msg: err.Error()})
		return
	}

	// 解析所有jar包路径
	classpath := parseJarPath(mavenDependencyList)
	classPath := fmt.Sprintf("%s/%s/target/classes", projectInfo.ProjectPath, projectInfo.ModulePath)
	classpath = fmt.Sprintf("%s%c%s", classPath, os.PathListSeparator, classpath)

	// 创建日志文件
	logFile, err := createLogFile()
	if err != nil {
		_ = io.SendMessage(conn, common.Result{Code: 500, Msg: err.Error()})
		return
	}

	// 运行项目
	cmd, err := runProject(classpath, logFile)
	if err != nil {
		_ = io.SendMessage(conn, common.Result{Code: 500, Msg: err.Error()})
		return
	}
	pid := cmd.Process.Pid

	// 返回结果
	fmt.Println("start process success", pid)
	_ = io.SendMessage(conn, common.Result{Code: 200, Msg: strconv.Itoa(pid)})

	// 等待停止项目
	processMap[projectName] = pid
	if _, err = cmd.Process.Wait(); err != nil {
		fmt.Println("wait error", pid, err)
	}
	fmt.Println(pid, "exit success")
}

func readParam(conn *net.TCPConn) {
	if err := io.ReadMessage(conn, &projectInfo); err != nil {
		common.PrintError("read project info error", err)
		return
	}
	fmt.Println("read project info success", projectInfo.ProjectPath, projectInfo.ModulePath, projectInfo.RunClass)
	fmt.Printf("zip time: %dms\n", projectInfo.ZipTime)
}

func stopOldProject(projectName string) {
	if pid, exist := processMap[projectName]; exist {
		fmt.Println("stop old project", pid)
		syscall.Kill(pid, 15)
		for {
			if err := syscall.Kill(pid, 0); err != nil {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		delete(processMap, projectName)
	}
}

func downloadZip(conn *net.TCPConn) ([]byte, error) {
	startTime := time.Now()
	zipData, err := io.ReadData(conn)
	if err != nil {
		common.PrintError("download zip data error", err)
		// 返回错误
		return nil, err
	}
	endTime := time.Now()
	fmt.Printf("dowwnload zip time: %dms\n", endTime.UnixMilli()-startTime.UnixMilli())
	return zipData, nil
}

func UnZip(zipData []byte) error {
	startTime := time.Now()
	fmt.Println("new project path", projectPath)
	if err := utils.UnZip(zipData, projectPath); err != nil {
		common.PrintError("un zip error", err)
		return err
	}
	endTime := time.Now()
	fmt.Printf("un zip time: %dms\n", endTime.UnixMilli()-startTime.UnixMilli())
	return nil
}

func parseJarPath(mavenDependencyList []MavenDependency) string {
	classpath := ""
	if len(mavenDependencyList) > 0 {
		for _, mavenDependency := range mavenDependencyList {
			for _, jarPath := range mavenDependency.filePath {
				classpath += fmt.Sprintf("%s/%s%c", common.MavenRepositoryPath, jarPath, os.PathListSeparator)
			}
		}
	}
	if len(classpath) > 0 {
		classpath = classpath[:len(classpath)-1]
	}
	return classpath
}

func createLogFile() (*os.File, error) {
	_ = os.MkdirAll(fmt.Sprintf("%s/remote-debug/logs", common.HomePath), 0777)
	logPath := fmt.Sprintf("%s/remote-debug/logs/%s.log", common.HomePath, projectName)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		common.PrintError("open log path error error", err)
		return nil, err
	}
	return logFile, nil
}

func runProject(classpath string, logFile *os.File) (*exec.Cmd, error) {
	cmd := &exec.Cmd{
		Path:   common.Java,
		Args:   []string{common.Java, "-Dfile.encoding=UTF-8", projectInfo.Params, "-classpath", classpath, projectInfo.RunClass}, // TODO 不知道怎么改成后台启动
		Dir:    projectInfo.ProjectPath,
		Stdout: logFile,
	}
	fmt.Println(cmd.Args)
	if err := cmd.Start(); err != nil {
		common.PrintError("start process error", err)
		// 返回错误
		return nil, err
	}
	return cmd, nil
}

// 获取所有依赖
func parseMavenDependency() ([]MavenDependency, error) {
	mavenDependencyList := make([]MavenDependency, 0)
	// 用来判断是否jar包是否已经添加了, 避免重复添加
	jarExistCache := make(map[string]int8, 0)

	// 编译项目的源文件 编译项目依赖的 jar包
	startTime := time.Now()
	cmdResult, err := common.RunCommand(common.Mvn, projectInfo.ProjectPath, []string{"mvn", "compile"})
	if err != nil {
		return nil, err
	}
	if !strings.Contains(cmdResult, "BUILD SUCCESS") {
		return nil, fmt.Errorf("mvn compile error: %s", cmdResult)
	}
	endTime := time.Now()
	fmt.Printf("mvn exec compile time: %dms\n", endTime.UnixMilli()-startTime.UnixMilli())
	startTime = time.Now()
	cmdResult, err = common.RunCommand(common.Mvn, projectInfo.ProjectPath, []string{"mvn", "install"})
	if err != nil {
		return nil, err
	}
	if !strings.Contains(cmdResult, "BUILD SUCCESS") {
		return nil, fmt.Errorf("mvn install error: %s", cmdResult)
	}
	endTime = time.Now()
	fmt.Printf("mvn exec install time: %dms\n", endTime.UnixMilli()-startTime.UnixMilli())

	// 获取 maven 返回的依赖列表
	startTime = time.Now()
	cmdResult, err = common.RunCommand(common.Mvn,
		fmt.Sprintf("%s/%s", projectInfo.ProjectPath, projectInfo.ModulePath),
		[]string{"mvn", "dependency:list"})
	if err != nil {
		return nil, err
	}
	if !strings.Contains(cmdResult, "BUILD SUCCESS") {
		return nil, fmt.Errorf("mvn dependency:list error: %s", cmdResult)
	}
	endTime = time.Now()
	fmt.Printf("mvn exec dependency:list time: %dms\n", endTime.UnixMilli()-startTime.UnixMilli())

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
			return nil, fmt.Errorf("mvn dependency list dependencyInfoList len error: %s", row)
		}

		// 计算 jar包 路径, 需要把 groupId 转成文件路径
		mavenDependency.dirPath = fmt.Sprintf("%s/%s/%s",
			strings.ReplaceAll(mavenDependency.groupId, ".", "/"),
			mavenDependency.artifactId, mavenDependency.version)

		// 获取此依赖项下的所有 jar包
		mavenDependency.filePath = make([]string, 0)
		dependencyDir, err := os.ReadDir(fmt.Sprintf("%s/%s", common.MavenRepositoryPath, mavenDependency.dirPath))
		if err != nil {
			return nil, fmt.Errorf("no find dependency dir: %s", err.Error())
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
		return nil, fmt.Errorf("mvn dependency list error: %s", cmdResult)
	}

	return mavenDependencyList, nil
}

func getProjectName(projectPath string) string {
	i1 := strings.LastIndex(projectPath, "/")
	i2 := strings.LastIndex(projectPath, "\\")

	if i1 > i2 {
		return projectPath[i1+1:]
	} else {
		return projectPath[i2+1:]
	}
}
