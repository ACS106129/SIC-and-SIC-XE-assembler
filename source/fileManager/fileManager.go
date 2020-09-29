package filemanager

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	sourceUtil "../util"
)

// Record struct
type Record struct {
	Loc       int
	Source    string
	Statement string
	Parameter []string
	Objcode   string
	Comment   string
}

// Block struct
type Block struct {
	BlockNum   int
	CurrentLoc int
}

var utilFunc = sourceUtil.NewUtil()

func check(errs ...error) {
	for _, e := range errs {
		if e != nil {
			panic(e)
		}
	}
}

// GetRecords return the file records
func GetRecords(sourceFilePath string) []Record {
	file, err := os.Open(sourceFilePath)
	check(err)
	records := []Record{}
	for i, line2_1 := 0, scanStrings(file); i < len(line2_1); i++ {
		strWithoutSpace := strings.Fields(strings.TrimSpace(line2_1[i]))
		dotIndex := strings.Index(line2_1[i], ".")
		if dotIndex != -1 {
			// if there is instruction before comment
			if matched, err := regexp.MatchString("[\\S]+", line2_1[i][:dotIndex]); matched &&
				line2_1[i][:dotIndex] != "" {
				check(err)
				segment := strings.Fields(line2_1[i][:dotIndex])
				recordStringWithComment(&records, segment, string(line2_1[i][dotIndex:]))
			} else {
				// if line is all comment
				recordStringWithComment(&records, []string{}, line2_1[i])
			}
		} else {
			recordString(&records, strWithoutSpace)
		}
	}
	file.Close()
	return records
}

// IsPureComment return true if all are comment
func IsPureComment(record Record) bool {
	return record.Comment != "" && len(record.Parameter) == 0 && record.Source == "" && record.Statement == ""
}

// return string lines
func scanStrings(file *os.File) []string {
	lines := []string{}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func compute(x int, y int, op string) int {
	switch op {
	case "+":
		return x + y
	case "-":
		return x - y
	case "*":
		return x * y
	case "/":
		return x / y
	}
	panic("Wrong compute operator")
}

// return the new record index
func recordString(records *[]Record, words []string) int {
	record := Record{}
	wordLen := len(words)
	if hasState, index := utilFunc.HasStatement(words); hasState {
		switch index {
		case 0:
			record.Statement = words[index]
			break
		case 1:
			record.Statement = words[index]
			record.Source = words[index-1]
			break
		default:
			panic("Statement over the certain index!!!")
		}
		if index+1 < wordLen {
			record.Parameter = strings.Split(strings.Join(words[index+1:], ""), ",")
		}
	} else {
		record.Comment = "\r"
	}
	*records = append(*records, record)
	return len(*records) - 1
}

func recordStringWithComment(records *[]Record, words []string, comment string) {
	index := recordString(records, words)
	(*records)[index].Comment = comment
}

func getOffset(statement string, parameter []string) int {
	switch statement {
	case "RESW":
		num, err := strconv.Atoi(parameter[0])
		check(err)
		return num * 3
	case "RESB":
		num, err := strconv.Atoi(parameter[0])
		check(err)
		return num * 1
	case "BYTE":
		word := strings.Replace(parameter[0], "'", "", -1)
		if word[0] == 'X' {
			return len(word[1:]) / 2
		}
		return len(word[1:])
	default:
		if strings.Contains(statement, "=") {
			word := strings.Replace(statement, "'", "", -1)
			if word[1] == 'X' {
				return len(word[2:]) / 2
			}
			return len(word[2:])
		}
		fmt, _ := utilFunc.GetFormatAndOpcode(statement)
		return fmt
	}
}

func findRecordBySource(source string, records *[]Record) Record {
	for _, e := range *records {
		if e.Source == source {
			return e
		}
	}
	return Record{}
}

func getBlock(blocks *map[string]Block, record Record) (string, int) {
	useName := strings.Join(record.Parameter, "")
	state := record.Statement
	blank := "blank"
	if e, b := (*blocks)[useName]; b {
		return useName, e.CurrentLoc
	} else if useName != "" {
		(*blocks)[useName] = Block{len(*blocks), 0}
		return useName, 0
	} else if state == "CSECT" {
		(*blocks)[blank] = Block{0, 0}
	}
	return blank, (*blocks)[blank].CurrentLoc
}

// WriteRecordsList Write the records into the file
func WriteRecordsList(records *[]Record, resourceFilePath string) {
	resource, err := os.OpenFile(resourceFilePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	check(err)
	currentBlock := "blank"
	blocks := map[string]Block{
		currentBlock: Block{0, 0},
	}
	literalRecords := map[string]*Record{}
	fmt.Fprintf(resource, "%-11s %-8s %-10s %-20s\r\n", "Loc/Block", "Source", "Statement", "Parameter")
	for i := 0; i < len(*records); i++ {
		// scan statement
		if (*records)[i].Statement == "START" {
			temp, err := strconv.ParseInt((*records)[i].Parameter[0], 16, 64)
			check(err)
			(*records)[i].Loc = int(temp)
			blocks[currentBlock] = Block{blocks[currentBlock].BlockNum, (*records)[i].Loc}
		} else if IsPureComment((*records)[i]) {
			(*records)[i].Loc = -1
		} else {
			offset := getOffset((*records)[i].Statement, (*records)[i].Parameter)
			if offset == -1 {
				(*records)[i].Loc = -1
				if (*records)[i].Statement == "END" {
					temps := []Record{}
					for key := range literalRecords {
						literal := literalRecords[key]
						strLen := 0
						if key[1] == 'X' {
							strLen = len(key[3:len(key)-1]) / 2
						} else {
							strLen = len(key[3 : len(key)-1])
						}
						literal.Loc = blocks[currentBlock].CurrentLoc
						blocks[currentBlock] = Block{blocks[currentBlock].BlockNum, literal.Loc + strLen}
						temps = append(temps, *literal)
					}
					*records = append((*records)[:i+1], append(temps, (*records)[i+1:]...)...)
					literalRecords = map[string]*Record{}
				}
			} else if offset == -2 {
				if strings.Contains(strings.Join((*records)[i].Parameter, ""), "*") {
					(*records)[i].Loc = blocks[currentBlock].CurrentLoc
				} else if OpIndex := strings.IndexAny((*records)[i].Parameter[0], "+-*/"); OpIndex != -1 {
					op := (*records)[i].Parameter[0][OpIndex]
					sources := strings.Split((*records)[i].Parameter[0], string(op))
					result := compute(findRecordBySource(sources[0], records).Loc,
						findRecordBySource(sources[1], records).Loc, string(op))
					(*records)[i].Loc = result
				}
			} else if offset == -3 {
				currentBlock, (*records)[i].Loc = getBlock(&blocks, (*records)[i])
			} else if offset == -4 {
				temps := []Record{}
				for key := range literalRecords {
					literal := literalRecords[key]
					strLen := 0
					if key[1] == 'X' {
						strLen = len(key[3:len(key)-1]) / 2
					} else {
						strLen = len(key[3 : len(key)-1])
					}
					literal.Loc = blocks[currentBlock].CurrentLoc
					blocks[currentBlock] = Block{blocks[currentBlock].BlockNum, literal.Loc + strLen}
					temps = append(temps, *literal)
				}
				*records = append((*records)[:i+1], append(temps, (*records)[i+1:]...)...)
				literalRecords = map[string]*Record{}
				(*records)[i].Loc = -1
			} else if offset != 0 {
				if !strings.Contains((*records)[i].Statement, "=") {
					(*records)[i].Loc = blocks[currentBlock].CurrentLoc
					blocks[currentBlock] = Block{blocks[currentBlock].BlockNum, (*records)[i].Loc + offset}
				}
			}
		}
		// scan parameter
		for _, para := range (*records)[i].Parameter {
			if strings.Index(para, "=") == 0 {
				if _, b := literalRecords[para]; !b {
					temp := Record{}
					temp.Source = "*"
					temp.Statement = para
					literalRecords[para] = &temp
				}
			}
		}
		// write in file
		if IsPureComment((*records)[i]) {
			fmt.Fprintln(resource, (*records)[i].Comment)
		} else if (*records)[i].Loc == -1 {
			fmt.Fprintf(resource, "%-11s %-8s %-10s %-30s %s\r\n", "", (*records)[i].Source,
				(*records)[i].Statement, strings.Join((*records)[i].Parameter, ","), (*records)[i].Comment)
		} else if (*records)[i].Source == "MAXLEN" {
			fmt.Fprintf(resource, "%-11.4X %-8s %-10s %-30s %s\r\n", (*records)[i].Loc, (*records)[i].Source,
				(*records)[i].Statement, strings.Join((*records)[i].Parameter, ","), (*records)[i].Comment)
		} else {
			fmt.Fprintf(resource, "%-6.4X %-4X %-8s %-10s %-30s %s\r\n", (*records)[i].Loc,
				blocks[currentBlock].BlockNum, (*records)[i].Source,
				(*records)[i].Statement, strings.Join((*records)[i].Parameter, ","),
				(*records)[i].Comment)
		}
	}

	resource.Close()
}

func isHeader(statement string) bool {
	if statement == "START" || statement == "CSECT" {
		return true
	}
	return false
}

func getRegisterNum(regs []string) int {
	str := ""
	for _, reg := range regs {
		switch reg {
		case "A":
			str += "0"
			break
		case "X":
			str += "1"
			break
		case "L":
			str += "2"
			break
		case "PC":
			str += "8"
			break
		case "SW":
			str += "9"
			break
		case "B":
			str += "3"
			break
		case "S":
			str += "4"
			break
		case "T":
			str += "5"
			break
		case "F":
			str += "6"
			break
		}
	}
	if len(str) == 1 {
		str += "0"
	}
	num, err := strconv.ParseInt(str, 16, 64)
	check(err)
	return int(num)
}

func opcodeReflash(opcode int, n int, i int) string {
	binstr := []byte(fmt.Sprintf("%.8b", opcode))
	binstr[len(binstr)-2] = strconv.Itoa(n)[0]
	binstr[len(binstr)-1] = strconv.Itoa(i)[0]
	bin, err := strconv.ParseInt(string(binstr), 2, 64)
	check(err)
	return fmt.Sprintf("%.2X", bin)
}

func dispReflash(disp int, x int, b int, p int, e int) string {
	// format 3
	if e == 0 {
		binstr := []byte(fmt.Sprintf("%.16b", uint64(disp)))
		binstr[0] = strconv.Itoa(x)[0]
		binstr[1] = strconv.Itoa(b)[0]
		binstr[2] = strconv.Itoa(p)[0]
		binstr[3] = strconv.Itoa(e)[0]
		bin, err := strconv.ParseUint(string(binstr), 2, 64)
		check(err)
		str := fmt.Sprintf("%.4X", bin)
		return str[len(str)-4:]
	}
	// format 4
	binstr := []byte(fmt.Sprintf("%.24b", uint64(disp)))
	binstr[0] = strconv.Itoa(x)[0]
	binstr[1] = strconv.Itoa(b)[0]
	binstr[2] = strconv.Itoa(p)[0]
	binstr[3] = strconv.Itoa(e)[0]
	bin, err := strconv.ParseUint(string(binstr), 2, 64)
	check(err)
	str := fmt.Sprintf("%.6X", bin)
	return str[len(str)-6:]
}

func getObjCode(current Record, records *[]Record, refs []string) string {
	n, i, x, b, p, e := 0, 0, 0, 0, 0, 0
	disp := 0
	if current.Statement == "RESW" || current.Statement == "RESB" ||
		current.Statement == "LTORG" || current.Statement == "BASE" ||
		current.Statement == "END" || current.Statement == "EQU" ||
		current.Statement == "USE" || current.Statement == "" {
		return ""
	}
	format, opcode := utilFunc.GetFormatAndOpcode(current.Statement)
	if len(current.Parameter) == 0 {
		if strings.ContainsRune(current.Statement, '=') {
			str := current.Statement[1 : len(current.Statement)-1]
			if str[0] == 'X' {
				return str[2:]
			} else if str[0] == 'C' {
				temp := str[2:]
				return fmt.Sprintf("%.2X%.2X%.2X", temp[0], temp[1], temp[2])
			}
		}
		n, i = 1, 1
		return opcodeReflash(opcode, n, i) + dispReflash(disp, x, b, p, e)
	}
	if current.Parameter[0][0] == '#' {
		if matched, err := regexp.MatchString("[^0-9]+", current.Parameter[0][1:]); matched {
			check(err)
			current.Parameter[0] = strings.Replace(current.Parameter[0], "#", "", -1)
		} else {
			num, err := strconv.ParseInt(current.Parameter[0][1:], 10, 64)
			check(err)
			disp = int(num)
		}
		i = 1
	} else if current.Parameter[0][0] == '@' {
		n = 1
	} else {
		i, n = 1, 1
	}
	if format == 1 {
		str := current.Parameter[0][2 : len(current.Parameter[0])-1]
		if current.Parameter[0][0] == 'X' {
			return str
		} else if current.Parameter[0][0] == 'C' {
			return fmt.Sprintf("%.2X%.2X%.2X", str[0], str[1], str[2])
		}
	} else if format == 2 {
		str := fmt.Sprintf("%.2X%.2X", opcode, getRegisterNum(current.Parameter))
		return str
	} else if format == 3 {
		for _, ref := range refs {
			if strings.Contains(current.Parameter[0], ref) {
				if strings.ContainsAny(current.Parameter[0], "+-*/") {
					return "000000"
				}
				// if para is reference
				disp = 0
			}
		}
		if current.Statement == "WORD" {
			num, err := strconv.ParseInt(current.Parameter[0], 10, 64)
			check(err)
			return fmt.Sprintf("%.6X", num)
		}
		currentLoc := current.Loc + format
		if len(current.Parameter) > 1 && current.Parameter[1][0] == 'X' {
			x = 1
		}
		disp = findRecordBySource(current.Parameter[0], records).Loc - currentLoc
		if i == 0 {
			if disp > -2048 || disp < 2047 {
				p = 1
			} else {
				b = 1
			}
		}
	} else if format == 4 {
		e = 1
		if len(current.Parameter) > 1 && current.Parameter[1][0] == 'X' {
			x = 1
		}
		disp = findRecordBySource(current.Parameter[0], records).Loc
		for _, ref := range refs {
			if strings.Contains(current.Parameter[0], ref) {
				if strings.ContainsAny(current.Parameter[0], "+-*/") {
					return "00000000"
				}
				// if para is reference
				disp = 0
			}
		}
	}
	return opcodeReflash(opcode, n, i) + dispReflash(disp, x, b, p, e)
}

// WriteRecordsObjectCode Write the records into the file
func WriteRecordsObjectCode(records *[]Record, resourceFilePath string) {
	resource, err := os.OpenFile(resourceFilePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	symbolResource, err2 := os.OpenFile(strings.Replace(resourceFilePath, "txt", "sym", -1), os.O_RDWR|os.O_CREATE, os.ModePerm)
	check(err, err2)
	symbol := map[string]int{}
	codeLen := 0
	startLoc := 0
	currentLoc := 0
	currentState := ""
	currentSource := ""
	objCodes := ""
	lineLen := 0
	action := true
	references := []string{}
	for i := range *records {
		if isHeader((*records)[i].Statement) {
			// get the header length and start location
			if (*records)[i].Statement == "START" {
				startLoc = (*records)[i].Loc
			}
			lotalLength := 0
			codeLen = 0
			currentLoc = (*records)[i].Loc
			currentState = (*records)[i].Statement
			currentSource = (*records)[i].Source
			for _, e := range (*records)[i+1:] {
				if isHeader(e.Statement) {
					break
				}
				if offset := getOffset(e.Statement, e.Parameter); offset > 0 {
					lotalLength += offset
				}
			}
			fmt.Fprintf(resource, "%s%-6s%.6X%.6X\r\n", "H", (*records)[i].Source, (*records)[i].Loc, lotalLength)
			// dereference
		} else if (*records)[i].Statement == "EXTDEF" {
			fmt.Fprintf(resource, "%s", "D")
			for _, para := range (*records)[i].Parameter {
				fmt.Fprintf(resource, "%-6s%.6X", para, findRecordBySource(para, records).Loc)
			}
			fmt.Fprintln(resource)
			// reference
		} else if (*records)[i].Statement == "EXTREF" {
			// initialize references
			references = []string{}
			fmt.Fprintf(resource, "%s", "R")
			for _, para := range (*records)[i].Parameter {
				fmt.Fprintf(resource, "%-6s", para)
				references = append(references, para)
			}
			fmt.Fprintln(resource)
		} else {
			objCode := getObjCode((*records)[i], records, references)
			result := codeLen + len(objCode)
			nextState := ""
			if i+1 <= len(*records)-1 {
				nextState = (*records)[i+1].Statement
			}
			//fmt.Printf("%.4X %10s %s %d\r\n", (*records)[i].Loc, (*records)[i].Statement, objCode, result)
			if codeLen < 0 || result > 60 ||
				i == len(*records)-1 && codeLen > 0 ||
				nextState == "CSECT" && codeLen > 0 {
				fmt.Fprintf(resource, "%s%.6X%.2X%s\r\n", "T", currentLoc, lineLen, objCodes)
				currentLoc = (*records)[i].Loc
				objCodes = ""
				codeLen = 0
				lineLen = 0
				if result > 60 {
					codeLen += len(objCode)
					objCodes += objCode
				}
			} else if objCode != "" {
				if !action {
					currentLoc = (*records)[i].Loc
				}
				lineLen += getOffset((*records)[i].Statement, (*records)[i].Parameter)
				codeLen += len(objCode)
				objCodes += objCode
				action = true
			} else if action && !IsPureComment((*records)[i]) && (*records)[i].Statement != "BASE" {
				fmt.Fprintf(resource, "%s%.6X%.2X%s\r\n", "T", currentLoc, lineLen, objCodes)
				objCodes = ""
				codeLen = 0
				lineLen = 0
				action = false
			}
			if currentState == "CSECT" && nextState == "CSECT" {
				mod := utilFunc.GetModify(resourceFilePath, currentSource)
				if mod != "" {
					fmt.Fprintf(resource, "%s", mod)
				}
				fmt.Fprintf(resource, "E\r\n\r\n")
			} else if currentState == "CSECT" && i == len(*records)-1 {
				mod := utilFunc.GetModify(resourceFilePath, currentSource)
				if mod != "" {
					fmt.Fprintf(resource, "%s", mod)
				}
				fmt.Fprintln(resource, "E")
			} else if currentState == "START" && nextState == "CSECT" {
				mod := utilFunc.GetModify(resourceFilePath, currentSource)
				if mod != "" {
					fmt.Fprintf(resource, "%s", mod)
				}
				fmt.Fprintf(resource, "%s%.6X\r\n\r\n", "E", startLoc)
			} else if currentState == "START" && i == len(*records)-1 {
				mod := utilFunc.GetModify(resourceFilePath, currentSource)
				if mod != "" {
					fmt.Fprintf(resource, "%s", mod)
				}
				fmt.Fprintf(resource, "%s%.6X\r\n", "E", startLoc)
			}
		}
		if (*records)[i].Source != "" && (*records)[i].Source != "*" {
			symbol[(*records)[i].Source] = (*records)[i].Loc
		}
	}
	for key, value := range symbol {
		fmt.Fprintf(symbolResource, "%.4X %s\r\n", value, key)
	}
}
