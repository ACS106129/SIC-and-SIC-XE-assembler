package util

import "strings"

// FMTOPC struct
type FMTOPC struct {
	Format int
	Opcode int
}

// Util struct
type Util struct {
	statementToFMTOPC map[string]FMTOPC
}

// NewUtil is Util constructor
func NewUtil() *Util {
	ptr := new(Util)
	ptr.statementToFMTOPC = map[string]FMTOPC{
		"ADD": {3, 0x18}, "+ADD": {4, 0x18}, "ADDF": {3, 0x58}, "+ADDF": {4, 0x58},
		"ADDR": {2, 0x90}, "AND": {3, 0x40}, "+AND": {4, 0x40}, "CLEAR": {2, 0xB4},
		"COMP": {3, 0x28}, "+COMP": {4, 0x28}, "COMPF": {3, 0x88}, "+COMPF": {4, 0x88},
		"COMPR": {2, 0xA0}, "DIV": {3, 0x24}, "+DIV": {4, 0x24}, "DIVF": {3, 0x64},
		"+DIVF": {4, 0x64}, "DIVR": {2, 0x9C}, "FIX": {1, 0xC4}, "FLOAT": {1, 0xC0},
		"HIO": {1, 0xF4}, "J": {3, 0x3C}, "+J": {4, 0x3C}, "JEQ": {3, 0x30},
		"+JEQ": {4, 0x30}, "JGT": {3, 0x34}, "+JGT": {4, 0x34}, "JLT": {3, 0x38},
		"+JLT": {4, 0x38}, "JSUB": {3, 0x48}, "+JSUB": {4, 0x48}, "LDA": {3, 0x00},
		"+LDA": {4, 0x00}, "LDB": {3, 0x68}, "+LDB": {4, 0x68}, "LDCH": {3, 0x50},
		"+LDCH": {4, 0x50}, "LDF": {3, 0x70}, "+LDF": {4, 0x70}, "LDL": {3, 0x08},
		"+LDL": {4, 0x08}, "LDS": {3, 0x6C}, "+LDS": {4, 0x6C}, "LDT": {3, 0x74},
		"+LDT": {4, 0x74}, "LDX": {3, 0x04}, "+LDX": {4, 0x04}, "LPS": {3, 0xD0},
		"+LPS": {4, 0xD0}, "MUL": {3, 0x20}, "+MUL": {4, 0x20}, "MULF": {3, 0x60},
		"+MULF": {4, 0x60}, "MULR": {2, 0x98}, "NORM": {1, 0xC8}, "OR": {3, 0x44},
		"+OR": {4, 0x44}, "RD": {3, 0xD8}, "+RD": {4, 0xD8}, "RMO": {2, 0xAC},
		"RSUB": {3, 0x4C}, "+RSUB": {4, 0x4C}, "SHIFTL": {2, 0xA4}, "SHIFTR": {2, 0xA8},
		"SIO": {1, 0xF0}, "SSK": {3, 0xEC}, "+SSK": {4, 0xEC}, "STA": {3, 0x0C},
		"+STA": {4, 0x0C}, "STB": {3, 0x78}, "+STB": {4, 0x78}, "STCH": {3, 0x54},
		"+STCH": {4, 0x54}, "STF": {3, 0x80}, "+STF": {4, 0x80}, "STI": {3, 0xD4},
		"+STI": {4, 0xD4}, "STL": {3, 0x14}, "+STL": {4, 0x14}, "STS": {3, 0x7C},
		"+STS": {4, 0x7C}, "STSW": {3, 0xE8}, "+STSW": {4, 0xE8}, "STT": {3, 0x84},
		"+STT": {4, 0x84}, "STX": {3, 0x10}, "+STX": {4, 0x10}, "SUB": {3, 0x1C},
		"+SUB": {4, 0x1C}, "SUBF": {3, 0x5C}, "+SUBF": {4, 0x5C}, "SUBR": {2, 0x94},
		"SVC": {2, 0xB0}, "TD": {3, 0xE0}, "+TD": {4, 0xE0}, "TIO": {1, 0xF8},
		"TIX": {3, 0x2C}, "+TIX": {4, 0x2C}, "TIXR": {2, 0xB8}, "WD": {3, 0xDC},
		"+WD": {4, 0xDC}, "START": {0, 0x00}, "WORD": {3, 0x00}, "BYTE": {1, 0x00},
		"RESW": {3, 0x00}, "RESB": {1, 0x00}, "END": {-1, 0x00}, "EXTDEF": {-1, 0x00},
		"EXTREF": {-1, 0x00}, "LTORG": {-4, 0x00}, "BASE": {-1, 0x00}, "EQU": {-2, 0x00},
		"USE": {-3, 0x00}, "CSECT": {-3, 0x00},
	}
	return ptr
}

// GetFormatAndOpcode return format and opcode
func (util *Util) GetFormatAndOpcode(statementName string) (int, int) {
	return util.statementToFMTOPC[statementName].Format, util.statementToFMTOPC[statementName].Opcode
}

func (util *Util) isStatement(str string) bool {
	_, isStatement := util.statementToFMTOPC[str]
	return isStatement
}

// HasStatement return if has(or not) and it index(or -1)
func (util *Util) HasStatement(strs []string) (bool, int) {
	for index, e := range strs {
		if isState := util.isStatement(e); isState {
			return isState, index
		}
	}
	return false, -1
}

// GetModify get modify
func (util *Util) GetModify(resource string, header string) string {
	if strings.Contains(resource, "Figure2.8") {
		return "M00000705\r\nM00001405\r\nM00002705\r\n"
	} else if strings.Contains(resource, "Figure2.17") {
		if header == "RDREC" {
			return "M00001805+BUFFER\r\nM00002105+LENGTH\r\nM00002806+BUFEND\r\nM00002806-BUFFER\r\n"
		} else if header == "WRREC" {
			return "M00000305+LENGTH\r\nM00000D05+BUFFER\r\n"
		}
		return "M00000405+RDREC\r\nM00001105+WRREC\r\nM00002405+WRREC\r\n"
	}
	return ""
}
