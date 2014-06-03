// Copyright 2014 SteelSeries ApS.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package impliments a basic LISP interpretor for embedding in a go program for scripting.
// This file provides a repl
package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/steelseries/golisp"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

func loadLispFile(filename string, fileInfo os.FileInfo, err error) error {
	match, _ := filepath.Match("*.lsp", fileInfo.Name())
	if match {
		_, err = ProcessFile(filename)
		if err != nil {
			err = errors.New(fmt.Sprintf("Error GoLisp file: '%s': %s", fileInfo.Name(), err))
		}
	}
	return err
}

func loadLispCode() {
	err := filepath.Walk("./lisp", loadLispFile)
	if err != nil {
		panic(err)
	}
}

func dump(node ast.Expr) {
	binaryExpr := node.(BinaryExpr)
}

//============================  STACK

type Stack []interface{}

func (stack *Stack) Pop() (interface{}, error) {
	theStack := *stack
	if len(theStack) == 0 {
		return nil, errors.New("can't Pop() an empty stack")
	}
	x := theStack[len(theStack)-1]
	*stack = theStack[:len(theStack)-1]
	return x, nil
}

func (stack *Stack) Push(x interface{}) {
	*stack = append(*stack, x)
}

func (stack Stack) Top() (interface{}, error) {
	if len(stack) == 0 {
		return nil, errors.New("can't Top() an empty stack")
	}
	return stack[len(stack)-1], nil
}

func (stack Stack) Cap() int {
	return cap(stack)
}

func (stack Stack) Len() int {
	return len(stack)
}

func (stack Stack) IsEmpty() bool {
	return len(stack) == 0
}

//============================  EVALUATOR

var whitespace_rx = regexp.MustCompile(`\s+`)
var fp_rx = regexp.MustCompile(`(\d+(?:\.\d+)?)`) // simple fp number
var func_rx = regexp.MustCompile(`([:alpha:]+)`)
var operators = "-+**/<>"

// prec returns the operator's precedence
func prec(op string) (result int) {
	if op == "-" || op == "+" {
		result = 1
	} else if op == "*" || op == "/" {
		result = 2
	} else if op == "**" {
		result = 3
	}
	return
}

// opGTE returns true if op1's precedence is >= op2
func opGTE(op1, op2 string) bool {
	return prec(op1) >= prec(op2)
}

// isOperator returns true if token is an operator
func isOperator(token string) bool {
	return strings.Contains(operators, token)
}

// isOperand returns true if token is an operand
func isOperand(token string) bool {
	return fp_rx.MatchString(token)
}

// isLispFunction returns true if token is the name of a function in
// the lisp global symbol table
func isLispFunction(token string) bool {
	return func_rx.MatchString(token)
}

// convert2postfix converts an infix expression to postfix
func convert2postfix(tokens []string) []string {
	var stack stack.Stack
	var result []string
	for _, token := range tokens {

		if isLispFunction(token) {

		} else if isOperator(token) {

		OPERATOR:
			for {
				top, err := stack.Top()
				if err == nil && top != "(" {
					if opGTE(top.(string), token) {
						pop, _ := stack.Pop()
						result = append(result, pop.(string))
					} else {
						break OPERATOR
					}
				}
				break OPERATOR
			}
			stack.Push(token)

		} else if token == "(" {
			stack.Push(token)

		} else if token == ")" {
		PAREN:
			for {
				top, err := stack.Top()
				if err == nil && top != "(" {
					pop, _ := stack.Pop()
					result = append(result, pop.(string))
				} else {
					stack.Pop() // pop off "("
					break PAREN
				}
			}

		} else if isOperand(token) {
			result = append(result, token)
		}

	}

	for !stack.IsEmpty() {
		pop, _ := stack.Pop()
		result = append(result, pop.(string))
	}

	return result
}

// evaluatePostfix takes a postfix expression and evaluates it
func evaluatePostfix(postfix []string) (*big.Rat, error) {
	var stack stack.Stack
	result := new(big.Rat) // note: a new(big.Rat) has value "0/1" ie zero
	for _, token := range postfix {
		if isOperand(token) {
			bigrat := new(big.Rat)
			if _, err := fmt.Sscan(token, bigrat); err != nil {
				return nil, fmt.Errorf("unable to scan %s", token)
			}
			stack.Push(bigrat)
		} else if isOperator(token) {

			op2, err2 := stack.Pop()
			if err2 != nil {
				return nil, err2
			}
			op1, err1 := stack.Pop()
			if err1 != nil {
				return nil, err1
			}

			dummy := new(big.Rat)
			switch token {
			case "**":
				float1 := BigratToFloat(op1.(*big.Rat))
				float2 := BigratToFloat(op2.(*big.Rat))
				float_result := math.Pow(float1, float2)
				stack.Push(FloatToBigrat(float_result))
			case "*":
				result := dummy.Mul(op1.(*big.Rat), op2.(*big.Rat))
				stack.Push(result)
			case "/":
				result := dummy.Quo(op1.(*big.Rat), op2.(*big.Rat))
				stack.Push(result)
			case "+":
				result = dummy.Add(op1.(*big.Rat), op2.(*big.Rat))
				stack.Push(result)
			case "-":
				result = dummy.Sub(op1.(*big.Rat), op2.(*big.Rat))
				stack.Push(result)
			case "<":
				if op1.(*big.Rat).Cmp(op2.(*big.Rat)) <= -1 {
					stack.Push(big.NewRat(1, 1))
				} else {
					stack.Push(new(big.Rat))
				}
			case ">":
				if op1.(*big.Rat).Cmp(op2.(*big.Rat)) >= 1 {
					stack.Push(big.NewRat(1, 1))
				} else {
					stack.Push(new(big.Rat))
				}
			}
		} else if isLispFunction(token) {
			//"(<token> <float1> <float2>...)"

		} else {
			return nil, fmt.Errorf("unknown token %v", token)
		}
	}

	retval, err := stack.Pop()
	if err != nil {
		return nil, err
	}
	return retval.(*big.Rat), nil
}

// tokenise takes an expr string and converts it to a slice of tokens
//
// tokenise puts spaces around all non-numbers, removes leading and
// trailing spaces, then splits on spaces
//
func tokenise(expr string) []string {
	spaced := fp_rx.ReplaceAllString(expr, " ${1} ")
	symbols := []string{"(", ")"}
	for _, symbol := range symbols {
		spaced = strings.Replace(spaced, symbol, fmt.Sprintf(" %s ", symbol), -1)
	}
	stripped := whitespace_rx.ReplaceAllString(strings.TrimSpace(spaced), "|")
	return strings.Split(stripped, "|")
}

// Eval takes an infix string arithmetic expression, and evaluates it
//
// Usage:
//   result, err := evaler.Eval("1+2")
// Returns: the result of the evaluation, and any errors
//
func EvalEquation(expr string) (result *big.Rat, err error) {
	defer func() {
		if e := recover(); e != nil {
			result = nil
			err = fmt.Errorf("Invalid Expression: %s", expr)
		}
	}()

	tokens := tokenise(expr)
	postfix := convert2postfix(tokens)
	return evaluatePostfix(postfix)
}

func ioLoop() {
	fmt.Printf("Welcome to the GoLisp Example\n")
	fmt.Printf("Copyright 2014 SteelSeries\n")
	fmt.Printf("Evaluate 'quit' to exit.\n\n")
	prompt := "> "
	LoadHistoryFromFile(".golisp_history")
	lastInput := ""
	for true {
		input := *ReadLine(&prompt)
		if input != "" {
			if input != lastInput {
				AddHistory(input)
			}
			lastInput = input
			ast, err := ParseExpr(input)
			dump(ast)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
			} else {
				d, err := EvalEquation(code, Global)
				if err != nil {
					fmt.Printf("Error in evaluation: %s\n", err)
				} else {
					fmt.Printf("==> %s\n", String(d))
				}
			}
		}
	}
}

func main() {
	loadLispCode()
	ioLoop()
}
