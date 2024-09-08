package fs

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestMkdir(t *testing.T) {
	format()
	MkDir("/", "/source")
	results, error := LsDir("/")
	assert.Equal(t, "", error)
	assert.Equal(t, len(results), 1)
	assert.Equal(t, results[0].name, "/source")

}

func TestFormat(t *testing.T) {
	format()
	var fat = read_fat()
	assert.Equal(t, len(fat), 1)
	results, error := LsDir("/")
	assert.Equal(t, "", error)
	assert.Equal(t, len(results), 0)
}

func TestLs(t *testing.T) {
	format()
	MkDir("/", "/source")
	results, error := LsDir("/")
	assert.Equal(t, "", error)
	assert.Equal(t, len(results), 1)

	results2, error2 := LsDir("/source")
	assert.Equal(t, "", error2)
	assert.Equal(t, len(results2), 0)

}

func TestRmFile(t *testing.T) {
	format()
	MkDir("/", "/source")
	MkDir("/", "/source2")
	results, error := LsDir("/")
	assert.Equal(t, "", error)
	assert.Equal(t, len(results), 2)
	RmDir("/source")
	results2, error2 := LsDir("/")
	assert.Equal(t, "", error2)
	assert.Equal(t, len(results2), 1)

}

func TestCreateFile(t *testing.T) {
	format()
	MkDir("/", "/source")
	CreateFile("/source", "/source/a.txt", make([]byte, 3))
	results, error := LsDir("/source")
	assert.Equal(t, "", error)
	assert.Equal(t, len(results), 1)
	assert.Equal(t, results[len(results)-1].name, "/source/a.txt")

}

func TestCreateAndReadFile(t *testing.T) {
	format()
	MkDir("/", "/source")
	x := make([]byte, 3)
	x[0] = 1
	x[1] = 2
	x[2] = 3
	CreateFile("/source", "/source/a.txt", x)
	var data = ReadFile("/source/a.txt")
	assert.Equal(t, len(data), 3)
	assert.Equal(t, data[0], uint8(1))
	assert.Equal(t, data[1], uint8(2))
	assert.Equal(t, data[2], uint8(3))

}

func TestDeleteFile(t *testing.T) {
	format()
	MkDir("/", "/source")
	CreateFile("/source", "/source/a.txt", []byte("hello"))
	results, _ := LsDir("/source")
	assert.Equal(t, len(results), 1)
	DeleteFile("/source/a.txt")
	results2, _ := LsDir("/source")
	assert.Equal(t, len(results2), 0)
	//assert.Equal(t, results2[0].name, "")

}
