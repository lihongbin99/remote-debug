package common

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type osType int8

const (
	_ osType = iota
	Linux
	Window
)

var (
	OsType osType

	Java  string
	Javac string
	Mvn   string

	HomePath            string
	MavenRepositoryPath string
)

func init() {
	if os.PathSeparator == '/' {
		OsType = Linux
	} else if os.PathSeparator == '\\' {
		OsType = Window
	} else {
		Exit(fmt.Sprintf("os type error: %c", os.PathSeparator), nil)
	}

	if OsType == Linux {
		Java = "/usr/bin/java"
		Javac = "/usr/bin/javac"
		Mvn = "/usr/bin/mvn"
		HomePath = os.Getenv("HOME")
	} else if OsType == Window {
		Java = fmt.Sprintf("%s/bin/java.exe", os.Getenv("JAVA_HOME"))
		Javac = fmt.Sprintf("%s/bin/javac.exe", os.Getenv("JAVA_HOME"))
		Mvn = fmt.Sprintf("%s/bin/mvn", os.Getenv("MAVEN_HOME"))
		HomePath = fmt.Sprintf("%s%s", os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"))
	} else {
		Exit(fmt.Sprintf("os type error: %v", OsType), nil)
	}

	MavenRepositoryPath = fmt.Sprintf("%s/.m2/repository", HomePath)

	flag.StringVar(&Java, "java", Java, "java.exe path")
	flag.StringVar(&Javac, "javac", Javac, "javac.exe path")
	flag.StringVar(&Mvn, "mvn", Mvn, "mvn path")

	flag.StringVar(&MavenRepositoryPath, "r", MavenRepositoryPath,
		"repository path: C:\\Users\\Lee\\.m2\\repository")
}

func VerifyParam() {
	// 判断 java 和 javac 是否存在
	if Java == "" {
		Exit("place input JAVA_HOME path param: -java <java.exe path>", nil)
	}
	if _, err := os.Stat(Java); err != nil {
		Exit("no find java.exe", err)
	}
	if Javac == "" {
		Exit("place input JAVA_HOME path param: -java <javac.exe path>", nil)
	}
	if _, err := os.Stat(Javac); err != nil {
		Exit("no find javac.exe", err)
	}

	// 判断 mvn 是否存在
	if Mvn == "" {
		Exit("place input MAVEN_HOME path param: -maven <mvn path>", nil)
	}
	if _, err := os.Stat(Mvn); err != nil {
		Exit("no find mvn", err)
	}

	// 判断 maven 仓库是否存在
	if MavenRepositoryPath == "" {
		Exit("place input repository path param: -r <repository path>", nil)
	}
	if _, err := os.Stat(MavenRepositoryPath); err != nil {
		Exit("repository path error", err)
	}
}

func RunCommand(path, dir string, args []string) (string, error) {
	cmd := &exec.Cmd{
		Path: path,
		Args: args,
		Dir:  dir,
	}
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("pipe error: %s", err.Error())
	}
	defer func() { _ = pipe.Close() }()
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start error: %s", err.Error())
	}
	result := make([]byte, 0)
	buf := make([]byte, 64*1024)
	for {
		readLen, err := pipe.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("read error: %s", err.Error())
		}
		result = append(result, buf[:readLen]...)
	}
	return string(result), nil
}

func Exit(msg string, err error) {
	PrintError(msg, err)
	os.Exit(1)
}

func PrintError(msg string, err error) int {
	if err == nil {
		fmt.Println(msg)
	} else {
		fmt.Println(msg, err)
	}
	return 1
}
