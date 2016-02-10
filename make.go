// make.go
//
// The rules for modifying this file is to only depend on standard go packages
// so that a build process can ge bootstrapped even if other things are not
// installed yet.
package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"text/template"
)

// Flags
var (
	// build logging output and help flags
	verboseFlag bool
	helpFlag    bool

	// misc
	xcnFlag      bool
	debugFlag    bool
	optimizeFlag bool
	dockerFlag   bool

	// build task filters
	noTestsFlag bool
	offlineFlag bool

	// build target filters
	osFlag   string
	archFlag string
	cmdFlag  string

	cmdFilter  = make(map[string]bool, 0)
	osFilter   = make(map[string]bool, 0)
	archFilter = make(map[string]bool, 0)
)

var tasks map[string]func()

func init() {
	if cpu := runtime.NumCPU(); cpu == 1 {
		runtime.GOMAXPROCS(2)
	} else {
		runtime.GOMAXPROCS(cpu)
	}

	// registry of all tasks
	tasks = map[string]func(){
		"all":                  AllTask,
		"dev":                  DevTask,
		"bindata":              BindataTask,
		"bindata-dev":          BindataDevTask,
		"browser":              BrowserTask,
		"docs":                 DocsTask,
		"hot":                  HotTask,
		"hot-build":            HotBuildTask,
		"bumpversion-patch":    BumpVersionPatchTask,
		"clean":                CleanTask,
		"deps":                 DepsTask,
		"dist":                 BuildTask,
		"dist-build":           DistBuildTask,
		"dist-build-go":        DistBuildGoTask,
		"release":              ReleaseTask,
		"releaseChromeExt":     ReleaseChromeExtTask,
		"tasks":                TasksTask,
		"test":                 TestTask,
		"test-all":             TestAllTask,
		"fmt":                  FmtTask,
		"ci":                   CITask,
		"genMakefile":          GenMakefileTask,
		"lint":                 LintTask,
		"govet":                GoVetTask,
		"chrome":               ChromeTask,
		"chrome-copy-messages": ChromeCopyMessagesTask,
		"translations-fixup":   TranslationsFixupTask,
	}

	// all flags
	flag.BoolVar(&noTestsFlag, "notests", false, "Do not run any tests")
	flag.BoolVar(&offlineFlag, "offline", false, "Do not run any tasks that require internet access")

	flag.StringVar(&osFlag, "os", "", "filter which OS to build")
	flag.StringVar(&archFlag, "arch", "", "filter which ARCH to buil")
	flag.StringVar(&cmdFlag, "cmd", "", "filter which CMD to build")
	flag.BoolVar(&helpFlag, "h", false, "Displays help")
	flag.BoolVar(&xcnFlag, "xcn", false, "Use native toolchains for cross platoform builds")
	flag.BoolVar(&dockerFlag, "docker", false, "Run builds inside a the docker builder container")
	flag.BoolVar(&verboseFlag, "verbose", false, "Verbose output")
	flag.BoolVar(&debugFlag, "debug", false, "Produce a debug/ verbose builds")
	flag.BoolVar(&optimizeFlag, "optimize", false, "Produce optimized buulds")
}

// Run runs a named build task
func Run(task string) {
	fn, ok := tasks[task]
	if !ok {
		fmt.Printf("%s is not an task", task)
		panic("")
	}
	log.Printf(">>> [%s]", task)
	fn()
	log.Printf("<<< [%s]", task)
}

// cleanupDirs are erased erased before exiting.
var cleanupDirs = make([]string, 0)

func main() {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println("Build paniced:", r)
		}
		var wg sync.WaitGroup
		for _, d := range cleanupDirs {
			if verboseFlag {
				log.Printf("Removing build dir: %s", d)
			}
			wg.Add(1)
			go func(dir string) {
				defer wg.Done()
				os.RemoveAll(dir)
			}(d)
			wg.Wait()
		}
		if r != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}()
	task := "tasks"
	if len(os.Args) > 1 {
		if !strings.HasPrefix(os.Args[1], "-") {
			task = os.Args[1]
			var args []string
			for k, v := range os.Args {
				if k == 1 {
					continue
				}
				args = append(args, v)
			}
			os.Args = args
		}
	}
	flag.Parse()
	for _, v := range strings.Split(cmdFlag, ",") {
		cmdFilter[v] = true
	}
	for _, v := range strings.Split(osFlag, ",") {
		osFilter[v] = true
	}
	for _, v := range strings.Split(archFlag, ",") {
		archFilter[v] = true
	}
	Run(task)
}

// newCmd returns a command with stdout/stderr attached to os.Std...
func newCmd(cmd string, arg ...string) *exec.Cmd {
	c := exec.Command(cmd, arg...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c
}

// runCmd runs a command and panics on non 0 exit
func runCmd(cmd string, arg ...string) {
	err := newCmd(cmd, arg...).Run()
	failErr(err, fmt.Sprintf("Failed to run command %s %+v", cmd, arg))
}

// runCmdIgnoreErr runs a command and prints an log message on non zero exit.
func runCmdIgnoreErr(cmd string, arg ...string) {
	err := newCmd(cmd, arg...).Run()
	if err != nil {
		log.Printf("*** Failed to run command %s %+v", cmd, arg)
	}
}

// fail aborts the build process after printing message
func fail(message string) {
	log.Printf("FAILED: %s", message)
	panic(message)
}

// failErr aborts the build process if err is not nil
func failErr(err error, message string) {
	if err != nil {
		fail(message)
	}
}

// tempDir returns a path to a temporary directory and queues it to be deleted
// when the builder exits.
func tempDir(prefix string) string {
	t, err := ioutil.TempDir("", fmt.Sprintf("as-%s", prefix))
	failErr(err, "Could not create temporary directory")
	cleanupDirs = append(cleanupDirs, t)
	return t
}

// Builder contains all configuration for a building a go cmd
type Builder struct {
	OS           string   // target os.
	Arch         string   // target arch.
	Cmd          string   // target cmd, builds cmd/[cmd]/[cmd].go.
	GUI          bool     // build without creating a text console window.
	BinaryName   string   // resulting binary name, autogenerated if empty.
	FlatArchive  bool     // skips creation of top level directory in archives.
	XCN          bool     // true for cross compilation with native tool chains.
	Built        bool     // true if the builder has run and succeeded.
	BrowserBuild bool     // true if the browser/chrome build is required.
	Vars         []string // vars to be injected as ldflags -X ENTRY when building
}

// Build executes the build
func (b *Builder) Build() {
	log.Printf(
		"Building: %s %s %s xcn:%t xcnF:%t",
		b.Cmd, b.OS, b.Arch, b.XCN, xcnFlag)
	var env []string  // cmd env
	var args []string // cmd args

	env = append(env, fmt.Sprintf("GOOS=%s", b.OS))
	env = append(env, fmt.Sprintf("GOARCH=%s", b.Arch))
	env = append(env, "GO15VENDOREXPERIMENT=1")

	if xcnFlag && b.XCN {
		osarch := fmt.Sprintf("%s-%s", b.OS, b.Arch)
		osArchCCs := map[string]string{
			"darwin-386":    "CC=o32-clang",
			"darwin-amd64":  "CC=o64-clang",
			"linux-386":     "",
			"linux-amd64":   "",
			"windows-386":   "CC=i686-w64-mingw32-gcc",
			"windows-amd64": "CC=x86_64-w64-mingw32-gcc",
		}
		ccenv, ok := osArchCCs[osarch]
		if !ok {
			log.Printf("%s is not supported", osarch)
			panic("")
		}
		if ccenv != "" {
			env = append(env, ccenv)
		}
		env = append(env, "CGO_ENABLED=1")
	}

	args = append(args, "go", "build")
	var ldflags []string
	if optimizeFlag || (xcnFlag && b.OS == "darwin") {
		// workaround because non existent dsymutil in current osxcross
		// https://github.com/golang/go/issues/11994
		ldflags = append(ldflags, "-s")
	}

	if b.GUI && b.OS == "windows" {
		ldflags = append(ldflags, "-H=windowsgui")
	}
	for _, v := range b.Vars {
		ldflags = append(ldflags, fmt.Sprintf("-X %s", v))
	}

	ldflagsStr := strings.Join(ldflags, " ")
	if ldflagsStr != "" {
		args = append(args, "-ldflags", ldflagsStr)
	}

	if verboseFlag {
		// args = append(args, "-v")
		// args = append(args, "-a")
		// args = append(args, "-x")
	}

	args = append(args, "-o", b.Filename())
	args = append(args, fmt.Sprintf("cmd/%s/%s.go", b.Cmd, b.Cmd))

	var c *exec.Cmd
	if !dockerFlag {
		mergedEnv := os.Environ()
		mergedEnv = append(mergedEnv, env...)
		c = newCmd(args[0], args[1:]...)
		c.Env = mergedEnv
	} else {
		var dargs []string
		dargs = append(dargs, "docker", "run", "--rm")
		u, err := user.Current()
		failErr(err, "Could not get user id")
		dargs = append(dargs, "-u", u.Uid)

		wd, err := os.Getwd()
		failErr(err, "Could not read working directory")
		dargs = append(dargs, fmt.Sprintf("-v=%s:/go/src/github.com/alkasir/alkasir/", wd))

		for _, v := range env {
			dargs = append(dargs, "-e", v)
		}
		dargs = append(dargs, "alkasir-docker-xcn-builder")
		args = append(dargs, args...)
		c = newCmd(args[0], args[1:]...)
	}
	if verboseFlag {
		log.Printf("*** args:%+v env:%+v", args, env)
	}
	err := c.Run()
	failErr(err, "error executing builder.Run")
	b.Built = true
}

// Filename returns the output binary filename
func (b *Builder) Filename() string {
	xOsExts := map[string]string{
		"darwin":  "",
		"linux":   "",
		"windows": ".exe",
	}
	binaryName := fmt.Sprintf("%s_%s_%s%s",
		b.Cmd, b.OS, b.Arch, xOsExts[b.OS])
	if b.BinaryName != "" {
		binaryName = fmt.Sprintf("%s%s",
			b.BinaryName, xOsExts[b.OS])
	}
	return filepath.Join(
		"./dist/",
		fmt.Sprintf("%s-%s-%s",
			b.Cmd, b.OS, b.Arch),
		binaryName,
	)
}

// ShouldBuild returns false if this build should be filtered according to
// filter flags.
func (b *Builder) ShouldBuild() bool {
	if osFlag != "" && !osFilter[b.OS] {
		return false
	}
	if archFlag != "" && !archFilter[b.Arch] {
		return false
	}
	if cmdFlag != "" && !cmdFilter[b.Cmd] {
		return false
	}
	return true
}

// Artifact returns a build artifact name
func (b *Builder) Artifact(postfixes ...string) string {
	postfix := ""
	if len(postfixes) > 0 {
		postfix = fmt.Sprintf("%s-", strings.Join(postfixes, "-"))
	}
	return fmt.Sprintf("%s_%s%s-%s",
		b.Cmd, postfix, b.OS, b.Arch)
}

func (b *Builder) ReleaseArchive() {
	if !b.Built {
		fail("Cannot create archives because not built")
	}

	archivesdir := "dist-archives"
	failErr(os.MkdirAll(archivesdir, 0775), "could not create dist-archives")
	var archive string
	switch b.OS {
	case "windows":
		archive = filepath.Join(archivesdir, b.Artifact()+".zip")
		log.Printf("Creating zip archive for %s: %s", b.Artifact(), archive)
		err := makeZip(archive, filepath.Dir(b.Filename()))
		failErr(err, "could not create zip file")
	case "darwin", "linux":
		archive = filepath.Join(archivesdir, b.Artifact()+".tar.gz")
		log.Printf("Creating tar.gz archive for %s: %s", b.Artifact(), archive)
		prefix := filepath.Base(filepath.Dir(b.Filename()))
		if b.FlatArchive {
			prefix = ""
		}
		err := makeTar(archive, filepath.Dir(b.Filename()), prefix)
		failErr(err, "Could not create tar.gz")
	default:
		fail(fmt.Sprintf("OS %s not supported", b.OS))
	}

	if b.OS == "darwin" && xcnFlag && b.XCN && b.GUI {
		artifact := b.Artifact("osxapp")
		archive := filepath.Join(archivesdir, artifact+".zip")
		log.Printf("Creating osxapp.zip archive for %s: %s", artifact, archive)
		err := b.makeOSXApp(archive)
		if err != nil {
			panic(err)
		}

		failErr(err, "Could not create osxapp archive")
	}
}

// Builders is a list of Builder instances with convinience functions.
type Builders []*Builder

// needsBrowser returns true if one or more of the builders requires the
// browser code to be compiled.
func (b Builders) needsBrowser() bool {
	for _, v := range b {
		if v.BrowserBuild {
			return true
		}
	}
	return false
}

// RunAll runs all builders
func (bs *Builders) RunAll() {
	for _, b := range *bs {
		b.Build()
	}
}

// UploadAll upload all build artifacts
func (bs *Builders) CreateReleaseArchives() {
	for _, b := range *bs {
		b.ReleaseArchive()
	}
}

// getGoBuilders returns a filtered list of all go builds to run depeding on
// flags and build environment.
func getGoBuilders() Builders {
	defaultBuilders := []Builder{
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "386"},
		{OS: "windows", Arch: "amd64"},
		{OS: "windows", Arch: "386"},
		{OS: "darwin", Arch: "amd64"},
		{OS: "darwin", Arch: "386"},
	}
	var bs []*Builder

	var clientBuildVars []string
	for envkey, fqn := range map[string]string{
		"ALKASIR_CLIENT_VERSION":     "github.com/alkasir/alkasir/pkg/client.VERSION",
		"ALKASIR_CLIENT_DIFF_URL":    "github.com/alkasir/alkasir/pkg/client.upgradeDiffsBaseURL",
		"ALKASIR_CLIENT_CENTRAL_URL": "github.com/alkasir/alkasir/pkg/client/internal/config.centralAddr",
	} {
		envval := os.Getenv(envkey)
		if envval != "" {
			clientBuildVars = append(clientBuildVars, fmt.Sprintf("%s=%s", fqn, envval))
		}
	}

	for _, bb := range defaultBuilders {
		b := &Builder{
			Cmd:          "alkasir-client",
			OS:           bb.OS,
			Arch:         bb.Arch,
			BrowserBuild: true,
			Vars:         clientBuildVars,
		}
		if b.ShouldBuild() {
			bs = append(bs, b)
		}
	}
	for _, bb := range defaultBuilders {
		b := &Builder{
			Cmd:  "alkasir-admin",
			OS:   bb.OS,
			Arch: bb.Arch,
		}
		if b.ShouldBuild() {
			bs = append(bs, b)
		}
	}

	if xcnFlag {
		for _, bb := range defaultBuilders {
			if bb.OS == "linux" && bb.Arch == "386" {
				continue
			}

			// ARC is used in the tray icon library which only exists for 64bit.
			if bb.OS == "darwin" && bb.Arch == "386" {
				continue
			}
			b := &Builder{
				Cmd:          "alkasir-gui",
				OS:           bb.OS,
				Arch:         bb.Arch,
				XCN:          true,
				GUI:          true,
				BinaryName:   "alkasir",
				FlatArchive:  true,
				BrowserBuild: true,
				Vars:         clientBuildVars,
			}
			if b.ShouldBuild() {
				bs = append(bs, b)
			}
		}
	}

	centralBuilder := &Builder{
		Cmd:        "alkasir-central",
		OS:         "linux",
		Arch:       "amd64",
		BinaryName: "alkasir-central",
	}
	if centralBuilder.ShouldBuild() {
		bs = append(bs, centralBuilder)
	}
	return bs
}

var osxPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleGetInfoString</key>
  <string>Alkasir</string>
  <key>CFBundleExecutable</key>
  <string>{{.Cmd}}</string>
  <key>CFBundleIdentifier</key>
  <string>com.alkasir.www</string>
  <key>CFBundleName</key>
  <string>Alkasir</string>
  <key>CFBundleIconFile</key>
  <string>alkasir.icns</string>
  <key>CFBundleShortVersionString</key>
  <string>1.0</string>
  <key>CFBundleInfoDictionaryVersion</key>
  <string>6.0</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>IFMajorVersion</key>
  <integer>0</integer>
  <key>IFMinorVersion</key>
  <integer>1</integer>
</dict>
</plist>`

// makeOSXApp packages an OSX App bundle
func (b *Builder) makeOSXApp(archive string) error {
	workdir := tempDir("osxapp")
	contents := filepath.Join(workdir, "Alkasir.app", "Contents")
	os.MkdirAll(filepath.Join(contents, "Resources"), 0755)
	os.MkdirAll(filepath.Join(contents, "MacOS"), 0755)

	plistT, err := template.New("plist").Parse(osxPlistTemplate)
	if err != nil {
		return errors.New("Could not parse plist template")
	}
	plistF, err := os.Create(filepath.Join(contents, "info.plist"))
	if err != nil {
		return errors.New("Could open plist file for writing")
	}
	err = plistT.Execute(plistF, b)
	if err != nil {
		return errors.New("plist file write error")
	}
	err = plistF.Close()
	if err != nil {
		return errors.New("plist file close error")
	}

	appBin := filepath.Join(contents, "MacOS", b.Cmd)
	err = cp(appBin, b.Filename())
	if err != nil {
		return errors.New("Could not copy build binary")
	}
	os.Chmod(appBin, 0775)
	err = cp(
		filepath.Join(contents, "Resources", "alkasir.icns"),
		filepath.Join("res-src", "osx-app", "alkasir.icns"))
	if err != nil {
		return errors.New("Could not copy icons")
	}
	err = makeZip(archive, workdir)
	if err != nil {
		return err
	}
	return nil
}

// copy file (does not copy attributes)
func cp(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

// makeTar creates a tar archive. More or less a straight copy of
// makerelease.go from the go source tree.
func makeTar(targ, workdir, prefix string) error {

	f, err := os.Create(targ)
	if err != nil {
		return err
	}
	zout := gzip.NewWriter(f)
	tw := tar.NewWriter(zout)

	err = filepath.Walk(workdir, func(path string, fi os.FileInfo, err error) error {

		if !strings.HasPrefix(path, workdir) {
			log.Panicf("walked filename %q doesn't begin with workdir %q", path, workdir)
		}
		name := path[len(workdir):]

		// Chop of any leading / from filename, leftover from removing workdir.
		if strings.HasPrefix(name, "/") {
			name = name[1:]
		}
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix = prefix + "/"
		}
		name = prefix + name
		if verboseFlag {
			fmt.Println(name)
			fmt.Println(fi)
		}

		target, _ := os.Readlink(path)
		hdr, err := tar.FileInfoHeader(fi, target)
		if err != nil {
			return err
		}
		hdr.Name = name
		hdr.Uname = "root"
		hdr.Gname = "root"
		hdr.Uid = 0
		hdr.Gid = 0

		// Force permissions to 0755 for executables, 0644 for everything else.
		if fi.Mode().Perm()&0111 != 0 {
			hdr.Mode = hdr.Mode&^0777 | 0755
		} else {
			hdr.Mode = hdr.Mode&^0777 | 0644
		}

		err = tw.WriteHeader(hdr)
		if err != nil {
			return fmt.Errorf("Error writing file %q: %v", name, err)
		}
		if fi.IsDir() {
			return nil
		}
		r, err := os.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		_, err = io.Copy(tw, r)
		return err
	})
	if err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := zout.Close(); err != nil {
		return err
	}
	return f.Close()
}

// makeZip creates a zip archive. More or less a straight copy of
// makerelease.go from the go source tree.
func makeZip(targ, workdir string) error {
	f, err := os.Create(targ)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)

	err = filepath.Walk(workdir, func(path string, fi os.FileInfo, err error) error {
		if !strings.HasPrefix(path, workdir) {
			log.Panicf("walked filename %q doesn't begin with workdir %q", path, workdir)
		}
		name := path[len(workdir):]

		// Convert to Unix-style named paths, as that's the
		// type of zip file that archive/zip creates.
		name = strings.Replace(name, "\\", "/", -1)
		// Chop of any leading / from filename, leftover from removing workdir.
		if strings.HasPrefix(name, "/") {
			name = name[1:]
		}

		if name == "" {
			return nil
		}

		fh, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}
		fh.Name = name
		fh.Method = zip.Deflate
		if fi.IsDir() {
			fh.Name += "/"        // append trailing slash
			fh.Method = zip.Store // no need to deflate 0 byte files
		}
		w, err := zw.CreateHeader(fh)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		r, err := os.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		_, err = io.Copy(w, r)
		return err
	})
	if err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return f.Close()
}

// ---------------------------------------
// ---------------T A S K S---------------
// ---------------------------------------

// AllTask runs many other tasks.
func AllTask() {
	Run("clean")
	Run("deps")
	Run("browser")
	Run("chrome")
	Run("bindata")
	Run("test")
	Run("dist-build")
}

// ReleaseTask builds and uploads a specific git tag, this implies xcnFlag.
func ReleaseTask() {
	Run("clean")
	optimizeFlag = true
	xcnFlag = true
	bs := getGoBuilders()
	if bs.needsBrowser() {
		Run("deps")
		Run("browser")
		Run("chrome")
		Run("bindata")
	}
	bs.RunAll()
	bs.CreateReleaseArchives()
}

func ReleaseChromeExtTask() {
	optimizeFlag = true
	Run("clean")
	Run("chrome")

	credCmd := exec.Command("gpg", "--decrypt", "browser/chrome-ext/publisher/publish.json.gpg")
	stdout, err := credCmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	err = credCmd.Start()
	if err != nil {
		panic(err)
	}
	var creds struct {
		ClientID     string
		ClientSecret string
		AppID        string
	}
	jd := json.NewDecoder(stdout)
	err = jd.Decode(&creds)
	if err != nil {
		panic(err)
	}

	// NOTE: Go's arch/zip writes zip files that the chrome webstore api does
	// not accept.
	os.Remove("browser/chrome-ext/src/src.zip")
	zipper := newCmd("zip", "-r", "src", ".")
	zipper.Dir = "browser/chrome-ext/src"
	err = zipper.Run()
	if err != nil {
		panic(err)
	}

	publisher := newCmd("grunt")
	publisher.Dir = "browser/chrome-ext/publisher/"
	env := os.Environ()
	env = append(env, fmt.Sprintf("CLIENT_ID=%s", creds.ClientID))
	env = append(env, fmt.Sprintf("CLIENT_SECRET=%s", creds.ClientSecret))
	env = append(env, fmt.Sprintf("APP_ID=%s", creds.AppID))
	publisher.Env = env
	err = publisher.Run()
	if err != nil {
		panic(err)
	}

}

// DepsTask installs various build dependencies
func DepsTask() {
	if offlineFlag {
		log.Println("*** Skipping: offline ")
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runCmd("go", "get", "-u",
			"github.com/jteeuwen/go-bindata/...",
		)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if verboseFlag {
			runCmd("npm", "install")
		} else {
			runCmd("npm", "install", "--silent")
		}
	}()
	wg.Wait()
}

// FmtTask runs code formatting
func FmtTask() {
	runCmd("gofmt", "-w", "-s", "cmd", "pkg")
	runCmd("goimports", "-w", "cmd", "pkg")
}

// TasksTask prints a list of all tasks and command line options
func TasksTask() {
	fmt.Println("Tasks:")
	fmt.Print(" ")
	for key := range tasks {
		fmt.Print(" ", key)
	}
	fmt.Println("")
	fmt.Println("Switches:")
	flag.PrintDefaults()
}

// TestTask runs all tests
func TestTask() {
	if noTestsFlag {
		log.Println("*** Skipping: notests")
		return
	}
	runCmd("./pkg/pac/make.sh")
	runCmd("go", "test",
		"./cmd/...", "./pkg/...")
}

// TestTask runs all tests
func TestAllTask() {
	if noTestsFlag {
		log.Println("*** Skipping: notests")
		return
	}
	runCmd("./pkg/pac/make.sh")
	runCmd("go", "test", "-tags=\"net databases\"",
		"./cmd/...", "./pkg/...")
}

// GoVetTask runs go vet w. exclusions.
func GoVetTask() {
	ignoredPaths := map[string]bool{
		"pkg/res":                      true,
		"pkg/lg":                       true,
		"pkg/client/ui/wm/platform":    true,
		"cmd/alkasir-torpt-server/ptc": true,
	}
	var alldirs []string
	for _, root := range []string{"pkg", "cmd"} {
		err := filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
			if fi.Mode().IsDir() {
				if _, ok := ignoredPaths[path]; ok {
					return filepath.SkipDir
				}
				if strings.HasPrefix(path, "pkg") ||
					strings.HasPrefix(path, "cmd") {
					alldirs = append(alldirs, path)
				} else {
					return filepath.SkipDir
				}
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	pkgPaths := []string{}
	for _, v := range alldirs {
		pkgPaths = append(pkgPaths, fmt.Sprintf("github.com/alkasir/alkasir/%s", v))
	}
	runCmd("go", append([]string{"vet"}, pkgPaths...)...)
}

// LintTask runs various QA tools
func LintTask() {
	Run("govet")
	runCmd("jsxhint", "browser/")
	runCmd("coffeelint", "browser/")
	runCmd("eslint", "-c", "browser/.eslintrc", "browser/script/")
}

func ChromeCopyMessagesTask() {
	messagefiles := []string{"en", "sv", "fa", "ar", "zh"}
	for _, v := range messagefiles {
		fdst := fmt.Sprintf(
			"browser/chrome-ext/src/_locales/%s/messages.json", v)
		fsrc := fmt.Sprintf("res/messages/%s/messages.json", v)

		failErr(os.MkdirAll(filepath.Dir(fdst), 0775),
			fmt.Sprintf("Could not create dir %s", filepath.Dir(fdst)))
		failErr(cp(fdst, fsrc),
			fmt.Sprintf("Could not copy %s to %s", fsrc, fdst))
	}
}

// ChromeTask compiles the chrome extension
func ChromeTask() {
	Run("chrome-copy-messages")
	var args []string
	args = append(args, "--config", "webpack.chrome.config.js")
	if debugFlag {
		args = append(args, "-d")
	} else if optimizeFlag {
		args = append(args, "-p")
	}
	runCmd("webpack", args...)
}

// DocsTask renders the docs/ with mkdocs
func DocsTask() {
	runCmd("mkdocs", "build", "--clean")
}

// BrowserTask builds browser assets for the client internal web server.
func BrowserTask() {
	var args []string
	if debugFlag {
		args = append(args, "-d")
	} else if optimizeFlag {
		args = append(args, "-p")
	}

	runCmd("webpack", args...)
}

// HotBuildTask builds according to hot reload settings and then launches the hot task.
func HotBuildTask() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		Run("browser")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		Run("chrome")
	}()
	wg.Wait()
	Run("bindata-dev")
	Run("hot")
}

// HotTask runs webpack watch and hot reloading
func HotTask() {
	go runCmd("webpack", "--config", "webpack.chrome.config.js", "--watch", "--debug")
	go runCmd("node", "server.js")
	c := newCmd("webpack", "--watch", "--hot", "--debug")
	env := os.Environ()
	env = append(env, "ALKASIR_HOT=1")
	c.Env = env
	err := c.Run()
	failErr(err, fmt.Sprintf("Failed to watch"))

}

// DevTask prepares development work
func DevTask() {
	Run("bindata-dev")
}

// BindataTask updates the list of resources to be compiled into client binary
func BindataTask() {
	runCmd("go-bindata",
		"-ignore=\\.gitignore",
		"-prefix", "res/",
		"-pkg", "res",
		"-o", "pkg/res/data.go",
		"res/...")

	runCmd("go-bindata",
		"-ignore=\\.gitignore",
		"-prefix", "browser/chrome-ext/src/",
		"-pkg", "chrome",
		"-o", "pkg/res/chrome/data.go",
		"browser/chrome-ext/src/...")
}

// BindataDevTask updates the list of resources to be compiled into client binary
func BindataDevTask() {
	runCmd("go-bindata", "-dev",
		"-ignore=\\.gitignore",
		"-prefix", "res/",
		"-pkg", "res",
		"-o", "pkg/res/data.go",
		"res/...")

	runCmd("go-bindata", "-dev",
		"-ignore=\\.gitignore",
		"-prefix", "browser/chrome-ext/src/",
		"-pkg", "chrome",
		"-o", "pkg/res/chrome/data.go",
		"browser/chrome-ext/src/...")

}

// BumpVersionPatchTask updates the patch version and commits it.
func BumpVersionPatchTask() {
	runCmd("bumpversion", "release")
	runCmd("bumpversion", "--no-tag", "patch")
}

// CleanTask removes build data
func CleanTask() {
	var wg sync.WaitGroup
	remove := []string{
		"pkg/res/data.go",
		"res/generated/bundle.js",
		"res/generated/style.css",
		"browser/chrome-ext/src/src.zip",
	}
	removeAll := []string{
		"dist/",
		"dist-archives/",
		"site/",
		"build/",
		"res/generated/",
		"res/messages/_ref",
		"browser/chrome-ext/src/javascripts",
		"AlkasirChromeExtension/",
	}
	wg.Add(len(remove))
	wg.Add(len(removeAll))
	for _, v := range remove {
		go func(f string) {
			defer wg.Done()
			os.Remove(f)
		}(v)
	}
	for _, v := range removeAll {
		go func(f string) {
			defer wg.Done()
			os.RemoveAll(f)
		}(v)
	}
	wg.Wait()
}

// DistTask compiles binaries and browser assets.
func DistBuildTask() {
	Run("clean")
	Run("browser")
	Run("chrome")
	Run("bindata")
	Run("dist-build-go")
}

// BuildTask compiles binaries and browser assets.
func BuildTask() {
	Run("dist-build-go")
}

// DistBuildTask build all available go binaries
func DistBuildGoTask() {
	b := getGoBuilders()
	b.RunAll()
}

func TranslationsFixupTask() {
	for _, v := range []string{"fa", "sv", "zh", "ar"} {
		runCmd("go", "run",
			"cmd/alkasir-translator-tool/alkasir-translator-tool.go",
			"-basedir=res",
			"-reflang=en",
			fmt.Sprintf("-lang=%s", v),
			"-add_empty=true",
			"keys")

	}

	runCmd("go", "run",
		"cmd/alkasir-translator-tool/alkasir-translator-tool.go",
		"-basedir=res",
		"-lang=en",
		"format")
}

func CITask() {
	Run("clean")
	runCmd("npm", "install", "--silent")
	Run("browser")
	Run("bindata")
	runCmd("go", "test", "-tags=\"net\"", "github.com/alkasir/alkasir/pkg/...")
	runCmd("go", "test", "-tags=\"net\"", "github.com/alkasir/alkasir/cmd/...")
	Run("govet")
	runCmd("./pkg/pac/make.sh")
	Run("chrome")
	runCmd("jsxhint", "browser/")
	runCmd("coffeelint", "-f", "browser/.coffeelint", "browser/")
	runCmd("eslint", "-c", "browser/.eslintrc", "browser/script/")

	releaseTag := os.Getenv("ALKASIR_RELEASE_TAG")
	if releaseTag != "" {
		Run("release")
	}

	// buildRelease := false
	// branch := os.Getenv("DRONE_BRANCH")
	// if branch == "release-snapshot" {
	// 	buildRelease = true
	// } else if strings.HasPrefix(branch, "refs/tags/") {
	// 	trimmed := strings.TrimPrefix(branch, "refs/tags/")
	// 	if len(trimmed) > 0 {
	// 		_, err := strconv.Atoi(string(trimmed[0]))
	// 		if err == nil {
	// 			buildRelease = true
	// 		}
	// 	}
	// }
	// if buildRelease {
	// Run("release")
	// }
}

// GenMakefileTask creates a simple Makefile. Using make.go directly is more
// powerful.
func GenMakefileTask() {
	const header = `
GOMAINS = make.go

%.bin %.go: $(GOMAINS)
	go build -o $@ $<

default: all-offline

all-offline:
	go run make.go all --offline
`

	file, err := os.Create("Makefile")
	if err != nil {
		panic("could not create Makefile")
	}
	defer file.Close()
	_, err = file.WriteString(header)
	if err != nil {
		panic("Could not write to Makefile")
	}
	w := bufio.NewWriter(file)
	keys := []string{}
	for k := range tasks {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, t := range keys {
		fmt.Fprintln(w, ".PHONY: "+t)
		fmt.Fprintln(w, t+": make.bin")
		fmt.Fprintln(w, "\t./make.bin", t)
	}
	err = w.Flush()
	if err != nil {
		panic("could not flush write buffer")
	}
}
