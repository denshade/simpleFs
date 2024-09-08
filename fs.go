package fs

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type FileRecord struct {
	idx          uint32
	name         string
	directoryidx uint32
	size         uint32
	filetype     byte //0 = dir, 1 = file
	offset       uint32
}

const (
	recordSize     = 49
	nameSize       = 32
	numRecords     = 1000
	FILE_TYPE_DIR  = 0
	FILE_TYPE_FILE = 1
)

func read_fat() []FileRecord {
	file, err := os.Open("FILE.DAT")
	if err != nil {
		fmt.Println("Error opening file:", err)

	}
	defer file.Close()

	records := make([]FileRecord, 0, numRecords)

	buf := make([]byte, recordSize)

	for i := 0; i < numRecords; i++ {
		_, err := file.Read(buf)
		if err != nil {
			fmt.Println("Error reading file:", err)
			break
		}

		record := FileRecord{
			idx:          binary.LittleEndian.Uint32(buf[0:4]),
			name:         string(buf[4 : 4+nameSize]),
			directoryidx: binary.LittleEndian.Uint32(buf[36:40]),
			size:         binary.LittleEndian.Uint32(buf[40:44]),
			filetype:     buf[44],
			offset:       binary.LittleEndian.Uint32(buf[45:49]),
		}

		record.name = strings.Trim(string(record.name[:len(record.name)-1]), "\x00")
		if record.idx != 0 {
			records = append(records, record)
		}
	}
	return records
}

func format() {
	records := []FileRecord{
		{idx: 1, name: "/", directoryidx: 0, size: 1024, filetype: FILE_TYPE_DIR, offset: 100},
	}

	shouldReturn := writeFatRecords(records)
	if shouldReturn {
		return
	}
}

func writeFatRecords(records []FileRecord) bool {
	file, err := os.OpenFile("FILE.DAT", os.O_RDWR, 0755)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return true
	}
	defer file.Close()

	for _, record := range records {

		buf := make([]byte, recordSize)

		binary.LittleEndian.PutUint32(buf[0:4], record.idx)
		copy(buf[4:4+nameSize], []byte(padString(record.name, nameSize)))
		binary.LittleEndian.PutUint32(buf[36:40], record.directoryidx)
		binary.LittleEndian.PutUint32(buf[40:44], record.size)
		buf[44] = record.filetype
		binary.LittleEndian.PutUint32(buf[45:49], record.offset)

		_, err := file.Write(buf)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return true
		}
	}
	for i := len(records); i <= 1000; i++ {
		buf := make([]byte, recordSize)

		binary.LittleEndian.PutUint32(buf[0:4], 0)
		copy(buf[4:4+nameSize], []byte(padString("", nameSize)))
		binary.LittleEndian.PutUint32(buf[36:40], 0)
		binary.LittleEndian.PutUint32(buf[40:44], 0)
		buf[44] = 0
		binary.LittleEndian.PutUint32(buf[45:49], 0)

		_, err := file.Write(buf)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return true
		}
	}
	for i := 0; i < 1000; i++ {
		buf := make([]byte, 1000)
		_, err := file.Write(buf)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return true
		}
	}
	return false
}

func padString(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	padding := make([]byte, length-len(s))
	return s + string(padding)
}

func MkDir(source string, directory string) (error string) {
	var fat = read_fat()
	var selectedDirIdx = uint32(0)
	selectedDirIdx = getIdx(fat, source, selectedDirIdx)
	if selectedDirIdx == 0 {
		return "directory not found"
	}
	var newRecord = FileRecord{idx: uint32(len(fat) + 1), name: directory, directoryidx: selectedDirIdx, size: 0, filetype: FILE_TYPE_DIR, offset: 0}
	fat = append(fat, newRecord)
	writeFatRecords(fat)
	return ""
}

func LsDir(source string) (ret []FileRecord, error string) {
	var fat = read_fat()
	// find directory index
	var selectedDirIdx = uint32(0)
	if !checkFileExists(source) {
		return ret, "directory not found"
	}
	selectedDirIdx = getIdx(fat, source, selectedDirIdx)

	for _, s := range fat {
		if int(s.directoryidx) == int(selectedDirIdx) {
			ret = append(ret, s)
		}
	}

	return ret, ""
}

func getIdx(fat []FileRecord, source string, idx uint32) uint32 {
	record := getRecordByName(fat, source)
	if record != nil {
		return record.idx
	}
	return idx
}

func getFatPosition(fat []FileRecord, source string, idx uint32) uint32 {
	record := getRecordByName(fat, source)
	if record == nil {
		return 0
	}
	for i := 0; i <= len(fat); i++ {
		if fat[i].idx == record.idx {
			return uint32(i)
		}
	}

	return idx
}

func getRecordByName(fat []FileRecord, source string) *FileRecord {
	for _, s := range fat {
		if s.name == source {
			return &s
		}
	}
	return nil
}

func RmDir(directory string) {
	var fat = read_fat()
	var idx = uint32(0)
	idx = getIdx(fat, directory, idx)
	fat = remove(fat, int(idx))
	writeFatRecords(fat)

}

func remove(slice []FileRecord, s int) []FileRecord {
	return append(slice[:s], slice[s+1:]...)
}

func checkFileExists(source string) bool {
	var fat = read_fat()
	var selectedDirIdx = uint32(0)
	selectedDirIdx = getIdx(fat, source, selectedDirIdx)
	return selectedDirIdx != 0
}

func CreateFile(directory string, source string, data []byte) (error string) {
	if checkFileExists(source) {
		return "File already exists"
	}
	var fat = read_fat()
	var selectedDirIdx = uint32(0)
	selectedDirIdx = getIdx(fat, directory, selectedDirIdx)
	if selectedDirIdx == 0 {
		return "directory not found"
	}
	var newRecord = FileRecord{idx: uint32(len(fat) + 1), name: source, directoryidx: selectedDirIdx, size: uint32(len(data)), filetype: FILE_TYPE_DIR, offset: uint32(getNextFreeOffset(fat) + 1)}
	fat = append(fat, newRecord)
	writeFatRecords(fat)

	// find the first offset that isn't taken and can fit the file contents
	file, err := os.OpenFile("FILE.DAT", os.O_RDWR, 0755)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return "cannot open file for writing"
	}
	defer file.Close()
	print("Writing file at ", newRecord.offset)
	_, err = file.WriteAt(data, int64(newRecord.offset))
	print(err)

	return ""
}

func getNextFreeOffset(fat []FileRecord) int {
	var offset = 0
	for _, r := range fat {
		if offset < int(r.offset)+int(r.size) {
			offset = int(r.offset) + int(r.size)
		}
	}
	return offset
}

func ReadFile(source string) (data []byte) {
	var record = getRecordByName(read_fat(), source)
	if record == nil {
		print("~ cannot find file", source)
		return nil
	}

	file, err := os.Open("FILE.DAT")
	if err != nil {
		fmt.Println("Error creating file:", err)
	}

	var buffer = make([]byte, record.size)
	print("Reading file from ", record.offset)

	file.ReadAt(buffer, int64(record.offset))

	return buffer
}

func DeleteFile(source string) {
	var fat = read_fat()
	var record = getRecordByName(fat, source)
	fat = remove(fat, int(getFatPosition(fat, source, record.idx)))
	writeFatRecords(fat)
}
