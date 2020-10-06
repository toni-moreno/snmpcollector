// +build ignore

package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	versionRe = regexp.MustCompile(`-[0-9]{1,3}-g[0-9a-f]{5,10}`)
	goarch    string
	goos      string
	version   string = "v1"
	// deb & rpm does not support semver so have to handle their version a little differently
	linuxPackageVersion   string = "v1"
	linuxPackageIteration string = ""
	race                  bool   = false
	workingDir            string
	serverBinaryName      string = "snmpcollector"
)

const minGoVersion = 1.6

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	ensureGoPath()
	//readVersionFromPackageJson()
	readVersionFromGit()

	log.Printf("Version: %s, Linux Version: %s, Package Iteration: %s\n", version, linuxPackageVersion, linuxPackageIteration)

	flag.StringVar(&goarch, "goarch", runtime.GOARCH, "GOARCH")
	flag.StringVar(&goos, "goos", runtime.GOOS, "GOOS")
	flag.BoolVar(&race, "race", race, "Use race detector")
	flag.Parse()

	if flag.NArg() == 0 {
		log.Println("Usage: go run build.go build")
		return
	}

	workingDir, _ = os.Getwd()

	for _, cmd := range flag.Args() {
		switch cmd {
		case "build":
			pkg := "./pkg/"
			clean()
			build(pkg, []string{}, []string{})

		case "build-static":
			pkg := "./pkg/"
			clean()
			build(pkg, []string{}, []string{"-linkmode", "external", "-extldflags", "-static"})
			//"-linkmode external -extldflags -static"
		case "test":
			test("./pkg/...")

		case "package":
			os.Mkdir("./dist", 0755)
			createLinuxPackages()
			sha1FilesInDist()
		case "pkg-min-tar":
			os.Mkdir("./dist", 0755)
			createMinTar()
			sha1FilesInDist()
		case "pkg-rpm":
			os.Mkdir("./dist", 0755)
			createRpmPackages()
			sha1FilesInDist()
		case "pkg-deb":
			os.Mkdir("./dist", 0755)
			createDebPackages()
			sha1FilesInDist()
		case "sha1-dist":
			sha1FilesInDist()
		case "latest":
			os.Mkdir("./dist", 0755)
			createLinuxPackages()
			makeLatestDistCopies()
			sha1FilesInDist()

		case "clean":
			clean()

		default:
			log.Fatalf("Unknown command %q", cmd)
		}
	}
}

func makeLatestDistCopies() {
	rpmIteration := "-1"
	if linuxPackageIteration != "" {
		rpmIteration = "-" + linuxPackageIteration
	}

	runError("cp", "dist/snmpcollector_"+version+"_amd64.deb", "dist/snmpcollector_latest_amd64.deb")
	runError("cp", "dist/snmpcollector-"+linuxPackageVersion+rpmIteration+".x86_64.rpm", "dist/snmpcollector-latest-1.x86_64.rpm")
	runError("cp", "dist/snmpcollector-"+version+".linux-x64.tar.gz", "dist/snmpcollector-latest.linux-x64.tar.gz")
}

func readVersionFromPackageJson() {
	reader, err := os.Open("package.json")
	if err != nil {
		log.Fatal("Failed to open package.json")
		return
	}
	defer reader.Close()

	jsonObj := map[string]interface{}{}
	jsonParser := json.NewDecoder(reader)

	if err := jsonParser.Decode(&jsonObj); err != nil {
		log.Fatal("Failed to decode package.json")
	}

	version = jsonObj["version"].(string)
	linuxPackageVersion = version
	linuxPackageIteration = ""

	// handle pre version stuff (deb / rpm does not support semver)
	parts := strings.Split(version, "-")

	if len(parts) > 1 {
		linuxPackageVersion = parts[0]
		linuxPackageIteration = parts[1]
	}
}

func readVersionFromGit() {
	cmd := "git describe --abbrev=0 --tag"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatal(err)
	}

	linuxPackageVersion = strings.TrimSpace(string(out))
	version = linuxPackageVersion
	linuxPackageIteration = ""

	// handle pre version stuff (deb / rpm does not support semver)
	parts := strings.Split(version, "-")

	if len(parts) > 1 {
		linuxPackageVersion = parts[0]
		linuxPackageIteration = parts[1]
	}
}

type linuxPackageOptions struct {
	packageType            string
	homeDir                string
	binPath                string
	configDir              string
	configFilePath         string
	etcDefaultPath         string
	etcDefaultFilePath     string
	initdScriptFilePath    string
	systemdServiceFilePath string

	postinstSrc    string
	initdScriptSrc string
	defaultFileSrc string
	systemdFileSrc string

	depends []string
}

func createDebPackages() {
	createFpmPackage(linuxPackageOptions{
		packageType:            "deb",
		homeDir:                "/usr/share/snmpcollector",
		binPath:                "/usr/sbin/snmpcollector",
		configDir:              "/etc/snmpcollector",
		configFilePath:         "/etc/snmpcollector/config.toml",
		etcDefaultPath:         "/etc/default",
		etcDefaultFilePath:     "/etc/default/snmpcollector",
		initdScriptFilePath:    "/etc/init.d/snmpcollector",
		systemdServiceFilePath: "/usr/lib/systemd/system/snmpcollector.service",

		postinstSrc:    "packaging/deb/control/postinst",
		initdScriptSrc: "packaging/deb/init.d/snmpcollector",
		defaultFileSrc: "packaging/deb/default/snmpcollector",
		systemdFileSrc: "packaging/deb/systemd/snmpcollector.service",

		depends: []string{"adduser"},
	})
}

func createRpmPackages() {
	createFpmPackage(linuxPackageOptions{
		packageType:            "rpm",
		homeDir:                "/usr/share/snmpcollector",
		binPath:                "/usr/sbin/snmpcollector",
		configDir:              "/etc/snmpcollector",
		configFilePath:         "/etc/snmpcollector/config.toml",
		etcDefaultPath:         "/etc/sysconfig",
		etcDefaultFilePath:     "/etc/sysconfig/snmpcollector",
		initdScriptFilePath:    "/etc/init.d/snmpcollector",
		systemdServiceFilePath: "/usr/lib/systemd/system/snmpcollector.service",

		postinstSrc:    "packaging/rpm/control/postinst",
		initdScriptSrc: "packaging/rpm/init.d/snmpcollector",
		defaultFileSrc: "packaging/rpm/sysconfig/snmpcollector",
		systemdFileSrc: "packaging/rpm/systemd/snmpcollector.service",

		depends: []string{"initscripts"},
	})
}

func createLinuxPackages() {
	createDebPackages()
	createRpmPackages()
}

func createMinTar() {
	packageRoot, _ := ioutil.TempDir("", "snmpcollector-linux-pack")
	// create directories
	runPrint("mkdir", "-p", filepath.Join(packageRoot, "/opt/snmpcollector"))
	runPrint("mkdir", "-p", filepath.Join(packageRoot, "/opt/snmpcollector/conf"))
	runPrint("mkdir", "-p", filepath.Join(packageRoot, "/opt/snmpcollector/bin"))
	runPrint("mkdir", "-p", filepath.Join(packageRoot, "/opt/snmpcollector/log"))
	runPrint("mkdir", "-p", filepath.Join(packageRoot, "/opt/snmpcollector/public"))
	runPrint("cp", "conf/sample.config.toml", filepath.Join(packageRoot, "/opt/snmpcollector/conf"))
	runPrint("cp", "bin/snmpcollector", filepath.Join(packageRoot, "/opt/snmpcollector/bin"))
	runPrint("cp", "bin/snmpcollector.md5", filepath.Join(packageRoot, "/opt/snmpcollector/bin"))
	runPrint("cp", "-a", filepath.Join(workingDir, "public")+"/.", filepath.Join(packageRoot, "/opt/snmpcollector/public"))
	tarname := fmt.Sprintf("dist/snmpcollector-%s-%s_%s_%s.tar.gz", version, getGitSha(), runtime.GOOS, runtime.GOARCH)
	runPrint("tar", "zcvf", tarname, "-C", packageRoot, ".")
	runPrint("rm", "-rf", packageRoot)
}

func createFpmPackage(options linuxPackageOptions) {
	packageRoot, _ := ioutil.TempDir("", "snmpcollector-linux-pack")

	// create directories
	runPrint("mkdir", "-p", filepath.Join(packageRoot, options.homeDir))
	runPrint("mkdir", "-p", filepath.Join(packageRoot, options.configDir))
	runPrint("mkdir", "-p", filepath.Join(packageRoot, "/etc/init.d"))
	runPrint("mkdir", "-p", filepath.Join(packageRoot, options.etcDefaultPath))
	runPrint("mkdir", "-p", filepath.Join(packageRoot, "/usr/lib/systemd/system"))
	runPrint("mkdir", "-p", filepath.Join(packageRoot, "/usr/sbin"))

	// copy binary
	runPrint("cp", "-p", filepath.Join(workingDir, "bin/"+serverBinaryName), filepath.Join(packageRoot, options.binPath))
	// copy init.d script
	runPrint("cp", "-p", options.initdScriptSrc, filepath.Join(packageRoot, options.initdScriptFilePath))
	// copy environment var file
	runPrint("cp", "-p", options.defaultFileSrc, filepath.Join(packageRoot, options.etcDefaultFilePath))
	// copy systemd filerunPrint("cp", "-a", filepath.Join(workingDir, "tmp")+"/.", filepath.Join(packageRoot, options.homeDir))
	runPrint("cp", "-p", options.systemdFileSrc, filepath.Join(packageRoot, options.systemdServiceFilePath))
	// copy release files
	runPrint("cp", "-a", filepath.Join(workingDir+"/public"), filepath.Join(packageRoot, options.homeDir))
	// remove bin path
	runPrint("rm", "-rf", filepath.Join(packageRoot, options.homeDir, "bin"))
	// copy sample ini file to /etc/snmpcollector
	runPrint("cp", "conf/sample.config.toml", filepath.Join(packageRoot, options.configFilePath))

	args := []string{
		"-s", "dir",
		"--description", "A full featured Generic SNMP data collector with Web Administration Interface",
		"-C", packageRoot,
		"--vendor", "snmpcollector",
		"--url", "http://github.org/toni-moreno/snmpcollector",
		"--license", "Apache 2.0",
		"--maintainer", "toni.moreno@gmail.com",
		"--config-files", options.configFilePath,
		"--config-files", options.initdScriptFilePath,
		"--config-files", options.etcDefaultFilePath,
		"--config-files", options.systemdServiceFilePath,
		"--after-install", options.postinstSrc,
		"--name", "snmpcollector",
		"--version", linuxPackageVersion,
		"-p", "./dist",
	}

	if linuxPackageIteration != "" {
		args = append(args, "--iteration", linuxPackageIteration)
	}

	// add dependenciesj
	for _, dep := range options.depends {
		args = append(args, "--depends", dep)
	}

	args = append(args, ".")

	fmt.Println("Creating package: ", options.packageType)
	runPrint("fpm", append([]string{"-t", options.packageType}, args...)...)
}

func verifyGitRepoIsClean() {
	rs, err := runError("git", "ls-files", "--modified")
	if err != nil {
		log.Fatalf("Failed to check if git tree was clean, %v, %v\n", string(rs), err)
		return
	}
	count := len(string(rs))
	if count > 0 {
		log.Fatalf("Git repository has modified files, aborting")
	}

	log.Println("Git repository is clean")
}

func ensureGoPath() {
	if os.Getenv("GOPATH") == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		gopath := filepath.Clean(filepath.Join(cwd, "../../../../"))
		log.Println("GOPATH is", gopath)
		os.Setenv("GOPATH", gopath)
	}
}

func ChangeWorkingDir(dir string) {
	os.Chdir(dir)
}

func test(pkg string) {
	setBuildEnv()
	runPrint("go", "test", "-short", "-timeout", "60s", pkg)
}

func build(pkg string, tags []string, flags []string) {
	binary := "./bin/" + serverBinaryName
	if goos == "windows" {
		binary += ".exe"
	}

	rmr(binary, binary+".md5")
	args := []string{"build", "-ldflags", ldflags(flags)}
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}
	if race {
		args = append(args, "-race")
	}
	args = append(args, "-v")
	args = append(args, "-o", binary)
	args = append(args, pkg)
	setBuildEnv()

	runPrint("go", "version")
	runPrint("go", args...)

	// Create an md5 checksum of the binary, to be included in the archive for
	// automatic upgrades.
	err := md5File(binary)
	if err != nil {
		log.Fatal(err)
	}
}

func ldflags(flags []string) string {
	var b bytes.Buffer
	b.WriteString("-w")
	b.WriteString(fmt.Sprintf(" -X github.com/toni-moreno/snmpcollector/pkg/agent.Version=%s", version))
	b.WriteString(fmt.Sprintf(" -X github.com/toni-moreno/snmpcollector/pkg/agent.Commit=%s", getGitSha()))
	b.WriteString(fmt.Sprintf(" -X github.com/toni-moreno/snmpcollector/pkg/agent.BuildStamp=%d", buildStamp()))
	for _, f := range flags {
		b.WriteString(fmt.Sprintf(" %s", f))
	}
	return b.String()
}

func rmr(paths ...string) {
	for _, path := range paths {
		log.Println("rm -r", path)
		os.RemoveAll(path)
	}
}

func clean() {
	//	rmr("bin", "Godeps/_workspace/pkg", "Godeps/_workspace/bin")
	rmr("public")
	//rmr("tmp")
	rmr(filepath.Join(os.Getenv("GOPATH"), fmt.Sprintf("pkg/%s_%s/github.com/toni-moreno/snmpcollector", goos, goarch)))
}

func setBuildEnv() {
	os.Setenv("GOOS", goos)
	if strings.HasPrefix(goarch, "armv") {
		os.Setenv("GOARCH", "arm")
		os.Setenv("GOARM", goarch[4:])
	} else {
		os.Setenv("GOARCH", goarch)
	}
	if goarch == "386" {
		os.Setenv("GO386", "387")
	}
	/*wd, err := os.Getwd()
	if err != nil {
		log.Println("Warning: can't determine current dir:", err)
		log.Println("Build might not work as expected")
	}
	os.Setenv("GOPATH", fmt.Sprintf("%s%c%s", filepath.Join(wd, "Godeps", "_workspace"), os.PathListSeparator, os.Getenv("GOPATH")))*/
	log.Println("GOPATH=" + os.Getenv("GOPATH"))
}

func getGitSha() string {
	v, err := runError("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "unknown-dev"
	}
	v = versionRe.ReplaceAllFunc(v, func(s []byte) []byte {
		s[0] = '+'
		return s
	})
	return string(v)
}

func buildStamp() int64 {
	bs, err := runError("git", "show", "-s", "--format=%ct")
	if err != nil {
		return time.Now().Unix()
	}
	s, _ := strconv.ParseInt(string(bs), 10, 64)
	return s
}

func buildArch() string {
	os := goos
	if os == "darwin" {
		os = "macosx"
	}
	return fmt.Sprintf("%s-%s", os, goarch)
}

func run(cmd string, args ...string) []byte {
	bs, err := runError(cmd, args...)
	if err != nil {
		log.Println(cmd, strings.Join(args, " "))
		log.Println(string(bs))
		log.Fatal(err)
	}
	return bytes.TrimSpace(bs)
}

func runError(cmd string, args ...string) ([]byte, error) {
	ecmd := exec.Command(cmd, args...)
	bs, err := ecmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return bytes.TrimSpace(bs), nil
}

func runPrint(cmd string, args ...string) {
	log.Println(cmd, strings.Join(args, " "))
	ecmd := exec.Command(cmd, args...)
	ecmd.Stdout = os.Stdout
	ecmd.Stderr = os.Stderr
	err := ecmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func md5File(file string) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	h := md5.New()
	_, err = io.Copy(h, fd)
	if err != nil {
		return err
	}

	out, err := os.Create(file + ".md5")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "%x\n", h.Sum(nil))
	if err != nil {
		return err
	}

	return out.Close()
}

func sha1FilesInDist() {
	filepath.Walk("./dist", func(path string, f os.FileInfo, err error) error {
		if strings.Contains(path, ".sha1") == false {
			sha1File(path)
		}
		return nil
	})
}

func sha1File(file string) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	h := sha1.New()
	_, err = io.Copy(h, fd)
	if err != nil {
		return err
	}

	out, err := os.Create(file + ".sha1")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "%x\n", h.Sum(nil))
	if err != nil {
		return err
	}

	return out.Close()
}
