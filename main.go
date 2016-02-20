package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

type UserInfo struct {
	Home     string
	Name     string
	Username string
	Uid      string
	Gid      string
}

var User = &UserInfo{}
var Directory string

func main() {
	flag.StringVar(&Directory, "dir", "", "The directory you want to sort. Ex: sortdir -dir=\"/my/folder\"")
	flag.Parse()
	if Directory == "" {
		Directory = "."
	}
	User.Setup()

	// Before anything else, perform a security check
	Safeguard(Directory)

	MoveTo(Directory)

	ls := Ls(Directory, false)

	SortByExtension(ls)

}

// InitUser sets some basic info about the current user
func (u *UserInfo) Setup() *UserInfo {
	user, err := user.Current()
	Check(err)

	u.Home = user.HomeDir
	u.Name = user.Name
	u.Username = user.Username
	u.Uid = user.Uid
	u.Gid = user.Gid

	return u
}

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

func GoHome(u *UserInfo) string {
	err := os.Chdir(u.Home)
	Check(err)
	wd, err := os.Getwd()
	Check(err)
	return wd
}

func MoveTo(dir string) {
	err := os.Chdir(dir)
	Check(err)
}

func Pwd() string {
	wd, err := os.Getwd()
	Check(err)
	return wd
}

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

func TrimEmpty(ls []string) []string {
	var trimed []string
	for _, r := range ls {
		if r != "" {
			trimed = append(trimed, r)
		}
	}
	return trimed
}

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
	// Remove empty vals
	extensions = TrimEmpty(extensions)
	files = TrimEmpty(files)

	// Create the folders
	folders := CreateFolders(extensions)

	// Move the files into their corresponding folders
	MoveFilesTo(files, folders)

	log.Println("Done !")
}

func UpperFirst(s string) string {
	if s == "" {
		return ""
	}

	r, n := utf8.DecodeRuneInString(s)

	return string(unicode.ToUpper(r)) + s[n:]
}

func MoveFilesTo(files []string, folders []string) {
	for _, file := range files {
		assortToFolder(file, folders)
	}
}

func assortToFolder(file string, folders []string) {
	for _, folder := range folders {
		tmpfile := filepath.Ext(file)
		tmpfile = strings.TrimPrefix(tmpfile, ".")
		tmpfile = UpperFirst(tmpfile)
		tmpfile = tmpfile + " Files"
		tmpfile = UpperFirst(tmpfile)
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
		ex = UpperFirst(ex)
		folderName := ex + " Files"
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
