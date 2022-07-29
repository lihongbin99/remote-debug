package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"remote-debug/common/io"
	"remote-debug/common/utils"
	"remote-debug/java/common"
	"strings"
	"time"
)

var (
	serverAddrS = "0.0.0.0:50005"
	ignore      = ".git,.idea,target"

	projectInfo = common.ProjectInfo{}
)

func init() {
	flag.StringVar(&serverAddrS, "s", serverAddrS, "server address: 0.0.0.0:50005")
	flag.StringVar(&ignore, "i", ignore, "ignore: .git,.idea,target")

	flag.StringVar(&projectInfo.ProjectPath, "p", "",
		"project path: C:\\Users\\Lee\\IdeaProjects\\remote-debug-test")
	flag.StringVar(&projectInfo.ModulePath, "m", "",
		"module path: service\\app")
	flag.StringVar(&projectInfo.RunClass, "c", "",
		"run class path: io.lihongbin.remote.debug.test.RemoteDebugTestApplication")
	flag.StringVar(&projectInfo.Params, "param", "",
		"params: -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005")

	flag.Parse()

	// 校验环境参数是否正确
	common.VerifyParam()
}

func main() {
	// 校验参数
	verifyParam()
	fmt.Println("verify param success")

	// 打包项目
	zipData := projectToZip()

	// 连接服务器
	conn := connectServer()

	// 上传参数
	sendParam(conn)

	// 上传文件
	sendData(conn, zipData)

	// 获取返回结果
	result := common.Result{}
	if err := io.ReadMessage(conn, &result); err != nil {
		common.Exit("read result error", err)
	}
	fmt.Println("read result success: ", result.Code, result.Msg)
}

func verifyParam() {
	// 判断 project 和 module 和 启动类 是否存在
	if projectInfo.ProjectPath == "" {
		common.Exit("place input project path param: -p <project path>", nil)
	}
	if _, err := os.Stat(projectInfo.ProjectPath); err != nil {
		common.Exit("project path error", err)
	}
	if projectInfo.ModulePath == "" {
		common.Exit("place input module path param: -m <module path>", nil)
	}
	if _, err := os.Stat(fmt.Sprintf("%s/%s", projectInfo.ProjectPath, projectInfo.ModulePath)); err != nil {
		common.Exit("module path error", err)
	}
	if projectInfo.RunClass == "" {
		common.Exit("place input run class param: -c <run class>", nil)
	}
	if _, err := os.Stat(fmt.Sprintf("%s/%s/src/main/java/%s.java",
		projectInfo.ProjectPath, projectInfo.ModulePath,
		strings.ReplaceAll(projectInfo.RunClass, ".", "/"))); err != nil {
		common.Exit("run class error", err)
	}
}

func projectToZip() []byte {
	startTime := time.Now()
	zipData, err := utils.Zip(projectInfo.ProjectPath, ignore)
	if err != nil {
		common.Exit("package zip error", err)
	}
	endTime := time.Now()
	projectInfo.ZipTime = int(endTime.UnixMilli() - startTime.UnixMilli())
	return zipData
}

func connectServer() *net.TCPConn {
	serverAddr, err := net.ResolveTCPAddr("tcp", serverAddrS)
	if err != nil {
		common.Exit("resolve tcp addr error", err)
	}
	conn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		common.Exit("dial tcp error", err)
	}
	defer func() { _ = conn.Close() }()
	fmt.Println("connect server success")
	return conn
}

func sendParam(conn *net.TCPConn) {
	if err := io.SendMessage(conn, &projectInfo); err != nil {
		common.Exit("send project info error", err)
	}
	fmt.Println("send project info success")
}

func sendData(conn *net.TCPConn, zipData []byte) {
	if err := io.SendData(conn, zipData); err != nil {
		common.Exit("send zip data error", err)
	}
}
