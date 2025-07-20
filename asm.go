package main

import (
	"io"
	"maps"
)

// These letters appear at the start of the listing line
// when an error occurs.
const (
	illMicErr   string = "O" // Illegal micro-coded combination
	illOPRErr          = "W" // Illegal off-page reference
	multiSymErr        = "M" // Symbol is defined multiple times
	offFieldErr        = "F" // Off-field reference warning
	numOORErr          = "A" // Number out of range
	pageFullErr        = "P" // Page full
	psopErr            = "Z" // Badly formed pseudo-op
	symMalErr          = "S" // Malformed (improperly defined) symbol
	syntaxErr          = "X" // Syntax error
	undefSymErr        = "U" // Symbol undefined
)

// These are the valid values for CPU type
const (
	pdp8 = iota // Assumes memory extension instructions present
	hm6120
	im6100
)

type opcodeData struct {
	value  uint16
	optype int
}

var opcodes = map[string]opcodeData{
	/* PDP-8 memory reference instructions */
	"AND": {0o0000, opMRI}, "TAD": {0o1000, opMRI}, "ISZ": {0o2000, opMRI},
	"DCA": {0o3000, opMRI}, "JMS": {0o4000, opMRI}, "JMP": {0o5000, opMRI},

	/* PDP-8 operate instructions */
	"NOP": {0o7000, opOPR}, "IAC": {0o7001, opOPR}, "RAL": {0o7004, opOPR},
	"RTL": {0o7006, opOPR}, "RAR": {0o7010, opOPR}, "RTR": {0o7012, opOPR},
	"BSW": {0o7002, opOPR}, "CML": {0o7020, opOPR}, "CMA": {0o7040, opOPR},
	"CIA": {0o7041, opOPR}, "CLL": {0o7100, opOPR}, "STL": {0o7120, opOPR},
	"CLA": {0o7200, opOPR}, "GLK": {0o7204, opOPR}, "STA": {0o7240, opOPR},
	"HLT": {0o7402, opOPR}, "OSR": {0o7404, opOPR}, "SKP": {0o7410, opOPR},
	"SNL": {0o7420, opOPR}, "SZL": {0o7430, opOPR}, "SZA": {0o7440, opOPR},
	"SNA": {0o7450, opOPR}, "SMA": {0o7500, opOPR}, "SPA": {0o7510, opOPR},
	"LAS": {0o7604, opOPR}, "MQL": {0o7421, opOPR}, "MQA": {0o7501, opOPR},
	"SWP": {0o7521, opOPR}, "CAM": {0o7621, opOPR}, "ACL": {0o7701, opOPR},

	/* Standard PDP-8 memory extension instructions */
	"CDF": {0o6201, opCXF}, "CIF": {0o6202, opCXF}, "CXF": {0o6203, opCXF},
	"RDF": {0o6214, opIOT}, "RIF": {0o6224, opIOT}, "RIB": {0o6234, opIOT},
	"RMF": {0o6244, opIOT},

	/* Standard PDP-8 processor IOT instructions */
	"SKON": {0o6000, opIOT}, "ION": {0o6001, opIOT}, "IOF": {0o6002, opIOT},
	"SRQ": {0o6003, opIOT}, "GTF": {0o6004, opIOT}, "RTF": {0o6005, opIOT},
	"SGT": {0o6006, opIOT}, "CAF": {0o6007, opIOT},
}

type pseudoOpData struct {
	fn         func(a *assembler, line, label istring, paramCount int) error
	paramCheck countPredicate
}

var pseudoOps = map[string]pseudoOpData{
	".END":    {fn: (*assembler).dotEnd, paramCheck: none},
	".ORG":    {fn: (*assembler).dotOrg, paramCheck: exactly(1)},
	".DATA":   {fn: (*assembler).dotData, paramCheck: any},
	".TITLE":  {fn: (*assembler).dotTitle, paramCheck: any},
	".ASCIZ":  {fn: (*assembler).dotAsciz, paramCheck: exactly(1)},
	".TEXT":   {fn: (*assembler).dotText, paramCheck: exactly(1)},
	".BLOCK":  {fn: (*assembler).dotBlock, paramCheck: atLeast(1)},
	".SIXBIT": {fn: (*assembler).dotSixBit, paramCheck: exactly(1)},
	".OPDEF":  {fn: (*assembler).dotOpDef, paramCheck: exactly(1)},
	".PAGE":   {fn: (*assembler).dotPage, paramCheck: atMost(1)},
	".FIELD":  {fn: (*assembler).dotField, paramCheck: exactly(1)},
	".IM6100": {fn: (*assembler).dotIM6100, paramCheck: none},
	".HM6120": {fn: (*assembler).dotHM6120, paramCheck: none},
	".VECTOR": {fn: (*assembler).dotVector, paramCheck: exactly(1)},
	".STACK":  {fn: (*assembler).dotStack, paramCheck: exactly(4)},
	".PUSH":   {fn: (*assembler).dotPush, paramCheck: exactly(1)},
	".POP":    {fn: (*assembler).dotPop, paramCheck: none},
	".PUSHJ":  {fn: (*assembler).dotPushj, paramCheck: exactly(1)},
	".POPJ":   {fn: (*assembler).dotPopj, paramCheck: none},
	".NOWARN": {fn: (*assembler).dotNoWarn, paramCheck: atLeast(1)},
}

type asmerror struct {
	lineNumber int
	errorStr   string
}

type assembler struct {
	cpu         int
	origin      int
	pc          int
	code        []byte
	r           io.Reader
	scopeSymbol *istring
	userSymbols map[string]int
	constants   map[string]*expr
	out         io.Writer
	verbose     bool
	errors      []asmerror
}

func (asm *assembler) harrisMnemonics() {
	harris := map[string]opcodeData{
		/* HM6120 "Extra" instructions */
		"R3L": {0o7014, opOPR}, "WSR": {0o6246, opIOT}, "GCF": {0o6256, opIOT},
		"PR0": {0o6206, opIOT}, "PR1": {0o6216, opIOT}, "PR2": {0o6226, opIOT},
		"PR3": {0o6236, opIOT}, "PRS": {0o6000, opIOT}, "PGO": {0o6003, opIOT},
		"PEX": {0o6004, opIOT}, "CPD": {0o6266, opIOT}, "SPD": {0o6276, opIOT},

		/* HM6120 stack instructions */
		"PPC1": {0o6205, opIOT}, "PPC2": {0o6245, opIOT}, "PAC1": {0o6215, opIOT},
		"PAC2": {0o6255, opIOT}, "RTN1": {0o6225, opIOT}, "RTN2": {0o6265, opIOT},
		"POP1": {0o6235, opIOT}, "POP2": {0o6275, opIOT}, "RSP1": {0o6207, opIOT},
		"RSP2": {0o6227, opIOT}, "LSP1": {0o6217, opIOT}, "LSP2": {0o6237, opIOT},
	}
	maps.Insert(opcodes, maps.All(harris))
}

func (asm *assembler) intersilMnemonics() {
	intersil := map[string]opcodeData{
		/* IM6101 Peripheral Interface Element (PIE) instructions */
		"READ1": {0o6000, opPIE}, "READ2": {0o6010, opPIE}, "WRITE1": {0o6001, opPIE},
		"WRITE2": {0o6011, opPIE}, "SKIP1": {0o6002, opPIE}, "SKIP2": {0o6003, opPIE},
		"SKIP3": {0o6012, opPIE}, "SKIP4": {0o6013, opPIE}, "RCRA": {0o6004, opPIE},
		"WCRA": {0o6005, opPIE}, "WCRB": {0o6015, opPIE}, "WVR": {0o6014, opPIE},
		"SFLAG1": {0o6006, opPIE}, "SFLAG3": {0o6016, opPIE}, "CFLAG1": {0o6007, opPIE},
		"CFLAG3": {0o6017, opPIE},

		/* IM6103 Parallel I/O (PIO) instructions */
		"SETPA": {0o6300, opPIO}, "CLRPA": {0o6301, opPIO}, "WPA": {0o6302, opPIO},
		"RPA": {0o6303, opPIO}, "SETPB": {0o6304, opPIO}, "CLRPB": {0o6305, opPIO},
		"WPB": {0o6306, opPIO}, "RPB": {0o6307, opPIO}, "SETPC": {0o6310, opPIO},
		"CLRPC": {0o6311, opPIO}, "WPC": {0o6312, opPIO}, "RPC": {0o6313, opPIO},
		"SKPOR": {0o6314, opPIO}, "SKPIR": {0o6315, opPIO}, "WSR": {0o6316, opPIO},
		"RSR": {0o6317, opPIO},

		/* IM6102 Memory Extension, DMA and clock (MEDIC) instructions */
		"LIF":  {0o6254, opIOT},
		"CLZE": {0o6130, opIOT}, "CLSK": {0o6131, opIOT}, "CLOE": {0o6132, opIOT},
		"CLAB": {0o6133, opIOT}, "CLEN": {0o6134, opIOT}, "CLSA": {0o6135, opIOT},
		"CLBA": {0o6136, opIOT}, "CLCA": {0o6137, opIOT},
		"LCAR": {0o6205, opIOT}, "RCAR": {0o6215, opIOT}, "LWCR": {0o6225, opIOT},
		"LEAR": {0o6206, opCXF}, "REAR": {0o6235, opIOT}, "LFSR": {0o6245, opIOT},
		"RFSR": {0o6255, opIOT}, "WRVR": {0o6275, opIOT}, "SKOF": {0o6265, opIOT},
	}
	maps.Insert(opcodes, maps.All(intersil))
}
