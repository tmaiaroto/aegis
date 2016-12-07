// golint:ignore
// Code generated by go-bindata.
// sources:
// bindata.go
// example_aegis
// example_main
// function.go
// DO NOT EDIT!

package function

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _bindataGo = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x01\x00\x00\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00")

func bindataGoBytes() ([]byte, error) {
	return bindataRead(
		_bindataGo,
		"bindata.go",
	)
}

func bindataGo() (*asset, error) {
	bytes, err := bindataGoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "bindata.go", size: 0, mode: os.FileMode(420), modTime: time.Unix(1480907942, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _example_aegis = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x7c\x91\xcf\xce\xd3\x30\x10\xc4\xef\x79\x8a\xd1\xc7\x95\xaf\xa8\xfc\xb9\xf8\x96\x4a\x20\x21\x21\xc4\x81\x0b\xa7\x6a\xe3\x6c\x5a\x0b\xc7\xb6\xbc\x76\x43\xdf\x9e\x75\xd2\x16\x01\x82\x53\xac\xcc\xec\xec\xfe\x76\x29\x25\xd3\x01\x81\x66\x36\x78\xff\x83\xe6\xe4\x19\x3d\x9f\x9c\xa0\x4f\x49\x95\xef\xcc\xe9\x50\x9d\x1f\x3f\x38\xcf\x62\x50\x72\xe5\x8e\x16\x69\x55\x59\x7d\x31\x18\x54\x79\x66\x92\xf2\xbc\xef\x3c\xcd\xc3\x48\x4d\x1b\x59\x6c\x76\xa9\xac\x86\x3e\x80\x7f\xcb\xa6\x35\x7b\xaa\xc1\x36\xc3\xe7\xb5\x3b\x35\xe5\x78\xf3\xa9\xfa\x02\xbd\xa6\x4c\x54\x7d\xd1\x6a\xb6\xb5\x59\x91\xa3\x86\x2c\xce\x7b\x0c\x0c\x9b\x99\x0a\x8f\x98\x62\xc6\x35\xd6\x97\x18\x6a\x69\x0f\x58\x0a\x20\xbf\xd0\x55\x74\x38\x06\x41\x12\x5b\x37\x39\x8b\x18\x78\xb7\x86\xb7\x20\x6d\x9a\x83\x69\x34\x8e\x66\x63\xf6\xaf\xdf\xbc\x7d\x67\x9a\xf0\x6a\x1b\x66\xc3\x39\xde\xe7\xec\x28\xb9\x7f\x6e\xeb\xcb\xc7\xff\x51\xaf\x6a\x23\x7a\x4a\x39\x8e\x4f\x90\x42\xa7\xbf\x39\xa8\x96\x38\x53\x71\x96\xbc\xbf\xee\xf0\xed\x46\xa2\x5b\x70\x81\xb7\x1a\xc1\x99\xb3\x12\x09\x16\xd6\x62\xfd\x6e\x51\x17\xca\x8e\x06\x3d\xd1\x46\xf7\xf5\xcc\xc2\x7f\x4a\x4a\xcb\x48\x24\xa2\xad\x5c\x40\x39\x33\xf8\xc2\xa1\x60\x66\x91\xe6\x2c\xb1\x6d\x2f\xe3\xd3\xca\xfd\xb8\xcf\x16\xb9\xb5\x37\xeb\x1b\x68\x18\xf7\xf7\x7d\x23\xed\xdf\xe3\xd7\xa3\xeb\x2f\x97\x5e\x3c\x46\x83\x03\xe5\x9f\x01\x00\x00\xff\xff\xbc\x4c\xe7\x68\x77\x02\x00\x00")

func example_aegisBytes() ([]byte, error) {
	return bindataRead(
		_example_aegis,
		"example_aegis",
	)
}

func example_aegis() (*asset, error) {
	bytes, err := example_aegisBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "example_aegis", size: 631, mode: os.FileMode(420), modTime: time.Unix(1479428728, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _example_main = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xb4\x53\xed\x6e\xea\x46\x10\xfd\x1d\x3f\xc5\xc8\xbf\x4c\x45\x6d\x1a\x45\x55\x95\x7e\xa8\x94\xd0\xc4\x2d\x35\x15\x26\x89\xf2\x73\x31\x63\x7b\xd5\xf5\xae\xbb\x3b\xc6\xa0\xaa\x0f\x74\x5f\xe3\x3e\xd9\x9d\x35\x4e\x2e\xd1\xfd\x7d\x11\xc2\xf2\xce\xd9\x33\xe7\x9c\x19\x92\x04\x16\xa6\x3d\x59\x59\xd5\x04\x1f\x3f\xc0\xf5\xec\xbb\xef\x61\x6b\x1a\xf8\x4b\x48\x61\x0d\x19\xf8\x89\x4c\xf3\xab\xab\x65\x49\x3f\x14\x16\x05\xc9\x03\xc6\x85\x69\x7e\x09\x92\x84\xbf\xb0\x92\x05\x6a\x87\x7b\xe8\xf4\x1e\x2d\x50\x8d\x30\x6f\x45\xc1\x8f\xb1\x32\x85\x27\xb4\x4e\x1a\x0d\xd7\xf1\x0c\x22\x0f\x08\xc7\x52\x38\xf9\xd1\x53\x9c\x4c\x07\x8d\x38\x81\x36\x04\x9d\x43\xe6\x90\x0e\x4a\xa9\x10\xf0\x58\x60\x4b\x20\x35\x70\xc7\x56\x49\xa1\x0b\x84\x5e\x52\x3d\xf4\x19\x59\x62\xcf\xf1\x32\x72\x98\x1d\x09\x86\x0b\xbe\xd0\xf2\x5b\x79\x09\x04\x41\xa3\x68\xff\xa9\x89\xda\xdb\x24\xe9\xfb\x3e\x16\x83\xe0\xd8\xd8\x2a\x51\x67\xa8\x4b\x56\xe9\x62\x99\xe5\xcb\x6f\x59\xf4\x78\xe9\x51\x2b\x74\x0e\x2c\xfe\xdb\x49\xcb\x86\x77\x27\x10\x2d\x8b\x2a\xc4\x8e\xa5\x2a\xd1\x83\xb1\x20\x2a\x8b\x5c\xe3\xdc\x58\x45\x6f\x25\x49\x5d\x4d\xc1\x99\x92\x7a\x61\xd1\xd3\xec\xa5\x23\x2b\x77\x1d\xbd\xcb\xec\x55\x22\x3b\xbf\x04\x70\x6a\x42\x43\x38\xcf\x21\xcd\x43\xf8\x6d\x9e\xa7\xf9\xd4\x93\x3c\xa7\xdb\x87\xf5\xe3\x16\x9e\xe7\x9b\xcd\x3c\xdb\xa6\xcb\x1c\xd6\x1b\x58\xac\xb3\xbb\x74\x9b\xae\x33\x7e\xfb\x1d\xe6\xd9\x0b\xfc\x99\x66\x77\x53\x40\x4e\x8c\xfb\xe0\xb1\xb5\xde\x01\xcb\x94\x3e\x4d\xdc\x0f\xd1\xe5\x88\xef\x24\x94\xe6\x2c\xc9\xb5\x58\xc8\x52\x16\x6c\x4d\x57\x9d\xa8\x10\x2a\x73\x40\xab\xd9\x11\xb4\x68\x1b\xe9\xfc\x54\x1d\x0b\xdc\x7b\x1a\x25\x1b\x49\xbc\x1e\xfe\xe8\x0b\x5f\x71\x10\x70\xc6\xff\x78\x92\x86\xe7\x13\x04\x2c\xc0\x58\x82\x28\xb8\x0a\x2b\x56\xd7\xed\xfc\x4e\x25\xd4\x8c\x5b\x97\x08\xac\xa4\x4b\x94\x68\x76\x7b\x11\x32\x48\x23\x25\x9d\x55\x61\x30\x09\x82\xb2\xd3\xc5\x40\x13\x4d\xe0\xbf\xe0\x8a\x7b\x3f\xb0\x06\x35\x6e\x86\x80\xc7\xcd\xca\x4f\x09\x3b\x47\xd0\x0a\x3e\xda\x18\x0e\xd3\x06\x57\x76\x78\xc2\xed\xcf\x70\x26\x8e\x33\xec\xcf\xb5\xa8\x14\x4a\x6d\x6b\x06\x54\x35\x77\x18\x91\xf1\x99\x37\x0a\xef\x97\xdb\x70\x0a\x61\xc2\x3f\xd6\x18\xba\x40\xac\x78\x58\xc8\x42\x82\xff\x47\x5d\x17\x44\x51\x41\x47\xf8\x66\x6c\xb5\x30\x9a\xf0\x48\x3c\x8c\x03\xbd\x1d\x2e\x0f\xa8\xf9\x88\xa7\xf2\x76\xf4\xb7\x35\xc7\xd3\x06\x5d\x6b\x86\xff\x4e\x2b\xac\x68\x38\x50\xab\xe2\x27\xa1\x3a\x74\x83\x65\xbe\x10\xe7\x48\x39\xe7\xdd\xb9\xe8\x66\x76\xf3\xb9\xbf\xd7\xf7\xf5\x1a\x8f\xd8\x95\xa9\xe2\x54\x97\x26\x0a\x95\xa9\x2a\xbf\x10\xbc\xf0\x0b\x65\xba\xfd\xb3\xa0\xa2\x86\x83\x14\xc0\x15\xdb\xb9\x70\x32\x0c\x88\x5f\xb8\x81\xd4\xa4\x74\x14\x6a\x63\x1b\xa1\xe0\xde\xc0\xeb\xed\x48\x28\x67\x78\xbd\x58\xce\x3b\xa2\x49\x38\x44\xcd\x6e\xff\xc8\xd7\x59\x74\x3d\x9b\x0d\x36\xbc\xdb\x4f\x01\x00\x00\xff\xff\xdf\xd0\xa6\x2b\xb9\x04\x00\x00")

func example_mainBytes() ([]byte, error) {
	return bindataRead(
		_example_main,
		"example_main",
	)
}

func example_main() (*asset, error) {
	bytes, err := example_mainBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "example_main", size: 1209, mode: os.FileMode(420), modTime: time.Unix(1480907901, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _functionGo = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x5c\xcc\x31\xae\xc2\x30\x10\x45\xd1\xde\xab\x78\x1b\x48\xdc\xff\x2e\xd5\x6f\x28\x90\x28\xa8\x5f\xe2\xc9\x60\x25\x78\x46\xb6\xc3\xfa\x81\x06\x24\xda\x7b\xa5\x13\xa3\xda\x9f\x4a\x91\xca\x2e\x50\x1b\xe6\x5c\x12\x3b\x31\xf8\xa6\x58\x8f\xb2\xf4\x6c\x05\x63\x08\x31\xe2\xcc\x65\xa3\xca\x37\x7b\xb5\x47\x4e\xd2\x40\xcc\x96\x77\xa9\xbe\xbf\x99\x7f\x03\xdd\xd1\x6f\xec\x38\xda\x6b\x4f\xa2\xb9\x61\xb5\x0a\x16\x4c\xd7\x0b\x4e\xbc\xcf\x89\x1f\x68\x0c\xfe\x43\x87\x67\x00\x00\x00\xff\xff\xe4\x5b\xe2\x72\x99\x00\x00\x00")

func functionGoBytes() ([]byte, error) {
	return bindataRead(
		_functionGo,
		"function.go",
	)
}

func functionGo() (*asset, error) {
	bytes, err := functionGoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "function.go", size: 153, mode: os.FileMode(420), modTime: time.Unix(1479428516, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"bindata.go":    bindataGo,
	"example_aegis": example_aegis,
	"example_main":  example_main,
	"function.go":   functionGo,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"bindata.go":    &bintree{bindataGo, map[string]*bintree{}},
	"example_aegis": &bintree{example_aegis, map[string]*bintree{}},
	"example_main":  &bintree{example_main, map[string]*bintree{}},
	"function.go":   &bintree{functionGo, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}