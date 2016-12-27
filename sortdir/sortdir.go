package sortdir

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// UserInfo stores some basic info about the user
type UserInfo struct {
	Home     string
	Name     string
	Username string
	Uid      string
	Gid      string
}

// User is just a shorthand for the UserInfo struct
// probably completely useless and could (should?) be taken out
var User = &UserInfo{}

// Directory is the value provided by the user's input, or, by default the current working directory
var Directory string
var ByExt bool

// RunAsCMD lets you run sortdir from within another go program
func RunAsCMD() {
	flag.StringVar(&Directory, "dir", "", "The directory you want to sort. Ex: sortdir -dir=\"/my/folder\"")
	flag.BoolVar(&ByExt, "e", false, "sortdir -e")

	flag.Parse()
	if Directory == "" {
		Directory = "."
	}

	User.Setup()

	// Before anything else, perform a security check
	Safeguard(Directory)

	MoveTo(Directory)

	ls := Ls(Directory, false)

	if ByExt {
		SortByExtension(ls)
	} else {
		SortByTypes(ls)
	}
}

// Setup grabs some basic info about the current user.
// I could do without it, but it's convenient and allowed me
// to learn and remember parts of the stardard library
func (u *UserInfo) Setup() *UserInfo {
	usr, err := user.Current()
	Check(err)

	u.Home = usr.HomeDir
	u.Name = usr.Name
	u.Username = usr.Username
	u.Uid = usr.Uid
	u.Gid = usr.Gid

	return u
}

// Safeguard provides basic security by preventing sortdir from running
// on the user's home directory.
func Safeguard(dir string) {
	if dir == "." {
		pwd := Pwd()
		if pwd == User.Home {
			log.Fatalln("sortdir must not be used directly on the home directory")
		}
	} else if dir == User.Home {
		log.Fatalln("sortdir must not be used directly on the home directory")
	}
}

// GoHome moves back into the user's home folder
func GoHome(u *UserInfo) string {
	err := os.Chdir(u.Home)
	Check(err)
	wd, err := os.Getwd()
	Check(err)
	return wd
}

// MoveTo is a convenience wrapper over os.Chdir()
func MoveTo(dir string) {
	err := os.Chdir(dir)
	Check(err)
}

// Pwd is a convenience wrapper over os.Getwd()
func Pwd() string {
	wd, err := os.Getwd()
	Check(err)
	return wd
}

// Ls is a convenience wrapper for readdir that just lists the
// contents of a directory
func Ls(dir string, dots bool) []string {
	var listing []string
	files, err := ioutil.ReadDir(dir)
	Check(err)
	for _, file := range files {
		// Logic abstracted from here
		f := shouldDisplay(dots, file.Name())
		if f != "" {
			listing = append(listing, f)
		}
	}
	// Make sure we dont have any empty strings in the slice
	listing = TrimEmpty(listing)
	return listing
}

// ShouldDisplay tells me wether I should display the file or not
// based on wether it is hidden or not etc...
func shouldDisplay(dots bool, fileName string) string {
	var file string
	if dots == true {
		file = fileName
	} else if dots == false {
		if IsHiddenFile(fileName) == false {
			file = fileName
		}
	}
	return file
}

// TrimEmpty removes the empty elements from the slice of strings
func TrimEmpty(ls []string) []string {
	var trimed []string
	for _, r := range ls {
		if r != "" {
			trimed = append(trimed, r)
		}
	}
	return trimed
}

// IsHiddenFile determines wether we're dealing with a hidden file or not
func IsHiddenFile(str string) bool {
	split := strings.Index(str, ".")
	if split == 0 {
		return true
	} else {
		return false
	}
}

// SortByExtension creates the necessary folders then moves the files into
// the folders to which they correspond
func SortByExtension(ls []string) {
	extensions, files := extractFilesAndExtenstions(ls)
	// Remove empty vals
	extensions = TrimEmpty(extensions)
	files = TrimEmpty(files)

	// Create the folders
	folders := CreateFolders(extensions)

	// Move the files into their corresponding folders
	MoveFilesTo(files, folders)
}
func extractFilesAndExtenstions(ls []string) ([]string, []string) {
	var extensions []string
	var files []string
	for _, l := range ls {

		// This basic check doesn't do everything
		// TrimEmpty is still necessary at the end
		if l != "" {
			extensions = append(extensions, filepath.Ext(l))
			files = append(files, l)
		}
	}
	return extensions, files
}

func SortByTypes(ls []string) {
	_, files := extractFilesAndExtenstions(ls)
	files = TrimEmpty(files)

	mapExtensions(files)
}

func mapExtensions(files []string) {
	for _, f := range files {
		e := strings.ToLower(strings.TrimPrefix(filepath.Ext(f), "."))

		_, ok := DefaultFilesMapping[e]
		if ok {
			_ = os.MkdirAll(DefaultFilesMapping[e], 0755)
			os.Rename(f, filepath.Join(DefaultFilesMapping[e], f))

		}
	}
}

// MoveFilesTo moves the files into their own directory
// determined by assortToFolder
func MoveFilesTo(files []string, folders []string) {
	for _, file := range files {
		assortToFolder(file, folders)
	}
}

// Determine which folder for which extension.
func assortToFolder(file string, folders []string) {
	for _, folder := range folders {
		tmpfile := filepath.Ext(file)
		tmpfile = strings.TrimPrefix(tmpfile, ".")
		if tmpfile == folder {
			os.Rename(file, folder+"/"+file)
		}
	}
}

// CreateFolders makes the folders for each extension
// TODO: find a way to exclude certain file extensions like .DS_Store from
// creating their own folders
// SORTA fixed by setting Ls dots to false but not really.
func CreateFolders(extensions []string) []string {
	var folders []string
	for _, ex := range extensions {
		ex = strings.TrimPrefix(ex, ".")
		folderName := ex
		os.MkdirAll(folderName, 0755)
		folders = append(folders, folderName)
	}
	return folders
}

// Check is a simple wrapper that logs errors to the console
func Check(err error) {
	if err != nil {
		log.Println(err)
	}
}
