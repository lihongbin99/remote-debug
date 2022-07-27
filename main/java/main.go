package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
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

/*

javac ^
-classpath C:\Users\Lee\.m2\repository\io\lihongbin\model\1.0.0\model-1.0.0.jar;C:\Users\Lee\.m2\repository\org\springframework\boot\spring-boot-starter\2.7.2\spring-boot-starter-2.7.2.jar;C:\Users\Lee\.m2\repository\org\springframework\boot\spring-boot\2.7.2\spring-boot-2.7.2.jar;C:\Users\Lee\.m2\repository\org\springframework\spring-context\5.3.22\spring-context-5.3.22.jar;C:\Users\Lee\.m2\repository\org\springframework\boot\spring-boot-autoconfigure\2.7.2\spring-boot-autoconfigure-2.7.2.jar;C:\Users\Lee\.m2\repository\org\springframework\boot\spring-boot-starter-logging\2.7.2\spring-boot-starter-logging-2.7.2.jar;C:\Users\Lee\.m2\repository\ch\qos\logback\logback-classic\1.2.11\logback-classic-1.2.11.jar;C:\Users\Lee\.m2\repository\ch\qos\logback\logback-core\1.2.11\logback-core-1.2.11.jar;C:\Users\Lee\.m2\repository\org\apache\logging\log4j\log4j-to-slf4j\2.17.2\log4j-to-slf4j-2.17.2.jar;C:\Users\Lee\.m2\repository\org\apache\logging\log4j\log4j-api\2.17.2\log4j-api-2.17.2.jar;C:\Users\Lee\.m2\repository\org\slf4j\jul-to-slf4j\1.7.36\jul-to-slf4j-1.7.36.jar;C:\Users\Lee\.m2\repository\jakarta\annotation\jakarta.annotation-api\1.3.5\jakarta.annotation-api-1.3.5.jar;C:\Users\Lee\.m2\repository\org\springframework\spring-core\5.3.22\spring-core-5.3.22.jar;C:\Users\Lee\.m2\repository\org\springframework\spring-jcl\5.3.22\spring-jcl-5.3.22.jar;C:\Users\Lee\.m2\repository\org\yaml\snakeyaml\1.30\snakeyaml-1.30.jar;C:\Users\Lee\.m2\repository\org\springframework\boot\spring-boot-starter-test\2.7.2\spring-boot-starter-test-2.7.2.jar;C:\Users\Lee\.m2\repository\org\springframework\boot\spring-boot-test\2.7.2\spring-boot-test-2.7.2.jar;C:\Users\Lee\.m2\repository\org\springframework\boot\spring-boot-test-autoconfigure\2.7.2\spring-boot-test-autoconfigure-2.7.2.jar;C:\Users\Lee\.m2\repository\com\jayway\jsonpath\json-path\2.7.0\json-path-2.7.0.jar;C:\Users\Lee\.m2\repository\net\minidev\json-smart\2.4.8\json-smart-2.4.8.jar;C:\Users\Lee\.m2\repository\net\minidev\accessors-smart\2.4.8\accessors-smart-2.4.8.jar;C:\Users\Lee\.m2\repository\org\ow2\asm\asm\9.1\asm-9.1.jar;C:\Users\Lee\.m2\repository\org\slf4j\slf4j-api\1.7.36\slf4j-api-1.7.36.jar;C:\Users\Lee\.m2\repository\jakarta\xml\bind\jakarta.xml.bind-api\2.3.3\jakarta.xml.bind-api-2.3.3.jar;C:\Users\Lee\.m2\repository\jakarta\activation\jakarta.activation-api\1.2.2\jakarta.activation-api-1.2.2.jar;C:\Users\Lee\.m2\repository\org\assertj\assertj-core\3.22.0\assertj-core-3.22.0.jar;C:\Users\Lee\.m2\repository\org\hamcrest\hamcrest\2.2\hamcrest-2.2.jar;C:\Users\Lee\.m2\repository\org\junit\jupiter\junit-jupiter\5.8.2\junit-jupiter-5.8.2.jar;C:\Users\Lee\.m2\repository\org\junit\jupiter\junit-jupiter-api\5.8.2\junit-jupiter-api-5.8.2.jar;C:\Users\Lee\.m2\repository\org\opentest4j\opentest4j\1.2.0\opentest4j-1.2.0.jar;C:\Users\Lee\.m2\repository\org\junit\platform\junit-platform-commons\1.8.2\junit-platform-commons-1.8.2.jar;C:\Users\Lee\.m2\repository\org\apiguardian\apiguardian-api\1.1.2\apiguardian-api-1.1.2.jar;C:\Users\Lee\.m2\repository\org\junit\jupiter\junit-jupiter-params\5.8.2\junit-jupiter-params-5.8.2.jar;C:\Users\Lee\.m2\repository\org\junit\jupiter\junit-jupiter-engine\5.8.2\junit-jupiter-engine-5.8.2.jar;C:\Users\Lee\.m2\repository\org\junit\platform\junit-platform-engine\1.8.2\junit-platform-engine-1.8.2.jar;C:\Users\Lee\.m2\repository\org\mockito\mockito-core\4.5.1\mockito-core-4.5.1.jar;C:\Users\Lee\.m2\repository\net\bytebuddy\byte-buddy\1.12.12\byte-buddy-1.12.12.jar;C:\Users\Lee\.m2\repository\net\bytebuddy\byte-buddy-agent\1.12.12\byte-buddy-agent-1.12.12.jar;C:\Users\Lee\.m2\repository\org\objenesis\objenesis\3.2\objenesis-3.2.jar;C:\Users\Lee\.m2\repository\org\mockito\mockito-junit-jupiter\4.5.1\mockito-junit-jupiter-4.5.1.jar;C:\Users\Lee\.m2\repository\org\skyscreamer\jsonassert\1.5.1\jsonassert-1.5.1.jar;C:\Users\Lee\.m2\repository\com\vaadin\external\google\android-json\0.0.20131108.vaadin1\android-json-0.0.20131108.vaadin1.jar;C:\Users\Lee\.m2\repository\org\springframework\spring-test\5.3.22\spring-test-5.3.22.jar;C:\Users\Lee\.m2\repository\org\xmlunit\xmlunit-core\2.9.0\xmlunit-core-2.9.0.jar;C:\Users\Lee\.m2\repository\org\springframework\boot\spring-boot-starter-web\2.7.2\spring-boot-starter-web-2.7.2.jar;C:\Users\Lee\.m2\repository\org\springframework\boot\spring-boot-starter-json\2.7.2\spring-boot-starter-json-2.7.2.jar;C:\Users\Lee\.m2\repository\com\fasterxml\jackson\core\jackson-databind\2.13.3\jackson-databind-2.13.3.jar;C:\Users\Lee\.m2\repository\com\fasterxml\jackson\core\jackson-annotations\2.13.3\jackson-annotations-2.13.3.jar;C:\Users\Lee\.m2\repository\com\fasterxml\jackson\core\jackson-core\2.13.3\jackson-core-2.13.3.jar;C:\Users\Lee\.m2\repository\com\fasterxml\jackson\datatype\jackson-datatype-jdk8\2.13.3\jackson-datatype-jdk8-2.13.3.jar;C:\Users\Lee\.m2\repository\com\fasterxml\jackson\datatype\jackson-datatype-jsr310\2.13.3\jackson-datatype-jsr310-2.13.3.jar;C:\Users\Lee\.m2\repository\com\fasterxml\jackson\module\jackson-module-parameter-names\2.13.3\jackson-module-parameter-names-2.13.3.jar;C:\Users\Lee\.m2\repository\org\springframework\boot\spring-boot-starter-tomcat\2.7.2\spring-boot-starter-tomcat-2.7.2.jar;C:\Users\Lee\.m2\repository\org\apache\tomcat\embed\tomcat-embed-core\9.0.65\tomcat-embed-core-9.0.65.jar;C:\Users\Lee\.m2\repository\org\apache\tomcat\embed\tomcat-embed-el\9.0.65\tomcat-embed-el-9.0.65.jar;C:\Users\Lee\.m2\repository\org\apache\tomcat\embed\tomcat-embed-websocket\9.0.65\tomcat-embed-websocket-9.0.65.jar;C:\Users\Lee\.m2\repository\org\springframework\spring-web\5.3.22\spring-web-5.3.22.jar;C:\Users\Lee\.m2\repository\org\springframework\spring-beans\5.3.22\spring-beans-5.3.22.jar;C:\Users\Lee\.m2\repository\org\springframework\spring-webmvc\5.3.22\spring-webmvc-5.3.22.jar;C:\Users\Lee\.m2\repository\org\springframework\spring-aop\5.3.22\spring-aop-5.3.22.jar;C:\Users\Lee\.m2\repository\org\springframework\spring-expression\5.3.22\spring-expression-5.3.22.jar;C:\Users\Lee\.m2\repository\org\springframework\boot\spring-boot-devtools\2.7.2\spring-boot-devtools-2.7.2.jar;C:\Users\Lee\.m2\repository\org\springframework\boot\spring-boot-configuration-processor\2.7.2\spring-boot-configuration-processor-2.7.2.jar;C:\Users\Lee\.m2\repository\org\projectlombok\lombok\1.18.24\lombok-1.18.24.jar ^
-d C:\Users\Lee\IdeaProjects\remote-debug-test\target\remote/classes ^
service\app\src\main\java\io\lihongbin\remote\debug\test\controller\Controller.java ^
service\app\src\main\java\io\lihongbin\remote\debug\test\RemoteDebugTestApplication.java

-sourcepath io ^

java \
-Dfile.encoding=UTF-8 \
-classpath /home/ubuntu/remote-debug-test/service/app/src/main/java/test:/home/ubuntu/.m2/repository/io/lihongbin/model/1.0.0/model-1.0.0.jar:/home/ubuntu/.m2/repository/org/springframework/boot/spring-boot-starter/2.7.2/spring-boot-starter-2.7.2.jar:/home/ubuntu/.m2/repository/org/springframework/boot/spring-boot/2.7.2/spring-boot-2.7.2.jar:/home/ubuntu/.m2/repository/org/springframework/spring-context/5.3.22/spring-context-5.3.22.jar:/home/ubuntu/.m2/repository/org/springframework/boot/spring-boot-autoconfigure/2.7.2/spring-boot-autoconfigure-2.7.2.jar:/home/ubuntu/.m2/repository/org/springframework/boot/spring-boot-starter-logging/2.7.2/spring-boot-starter-logging-2.7.2.jar:/home/ubuntu/.m2/repository/ch/qos/logback/logback-classic/1.2.11/logback-classic-1.2.11.jar:/home/ubuntu/.m2/repository/ch/qos/logback/logback-core/1.2.11/logback-core-1.2.11.jar:/home/ubuntu/.m2/repository/org/apache/logging/log4j/log4j-to-slf4j/2.17.2/log4j-to-slf4j-2.17.2.jar:/home/ubuntu/.m2/repository/org/apache/logging/log4j/log4j-api/2.17.2/log4j-api-2.17.2.jar:/home/ubuntu/.m2/repository/org/slf4j/jul-to-slf4j/1.7.36/jul-to-slf4j-1.7.36.jar:/home/ubuntu/.m2/repository/jakarta/annotation/jakarta.annotation-api/1.3.5/jakarta.annotation-api-1.3.5.jar:/home/ubuntu/.m2/repository/org/springframework/spring-core/5.3.22/spring-core-5.3.22.jar:/home/ubuntu/.m2/repository/org/springframework/spring-jcl/5.3.22/spring-jcl-5.3.22.jar:/home/ubuntu/.m2/repository/org/yaml/snakeyaml/1.30/snakeyaml-1.30.jar:/home/ubuntu/.m2/repository/org/springframework/boot/spring-boot-starter-test/2.7.2/spring-boot-starter-test-2.7.2.jar:/home/ubuntu/.m2/repository/org/springframework/boot/spring-boot-test/2.7.2/spring-boot-test-2.7.2.jar:/home/ubuntu/.m2/repository/org/springframework/boot/spring-boot-test-autoconfigure/2.7.2/spring-boot-test-autoconfigure-2.7.2.jar:/home/ubuntu/.m2/repository/com/jayway/jsonpath/json-path/2.7.0/json-path-2.7.0.jar:/home/ubuntu/.m2/repository/net/minidev/json-smart/2.4.8/json-smart-2.4.8.jar:/home/ubuntu/.m2/repository/net/minidev/accessors-smart/2.4.8/accessors-smart-2.4.8.jar:/home/ubuntu/.m2/repository/org/ow2/asm/asm/9.1/asm-9.1.jar:/home/ubuntu/.m2/repository/org/slf4j/slf4j-api/1.7.36/slf4j-api-1.7.36.jar:/home/ubuntu/.m2/repository/jakarta/xml/bind/jakarta.xml.bind-api/2.3.3/jakarta.xml.bind-api-2.3.3.jar:/home/ubuntu/.m2/repository/jakarta/activation/jakarta.activation-api/1.2.2/jakarta.activation-api-1.2.2.jar:/home/ubuntu/.m2/repository/org/assertj/assertj-core/3.22.0/assertj-core-3.22.0.jar:/home/ubuntu/.m2/repository/org/hamcrest/hamcrest/2.2/hamcrest-2.2.jar:/home/ubuntu/.m2/repository/org/junit/jupiter/junit-jupiter/5.8.2/junit-jupiter-5.8.2.jar:/home/ubuntu/.m2/repository/org/junit/jupiter/junit-jupiter-api/5.8.2/junit-jupiter-api-5.8.2.jar:/home/ubuntu/.m2/repository/org/opentest4j/opentest4j/1.2.0/opentest4j-1.2.0.jar:/home/ubuntu/.m2/repository/org/junit/platform/junit-platform-commons/1.8.2/junit-platform-commons-1.8.2.jar:/home/ubuntu/.m2/repository/org/apiguardian/apiguardian-api/1.1.2/apiguardian-api-1.1.2.jar:/home/ubuntu/.m2/repository/org/junit/jupiter/junit-jupiter-params/5.8.2/junit-jupiter-params-5.8.2.jar:/home/ubuntu/.m2/repository/org/junit/jupiter/junit-jupiter-engine/5.8.2/junit-jupiter-engine-5.8.2.jar:/home/ubuntu/.m2/repository/org/junit/platform/junit-platform-engine/1.8.2/junit-platform-engine-1.8.2.jar:/home/ubuntu/.m2/repository/org/mockito/mockito-core/4.5.1/mockito-core-4.5.1.jar:/home/ubuntu/.m2/repository/net/bytebuddy/byte-buddy/1.12.12/byte-buddy-1.12.12.jar:/home/ubuntu/.m2/repository/net/bytebuddy/byte-buddy-agent/1.12.12/byte-buddy-agent-1.12.12.jar:/home/ubuntu/.m2/repository/org/objenesis/objenesis/3.2/objenesis-3.2.jar:/home/ubuntu/.m2/repository/org/mockito/mockito-junit-jupiter/4.5.1/mockito-junit-jupiter-4.5.1.jar:/home/ubuntu/.m2/repository/org/skyscreamer/jsonassert/1.5.1/jsonassert-1.5.1.jar:/home/ubuntu/.m2/repository/com/vaadin/external/google/android-json/0.0.20131108.vaadin1/android-json-0.0.20131108.vaadin1.jar:/home/ubuntu/.m2/repository/org/springframework/spring-test/5.3.22/spring-test-5.3.22.jar:/home/ubuntu/.m2/repository/org/xmlunit/xmlunit-core/2.9.0/xmlunit-core-2.9.0.jar:/home/ubuntu/.m2/repository/org/springframework/boot/spring-boot-starter-web/2.7.2/spring-boot-starter-web-2.7.2.jar:/home/ubuntu/.m2/repository/org/springframework/boot/spring-boot-starter-json/2.7.2/spring-boot-starter-json-2.7.2.jar:/home/ubuntu/.m2/repository/com/fasterxml/jackson/core/jackson-databind/2.13.3/jackson-databind-2.13.3.jar:/home/ubuntu/.m2/repository/com/fasterxml/jackson/core/jackson-annotations/2.13.3/jackson-annotations-2.13.3.jar:/home/ubuntu/.m2/repository/com/fasterxml/jackson/core/jackson-core/2.13.3/jackson-core-2.13.3.jar:/home/ubuntu/.m2/repository/com/fasterxml/jackson/datatype/jackson-datatype-jdk8/2.13.3/jackson-datatype-jdk8-2.13.3.jar:/home/ubuntu/.m2/repository/com/fasterxml/jackson/datatype/jackson-datatype-jsr310/2.13.3/jackson-datatype-jsr310-2.13.3.jar:/home/ubuntu/.m2/repository/com/fasterxml/jackson/module/jackson-module-parameter-names/2.13.3/jackson-module-parameter-names-2.13.3.jar:/home/ubuntu/.m2/repository/org/springframework/boot/spring-boot-starter-tomcat/2.7.2/spring-boot-starter-tomcat-2.7.2.jar:/home/ubuntu/.m2/repository/org/apache/tomcat/embed/tomcat-embed-core/9.0.65/tomcat-embed-core-9.0.65.jar:/home/ubuntu/.m2/repository/org/apache/tomcat/embed/tomcat-embed-el/9.0.65/tomcat-embed-el-9.0.65.jar:/home/ubuntu/.m2/repository/org/apache/tomcat/embed/tomcat-embed-websocket/9.0.65/tomcat-embed-websocket-9.0.65.jar:/home/ubuntu/.m2/repository/org/springframework/spring-web/5.3.22/spring-web-5.3.22.jar:/home/ubuntu/.m2/repository/org/springframework/spring-beans/5.3.22/spring-beans-5.3.22.jar:/home/ubuntu/.m2/repository/org/springframework/spring-webmvc/5.3.22/spring-webmvc-5.3.22.jar:/home/ubuntu/.m2/repository/org/springframework/spring-aop/5.3.22/spring-aop-5.3.22.jar:/home/ubuntu/.m2/repository/org/springframework/spring-expression/5.3.22/spring-expression-5.3.22.jar:/home/ubuntu/.m2/repository/org/springframework/boot/spring-boot-devtools/2.7.2/spring-boot-devtools-2.7.2.jar:/home/ubuntu/.m2/repository/org/springframework/boot/spring-boot-configuration-processor/2.7.2/spring-boot-configuration-processor-2.7.2.jar:/home/ubuntu/.m2/repository/org/projectlombok/lombok/1.18.24/lombok-1.18.24.jar \
io.lihongbin.remote.debug.test.RemoteDebugTestApplication

*/

// 需要配置 MAVEN_HOME 环境变量, 否则需要在程序启动参数指定 -maven
// 需要配置 JAVA_HOME 环境变量, 否则需要在程序启动参数指定 -java
// 需要先对父工程 mvn install
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

	// 创建编译路径
	destPath := fmt.Sprintf("%s/target/remote/classes", projectPath)
	if err := os.MkdirAll(destPath, 0777); err != nil {
		exit("mkdir error", err)
	}

	// 生成 编译命令 和 运行命令
	allClass := getAllClass()
	compileCommand := ""
	runCommand := ""
	if len(classpath) > 0 {
		compileCommand = fmt.Sprintf("\"%s\" -classpath \"%s\" -d \"%s\"%s", javac, classpath, destPath, allClass)
		runCommand = fmt.Sprintf("\"%s\" -Dfile.encoding=UTF-8 -classpath \"%s%c%s\" %s", java, destPath, os.PathListSeparator, classpath, runClass)
	} else {
		compileCommand = fmt.Sprintf("\"%s\" -d \"%s%s\"", javac, destPath, allClass)
		runCommand = fmt.Sprintf("\"%s\" -Dfile.encoding=UTF-8 %s", java, runClass)
	}

	fmt.Println(compileCommand)
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

	// 获取 maven 返回的依赖列表
	cmd := &exec.Cmd{
		Path: mvn,
		Args: []string{"mvn", "dependency:list"},
		Dir:  fmt.Sprintf("%s/%s", projectPath, modulePath),
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

	// 分割每一行依赖项
	rows := strings.Split(string(mavenResult), "\n")
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
		exit("mvn dependency list error", fmt.Errorf("%s", string(mavenResult)))
	}

	return mavenDependencyList
}

func getAllClass() string {
	path := fmt.Sprintf("%s/%s/src/main/java", projectPath, modulePath)
	moduleDir, err := os.ReadDir(path)
	if err != nil {
		exit("open module path error", err)
	}
	return doGetAllClass("", path, fmt.Sprintf("%s/src/main/java", modulePath), moduleDir)
}

func doGetAllClass(result, path, classPath string, dir []os.DirEntry) string {
	for _, fileInfo := range dir {
		if fileInfo.IsDir() {
			classPath = fmt.Sprintf("%s/%s", classPath, fileInfo.Name())
			subDirPath := fmt.Sprintf("%s/%s", path, fileInfo.Name())
			subDir, err := os.ReadDir(subDirPath)
			if err != nil {
				exit("open sub dir error", err)
			}
			result = doGetAllClass(result, subDirPath, classPath, subDir)
		} else if strings.HasSuffix(fileInfo.Name(), ".java") {
			result = fmt.Sprintf("%s %s/%s", result, classPath, fileInfo.Name())
		}
	}
	return result
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
