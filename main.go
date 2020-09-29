package main

import (
	"os"

	fileManager "./source/fileManager"
)

func main() {
	// make resource directory
	srcDir := "source/files/"
	resDir := "output/"
	os.MkdirAll(resDir, os.ModePerm)
	// load source file
	records2_1 := fileManager.GetRecords(srcDir + "Figure2.1.txt")
	records2_5 := fileManager.GetRecords(srcDir + "Figure2.5.txt")
	records2_9 := fileManager.GetRecords(srcDir + "Figure2.9.txt")
	records2_11 := fileManager.GetRecords(srcDir + "Figure2.11.txt")
	records2_15 := fileManager.GetRecords(srcDir + "Figure2.15.txt")
	// write resource file(loc file)
	fileManager.WriteRecordsList(&records2_1, resDir+"Figure2.2.txt")
	fileManager.WriteRecordsList(&records2_5, resDir+"Figure2.6.txt")
	fileManager.WriteRecordsList(&records2_9, resDir+"Figure2.10.txt")
	fileManager.WriteRecordsList(&records2_11, resDir+"Figure2.12.txt")
	fileManager.WriteRecordsList(&records2_15, resDir+"Figure2.16.txt")
	// write resource file(objcode file)
	fileManager.WriteRecordsObjectCode(&records2_1, resDir+"Figure2.3.txt")
	fileManager.WriteRecordsObjectCode(&records2_5, resDir+"Figure2.8.txt")
	fileManager.WriteRecordsObjectCode(&records2_11, resDir+"Figure2.13.txt")
	fileManager.WriteRecordsObjectCode(&records2_15, resDir+"Figure2.17.txt")
}
