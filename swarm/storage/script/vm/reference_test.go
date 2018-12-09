package txscript

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
)

// parseExpectedResult parses the provided expected result string into allowed
// script error codes.  An error is returned if the expected result string is
// not supported.
func parseExpectedResult(expected string) ([]ErrorCode, error) {
	switch expected {
	case "OK":
		return nil, nil
	case "UNKNOWN_ERROR":
		return []ErrorCode{ErrNumberTooBig, ErrMinimalData}, nil
	case "PUBKEYTYPE":
		return []ErrorCode{ErrPubKeyType}, nil
	case "SIG_DER":
		return []ErrorCode{ErrSigTooShort, ErrSigTooLong,
			ErrSigInvalidSeqID, ErrSigInvalidDataLen, ErrSigMissingSTypeID,
			ErrSigMissingSLen, ErrSigInvalidSLen,
			ErrSigInvalidRIntID, ErrSigZeroRLen, ErrSigNegativeR,
			ErrSigTooMuchRPadding, ErrSigInvalidSIntID,
			ErrSigZeroSLen, ErrSigNegativeS, ErrSigTooMuchSPadding,
			ErrInvalidSigHashType}, nil
	case "EVAL_FALSE":
		return []ErrorCode{ErrEvalFalse, ErrEmptyStack}, nil
	case "EQUALVERIFY":
		return []ErrorCode{ErrEqualVerify}, nil
	case "NULLFAIL":
		return []ErrorCode{ErrNullFail}, nil
	case "SIG_HIGH_S":
		return []ErrorCode{ErrSigHighS}, nil
	case "SIG_HASHTYPE":
		return []ErrorCode{ErrInvalidSigHashType}, nil
	case "SIG_NULLDUMMY":
		return []ErrorCode{ErrSigNullDummy}, nil
	case "SIG_PUSHONLY":
		return []ErrorCode{ErrNotPushOnly}, nil
	case "CLEANSTACK":
		return []ErrorCode{ErrCleanStack}, nil
	case "BAD_OPCODE":
		return []ErrorCode{ErrReservedOpcode, ErrMalformedPush}, nil
	case "UNBALANCED_CONDITIONAL":
		return []ErrorCode{ErrUnbalancedConditional,
			ErrInvalidStackOperation}, nil
	case "OP_RETURN":
		return []ErrorCode{ErrEarlyReturn}, nil
	case "VERIFY":
		return []ErrorCode{ErrVerify}, nil
	case "INVALID_STACK_OPERATION", "INVALID_ALTSTACK_OPERATION":
		return []ErrorCode{ErrInvalidStackOperation}, nil
	case "DISABLED_OPCODE":
		return []ErrorCode{ErrDisabledOpcode}, nil
	case "DISCOURAGE_UPGRADABLE_NOPS":
		return []ErrorCode{ErrDiscourageUpgradableNOPs}, nil
	case "PUSH_SIZE":
		return []ErrorCode{ErrElementTooBig}, nil
	case "OP_COUNT":
		return []ErrorCode{ErrTooManyOperations}, nil
	case "STACK_SIZE":
		return []ErrorCode{ErrStackOverflow}, nil
	case "SCRIPT_SIZE":
		return []ErrorCode{ErrScriptTooBig}, nil
	case "PUBKEY_COUNT":
		return []ErrorCode{ErrInvalidPubKeyCount}, nil
	case "SIG_COUNT":
		return []ErrorCode{ErrInvalidSignatureCount}, nil
	case "MINIMALDATA":
		return []ErrorCode{ErrMinimalData}, nil
	case "NEGATIVE_LOCKTIME":
		return []ErrorCode{ErrNegativeLockTime}, nil
	case "UNSATISFIED_LOCKTIME":
		return []ErrorCode{ErrUnsatisfiedLockTime}, nil
	case "MINIMALIF":
		return []ErrorCode{ErrMinimalIf}, nil
	case "DISCOURAGE_UPGRADABLE_WITNESS_PROGRAM":
		return []ErrorCode{ErrDiscourageUpgradableWitnessProgram}, nil
	case "WITNESS_PROGRAM_WRONG_LENGTH":
		return []ErrorCode{ErrWitnessProgramWrongLength}, nil
	case "WITNESS_PROGRAM_WITNESS_EMPTY":
		return []ErrorCode{ErrWitnessProgramEmpty}, nil
	case "WITNESS_PROGRAM_MISMATCH":
		return []ErrorCode{ErrWitnessProgramMismatch}, nil
	case "WITNESS_MALLEATED":
		return []ErrorCode{ErrWitnessMalleated}, nil
	case "WITNESS_MALLEATED_P2SH":
		return []ErrorCode{ErrWitnessMalleatedP2SH}, nil
	case "WITNESS_UNEXPECTED":
		return []ErrorCode{ErrWitnessUnexpected}, nil
	case "WITNESS_PUBKEYTYPE":
		return []ErrorCode{ErrWitnessPubKeyType}, nil
	}

	return nil, fmt.Errorf("unrecognized expected result in test data: %v",
		expected)
}

// parseScriptFlags parses the provided flags string from the format used in the
// reference tests into ScriptFlags suitable for use in the script engine.
func parseScriptFlags(flagStr string) (ScriptFlags, error) {
	var flags ScriptFlags

	sFlags := strings.Split(flagStr, ",")
	for _, flag := range sFlags {
		switch flag {
		case "":
			// Nothing.
		case "CHECKLOCKTIMEVERIFY":
			flags |= ScriptVerifyCheckLockTimeVerify
		case "CHECKSEQUENCEVERIFY":
			flags |= ScriptVerifyCheckSequenceVerify
		case "CLEANSTACK":
			flags |= ScriptVerifyCleanStack
		case "DERSIG":
			flags |= ScriptVerifyDERSignatures
		case "DISCOURAGE_UPGRADABLE_NOPS":
			flags |= ScriptDiscourageUpgradableNops
		case "LOW_S":
			flags |= ScriptVerifyLowS
		case "MINIMALDATA":
			flags |= ScriptVerifyMinimalData
		case "NONE":
			// Nothing.
		case "NULLDUMMY":
			flags |= ScriptStrictMultiSig
		case "NULLFAIL":
			flags |= ScriptVerifyNullFail
		case "P2SH":
			flags |= ScriptBip16
		case "SIGPUSHONLY":
			flags |= ScriptVerifySigPushOnly
		case "STRICTENC":
			flags |= ScriptVerifyStrictEncoding
		case "WITNESS":
			flags |= ScriptVerifyWitness
		case "DISCOURAGE_UPGRADABLE_WITNESS_PROGRAM":
			flags |= ScriptVerifyDiscourageUpgradeableWitnessProgram
		case "MINIMALIF":
			flags |= ScriptVerifyMinimalIf
		case "WITNESS_PUBKEYTYPE":
			flags |= ScriptVerifyWitnessPubKeyType
		default:
			return flags, fmt.Errorf("invalid flag: %s", flag)
		}
	}
	return flags, nil
}

// testScripts ensures all of the passed script tests execute with the expected
// results with or without using a signature cache, as specified by the
// parameter.
func testScripts(t *testing.T, tests [][]interface{}) {

	blackList := make(map[int]bool)

	for i, test := range tests {
		// "Format is: [[wit..., amount]?, scriptSig, scriptPubKey,
		//    flags, expected_scripterror, ... comments]"

		// Skip single line comments.
		if len(test) == 1 {
			continue
		}

		// Construct a name for the test based on the comment and test
		// data.
		name := fmt.Sprintf("Test #%d", i)

		// When the first field of the test data is a slice it contains
		// witness data and everything else is offset by 1 as a result.
		witnessOffset := 0
		if _, ok := test[0].([]interface{}); ok {
			witnessOffset++

		}

		// Extract and parse the signature script from the test fields.
		scriptSigStr, ok := test[witnessOffset].(string)
		if !ok {
			t.Errorf("%s: signature script is not a string", name)
			blackList[i] = true
			continue
		}
		scriptSig, err := parseShortForm(scriptSigStr)
		if err != nil {
			t.Errorf("%s: can't parse signature script: %v", name,
				err)
			blackList[i] = true

			continue
		}

		// Extract and parse the public key script from the test fields.
		scriptPubKeyStr, ok := test[witnessOffset+1].(string)
		if !ok {
			t.Errorf("%s: public key script is not a string", name)
			blackList[i] = true

			continue
		}
		scriptPubKey, err := parseShortForm(scriptPubKeyStr)
		if err != nil {
			t.Errorf("%s: can't parse public key script: %v", name,
				err)
			blackList[i] = true

			continue
		}

		// Extract and parse the script flags from the test fields.
		flagsStr, ok := test[witnessOffset+2].(string)
		if !ok {
			t.Errorf("%s: flags field is not a string", name)
			blackList[i] = true

			continue
		}
		flags, err := parseScriptFlags(flagsStr)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			blackList[i] = true

			continue
		}

		// Extract and parse the expected result from the test fields.
		//
		// Convert the expected result string into the allowed script
		// error codes.  This is necessary because txscript is more
		// fine grained with its errors than the reference test data, so
		// some of the reference test data errors map to more than one
		// possibility.
		resultStr, ok := test[witnessOffset+3].(string)
		if !ok {
			t.Errorf("%s: result field is not a string", name)
			blackList[i] = true

			continue
		}
		allowedErrorCodes, err := parseExpectedResult(resultStr)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			blackList[i] = true

			continue
		}

		testPayload := []byte("PAYLOAD")
		vm, err := NewEngine(scriptPubKey, scriptSig, testPayload, flags)
		if err == nil {
			err = vm.Execute()
		}

		// Ensure there were no errors when the expected result is OK.
		if resultStr == "OK" {
			if err != nil {
				t.Errorf("%s failed to execute: %v", name, err)
				blackList[i] = true
			}

			continue
		}

		// At this point an error was expected so ensure the result of
		// the execution matches it.
		success := false
		for _, code := range allowedErrorCodes {
			if IsErrorCode(err, code) {
				success = true
				break
			}
		}
		if !success {
			if serr, ok := err.(Error); ok {
				t.Errorf("%s: want error codes %v, got %v", name,
					allowedErrorCodes, serr.ErrorCode)
				blackList[i] = true

				continue
			}
			t.Errorf("%s: want error codes %v, got err: %v (%T)",
				name, allowedErrorCodes, err, err)
			blackList[i] = true

			continue
		}
	}
	/*
		var good [][]interface{}

		for i, test := range tests {
			if !blackList[i] {
				good = append(good, test)
			}
		}
		bytes, _ := json.MarshalIndent(good, "", "\t")
		ioutil.WriteFile("data/script_tests2.json", bytes, 0666)
	*/
}

// TestScripts ensures all of the tests in script_tests.json execute with the
// expected results as defined in the test data.
func TestScripts(t *testing.T) {
	file, err := ioutil.ReadFile("data/script_tests.json")
	if err != nil {
		t.Fatalf("TestScripts: %v\n", err)
	}

	var tests [][]interface{}
	err = json.Unmarshal(file, &tests)
	if err != nil {
		t.Fatalf("TestScripts couldn't Unmarshal: %v", err)
	}

	testScripts(t, tests)

}

// parse hex string into a []byte.
func parseHex(tok string) ([]byte, error) {
	if !strings.HasPrefix(tok, "0x") {
		return nil, errors.New("not a hex number")
	}
	return hex.DecodeString(tok[2:])
}

// shortFormOps holds a map of opcode names to values for use in short form
// parsing.  It is declared here so it only needs to be created once.
var shortFormOps map[string]byte

// parseShortForm parses a string as as used in the Bitcoin Core reference tests
// into the script it came from.
//
// The format used for these tests is pretty simple if ad-hoc:
//   - Opcodes other than the push opcodes and unknown are present as
//     either OP_NAME or just NAME
//   - Plain numbers are made into push operations
//   - Numbers beginning with 0x are inserted into the []byte as-is (so
//     0x14 is OP_DATA_20)
//   - Single quoted strings are pushed as data
//   - Anything else is an error
func parseShortForm(script string) ([]byte, error) {
	// Only create the short form opcode map once.
	if shortFormOps == nil {
		ops := make(map[string]byte)
		for opcodeName, opcodeValue := range OpcodeByName {
			if strings.Contains(opcodeName, "OP_UNKNOWN") {
				continue
			}
			ops[opcodeName] = opcodeValue

			// The opcodes named OP_# can't have the OP_ prefix
			// stripped or they would conflict with the plain
			// numbers.  Also, since OP_FALSE and OP_TRUE are
			// aliases for the OP_0, and OP_1, respectively, they
			// have the same value, so detect those by name and
			// allow them.
			if (opcodeName == "OP_FALSE" || opcodeName == "OP_TRUE") ||
				(opcodeValue != OP_0 && (opcodeValue < OP_1 ||
					opcodeValue > OP_16)) {

				ops[strings.TrimPrefix(opcodeName, "OP_")] = opcodeValue
			}
		}
		shortFormOps = ops
	}

	// Split only does one separator so convert all \n and tab into  space.
	script = strings.Replace(script, "\n", " ", -1)
	script = strings.Replace(script, "\t", " ", -1)
	tokens := strings.Split(script, " ")
	builder := NewScriptBuilder()

	for _, tok := range tokens {
		if len(tok) == 0 {
			continue
		}
		// if parses as a plain number
		if num, err := strconv.ParseInt(tok, 10, 64); err == nil {
			builder.AddInt64(num)
			continue
		} else if bts, err := parseHex(tok); err == nil {
			// Concatenate the bytes manually since the test code
			// intentionally creates scripts that are too large and
			// would cause the builder to error otherwise.
			if builder.err == nil {
				builder.script = append(builder.script, bts...)
			}
		} else if len(tok) >= 2 &&
			tok[0] == '\'' && tok[len(tok)-1] == '\'' {
			builder.AddFullData([]byte(tok[1 : len(tok)-1]))
		} else if opcode, ok := shortFormOps[tok]; ok {
			builder.AddOp(opcode)
		} else {
			return nil, fmt.Errorf("bad token %q", tok)
		}

	}
	return builder.Script()
}
