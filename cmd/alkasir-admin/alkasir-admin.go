// Management commands for alkasir-central
package main

import (
	"archive/tar"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/facebookgo/flagenv"
	"github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/central/db"
	"github.com/alkasir/alkasir/pkg/debugexport"
	"github.com/alkasir/alkasir/pkg/nexus"
	"github.com/alkasir/alkasir/pkg/upgradebin"
	"github.com/alkasir/alkasir/pkg/upgradebin/makepatch"
)

// Command
type Command struct {
	Name string
	Func func([]string) error // if set, it always runs before doing anything related to sub commands
	Subs Commands
	Help string
}

// Commands for subcommands
type Commands []*Command

var rootCommand = &Command{
	Subs: Commands{
		{
			Name: "download",
			Func: downloadSnapshot,
			Help: "artifact-id - Download latest snapshot from nexus",
		},
		{
			Name: "export-api",
			Subs: Commands{
				{
					Name: "insert",
					Func: insertExportAPIAuth,
					Help: "username password - set username/password for export api authentication.",
				},
			},
		},

		{
			Name: "upgrade",
			Subs: Commands{
				{
					Name: "autocreate",
					Func: createUpgradeAuto,
					Help: " - Creates binary diffs for the latest release",
				},
				{
					Name: "list",
					Func: listVersions,
					Help: " - List all versions in the nexus repository.",
				},
				{
					Name: "create",
					Func: createUpgrade,
					Help: "[-h] oldversion oldfilename  - Creates binary between any files, only used to test upgrades.",
				},
				{
					Name: "makekeys",
					Func: makeUpgradeKeys,
					Help: " - Generate a key pair for upgrade signing",
				},
				{
					Name: "dbimport",
					Func: insertUpgrades,
					Help: " - Import created upgrades into db",
				},
			},
		},
		{
			Name: "debug",
			Subs: Commands{
				{
					Name: "import",
					Func: debugImportDebug,
					Help: "[files...] ex. debugfile.txt- debugfile2.txt-...",
				},
				{
					Name: "heap",
					Func: debugPprof,
					Help: "[file] - if not speicied the latest import is used",
				},
			},
		},
	},
}

func (c *Command) PrintHelp(prefix string, depth int) {
	var msg string
	if c.Name != "" {
		prefix = prefix + " " + c.Name
		msg = prefix
	}
	if c.Help != "" {
		msg = msg + " " + c.Help
	}
	if c.Subs != nil {
		for _, sc := range c.Subs {
			sc.PrintHelp(prefix, depth+1)
		}
	} else {
		fmt.Println(msg)
	}
}

func printHelp() {
	flag.PrintDefaults()
	fmt.Println("")
	fmt.Println(" --- alkasir admin commands --  ")
	fmt.Println("")
	rootCommand.PrintHelp("alkasir-admin", 0)
	fmt.Println("")
}

// flags
var (
	pgConnFlag   string
	nWorkersFlag int
)

func init() {
	defWorkers := runtime.NumCPU() - 1
	if defWorkers < 1 {
		defWorkers = 1
	}
	if defWorkers > 6 {
		defWorkers = 6
	}
	flag.IntVar(&nWorkersFlag, "nworkers", defWorkers,
		"number of default worker goroutines for cpu or memory intensive tasks.")

	flag.StringVar(&pgConnFlag, "pgconn",
		"user=alkasir_central password=alkasir_central dbname=alkasir_central port=39558 sslmode=disable",
		"postgresql connection string")
}

func main() {
	mrand.Seed(time.Now().UnixNano())
	errors := []error{
		flag.Set("logtostderr", "true"),
		flag.Set("logcolor", "true"),
	}
	for _, err := range errors {
		if err != nil {
			panic(err)
		}
	}
	lg.SetSrcHighlight("alkasir/cmd", "alkasir/pkg")
	lg.CopyStandardLogTo("info")
	flag.Parse()
	flagenv.Prefix = "ALKASIR_"
	flagenv.Parse()
	err := commandHandler(flag.Args())
	if err != nil {
		if err == errCommandNotFound {
			fmt.Println("")
			fmt.Println("Command index:")
			fmt.Println("")
			rootCommand.PrintHelp("alkasir-admin", 0)
			fmt.Println("")
			os.Exit(1)
		}
		lg.Fatal(err)
		os.Exit(1)
	}

}

// Get returns a subcommand if it exitst
func (c *Command) Get(name string) (*Command, bool) {
	if c.Subs == nil {
		return nil, false
	}
	return c.Subs.Get(name)

}

func (c Commands) Get(name string) (*Command, bool) {
	for _, v := range c {
		if v.Name == name {
			return v, true
		}
	}
	return nil, false
}

var (
	errNoValue         = fmt.Errorf("value required")
	errCommandNotFound = fmt.Errorf("command not found")
)

// simple loop for commands / subcommands
func commandHandler(args []string) error {
	if len(args) < 1 {
		printHelp()
		os.Exit(1)
	}
	lg.Infoln(args)
	var trail []string
	node := rootCommand
	var findCmd string
	for {
		if len(args) < 1 {
			fmt.Println("")
			fmt.Printf("command '%s' not found in '%s'\n",
				findCmd, strings.Join(trail, " "))
			return errCommandNotFound
		}
		findCmd = args[0]
		trail = append(trail, findCmd)
		c, ok := node.Get(findCmd)
		if !ok {
			fmt.Println("")
			fmt.Printf("command '%s' not found in '%s'\n",
				findCmd, strings.Join(trail, " "))
			return errCommandNotFound
		}
		newArgs := args[1:]
		if c.Func != nil {
			err := c.Func(newArgs)
			if err != nil {
				switch err {
				case errNoValue:
					fmt.Println("")
					fmt.Printf("command '%s' requires additional arguments\n",
						strings.Join(trail, " "))
					return nil
				}
				fmt.Printf("error running cmd %v %v", trail, newArgs)
				return err
			}
			return nil
		}
		node = c
		args = newArgs
	}
}

var sqlDB *db.DB

// InitDBopens a connection to the database.
func OpenDB() error {
	var err error
	sqlDB, err = db.Open(pgConnFlag)
	if err != nil {
		return err
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		sqlDB = nil
		return err
	}
	lg.Infoln("Successfully connected to the database")
	return nil
}

var jobQs = []nexus.BuildQuery{
	{Cmd: "alkasir-gui", OS: "windows", Arch: "386"},
	{Cmd: "alkasir-gui", OS: "windows", Arch: "amd64"},
	{Cmd: "alkasir-gui", OS: "darwin", Arch: "amd64"},
	{Cmd: "alkasir-gui", OS: "linux", Arch: "amd64"},
	{Cmd: "alkasir-client", OS: "linux", Arch: "amd64"},
	{Cmd: "alkasir-client", OS: "windows", Arch: "amd64"},
	{Cmd: "alkasir-client", OS: "darwin", Arch: "amd64"},
	{Cmd: "alkasir-client", OS: "linux", Arch: "386"},
	{Cmd: "alkasir-client", OS: "windows", Arch: "386"},
	{Cmd: "alkasir-client", OS: "darwin", Arch: "386"},
}

func findJSONFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		if !strings.HasSuffix(f.Name(), ".json") {
			return nil
		}
		files = append(files, path)

		return nil
	})
	if err != nil {
		return []string{}, err
	}
	return files, err
}

func insertUpgrades([]string) error {
	if err := OpenDB(); err != nil {
		return err
	}

	files, err := findJSONFiles("diffs/")
	if err != nil {
		return err
	}
	var upgrades []db.UpgradeMeta

	for _, v := range files {
		lg.V(5).Infoln("reading", v)
		data, err := ioutil.ReadFile(v)
		if err != nil {
			return err
		}
		var cpr makepatch.CreatePatchResult
		err = json.Unmarshal(data, &cpr)
		if err != nil {
			return err
		}
		um, ok, err := sqlDB.GetUpgrade(db.GetUpgradeQuery{
			Artifact:        cpr.Artifact,
			Version:         cpr.NewVersion,
			AlsoUnpublished: true,
		})
		if err != nil {
			return err
		}
		if ok && um.Artifact == cpr.Artifact && um.Version == cpr.NewVersion {
			lgheader := cpr.Artifact + " " + cpr.NewVersion
			if um.ED25519Signature != cpr.ED25519Signature {
				lg.Warningf("%s signatures does not match!", lgheader)
			}
			if um.SHA256Sum != cpr.SHA256Sum {
				lg.Warningf("%s shasum does not match!", lgheader)
			}

			lg.Infof("%s is already imported, skipping", lgheader)
			continue
		}
		upgrades = append(upgrades, db.UpgradeMeta{
			Artifact:         cpr.Artifact,
			Version:          cpr.NewVersion,
			SHA256Sum:        cpr.SHA256Sum,
			ED25519Signature: cpr.ED25519Signature,
		})
	}
	{
		// NOTE: this will be removed later, a quick hack before other upgrades refactoring takes place
		uniqeUpgrades := make(map[string]db.UpgradeMeta, 0)
		for _, v := range upgrades {
			uniqeUpgrades[fmt.Sprintf("%s---%s", v.Artifact, v.Version)] = v
		}
		upgrades = upgrades[:0]
		for _, v := range uniqeUpgrades {
			upgrades = append(upgrades, v)
		}
	}

	fmt.Println(upgrades)
	err = sqlDB.InsertUpgrades(upgrades)
	if err != nil {
		lg.Errorln(err)
		return err
	}

	return nil
}

func makeUpgradeKeys([]string) error {
	priv, pub := upgradebin.GenerateKeys(rand.Reader)
	privPem, pubPem := upgradebin.EncodeKeys(priv, pub)
	fmt.Println("")
	fmt.Println(string(privPem))
	fmt.Println("")
	fmt.Println(string(pubPem))

	privFile, err1 := os.OpenFile("upgrades-private-key.pem",
		os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	defer privFile.Close()

	pubFile, err2 := os.OpenFile("upgrades-public-key.pem",
		os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)

	defer pubFile.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	_, err1 = privFile.Write(privPem)
	_, err2 = pubFile.Write(pubPem)

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}

func createUpgrade(args []string) error {
	var (
		privPemFlag string
		pubPemFlag  string
	)
	fs := flag.NewFlagSet("upgrade create", flag.ContinueOnError)
	fs.StringVar(&privPemFlag, "privpem", "upgrades-private-key.pem", "path to load private key file from")
	fs.StringVar(&pubPemFlag, "pubpem", "upgrades-public-key.pem", "path to load public key file from")
	fs.Parse(args)
	args = fs.Args()

	privPem, err := ioutil.ReadFile(privPemFlag)
	if err != nil {
		return err
	}

	pubPem, err := ioutil.ReadFile(pubPemFlag)
	if err != nil {
		return err
	}

	if len(args) != 2 {
		return errNoValue
	}

	newfile := args[0]
	oldfile := args[1]

	job := makepatch.CreatePatchJob{
		Artifact:   "UPGRADETEST",
		OldBinary:  oldfile,
		NewBinary:  newfile,
		OldVersion: "0.0.1",
		NewVersion: "0.0.2",
		PrivateKey: string(privPem),
		PublicKey:  string(pubPem),
	}
	res, err := makepatch.CreatePatch(job)
	if err != nil {
		lg.Fatalln(res)
		return err
	}
	return nil

}

func createUpgradeAuto(args []string) error {
	var (
		privPemFlag string
		pubPemFlag  string
	)
	fs := flag.NewFlagSet("upgrade create", flag.ExitOnError)
	fs.StringVar(&privPemFlag, "privpem", "upgrades-private-key.pem", "path to load private key file from")
	fs.StringVar(&pubPemFlag, "pubpem", "upgrades-public-key.pem", "path to load public key file from")
	fs.Parse(args)
	args = fs.Args()

	privPem, err := ioutil.ReadFile(privPemFlag)
	if err != nil {
		if os.IsNotExist(err) {
			lg.Errorf("%s does not exist", privPemFlag)
			return nil
		}
		return err
	}

	pubPem, err := ioutil.ReadFile(pubPemFlag)
	if err != nil {
		if os.IsNotExist(err) {
			lg.Errorf("%s does not exist", pubPemFlag)
			return nil
		}
		return err
	}

	results, err := makepatch.RunPatchesCreate(
		jobQs, string(privPem), string(pubPem), nWorkersFlag)
	if err != nil {
		panic(err)
	}
	if len(results) < 1 {
		lg.Fatalln("no patch results returned")
	}

	var allFiles []*tar.Header
	err = filepath.Walk("diffs", func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}

		allFiles = append(allFiles, &tar.Header{
			Name: path,
			Mode: 0600,
			Size: f.Size(),
		})
		return nil
	})
	if err != nil {
		lg.Fatal(err)
	}

	latestVersion := results[0].NewVersion
	filename := fmt.Sprintf("alkasir-binpatches-for-%s.tar", latestVersion)
	tarfile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	tw := tar.NewWriter(tarfile)

	for _, hdr := range allFiles {
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatalln(err)
		}
		s, err := os.Open(hdr.Name)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, s)
		if err != nil {
			lg.Fatal(err)
		}

		err = s.Close()
		if err != nil {
			lg.Fatal(err)
		}

	}

	if err := tw.Close(); err != nil {
		log.Fatalln(err)
	}

	lg.Infoln("done")
	return nil
}

func listVersions([]string) error {
	for _, v := range jobQs {
		versions, err := v.GetVersions()
		if err != nil {
			lg.Errorln(err)
			continue
		}
		for _, v := range versions {
			spew.Dump(v)
		}
	}
	return nil
}

func downloadSnapshot(args []string) error {
	if len(args) < 1 {
		fmt.Println("specifcy a cmd, alkasir-admin, alkasir-central etc.")
		return errNoValue
	}
	err := nexus.GetMasterSnapshot(args[0])
	if err != nil {
		return err
	}

	return nil
}

func insertExportAPIAuth(args []string) error {
	if err := OpenDB(); err != nil {
		return err
	}
	if len(args) != 2 {
		fmt.Println("need [username] and [password]")
		return errNoValue
	}

	creds := db.APICredentials{
		Username: args[0],
	}
	creds.SetPassword(args[1])

	if err := sqlDB.InsertExportAPICredentials(creds); err != nil {
		return err
	}

	return nil
}

func debugImportDebug(files []string) error {
	if len(files) == 0 {
		fmt.Println("need argument: files...")
		return errNoValue
	}
	for _, filename := range files {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var debuginfo debugexport.DebugResponse
		err = json.Unmarshal(data, &debuginfo)
		if err != nil {
			fmt.Printf("could not decode %s: %s \n", filename, err.Error())
			os.Exit(1)
		}
		debuginfo.WriteToDisk()
	}
	return nil
}

func debugLatestImport() string {
	files, err := ioutil.ReadDir("alkasir-debug-reports")
	if err != nil {
		lg.Fatal(err)
	}

	var dirs []string
	for _, v := range files {
		if v.IsDir() {
			dirs = append(dirs, v.Name())
		}
	}
	sort.Strings(dirs)
	return filepath.Join("alkasir-debug-reports", dirs[len(dirs)-1])
}

func debugPprof(args []string) error {
	var dir string
	if len(args) == 0 {
		dir = debugLatestImport()
	} else {
		dir = args[0]
	}

	const profile = "heap"

	var header debugexport.DebugHeader
	data, err := ioutil.ReadFile(filepath.Join(dir, "header.json"))
	if err != nil {
		lg.Fatal(err)
	}

	err = json.Unmarshal(data, &header)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("%+v", header)
	q := nexus.BuildQuery{
		OS:      header.OS,
		Arch:    header.Arch,
		Version: header.Version,
		Cmd:     "alkasir-gui",
	}
	cmdlocation, err := q.GetMatchingBuildBinary()
	if err != nil {
		lg.Fatal(err)
	}

	var cmdargs []string
	cmdargs = append(cmdargs,
		"tool",
		"pprof",
		cmdlocation,
		filepath.Join(dir, profile+".txt"))
	cmd := exec.Command("go", cmdargs...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()

}
