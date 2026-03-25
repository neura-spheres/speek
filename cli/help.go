package cli

import "fmt"

// PrintHelp displays help text. topic can be "" (all), "math", "strings", "lists".
func PrintHelp(topic string) {
	switch topic {
	case "math":
		printMathHelp()
	case "strings", "string":
		printStringHelp()
	case "lists", "list":
		printListHelp()
	default:
		printGeneralHelp()
	}
}

func printGeneralHelp() {
	fmt.Println(`
Speek Language - Quick Reference
=================================
Statements can be separated by commas (,) or periods (.).

VARIABLES
  make a variable called score
  create variable x, put 5 into x.
  set score to 100
  put "hello" into name
  delete score

PRINT
  show score
  print "Hello, world!"
  display name

MATH
  add 5 to score
  subtract 3 from score
  multiply score by 2
  divide score by 4
  increase score by 10

CONDITIONALS
  if score is greater than 50
    show "high"
  else if score equals 50
    show "exact"
  else
    show "low"
  end

LOOPS
  loop 10 times
    show i
  end

  for x from 1 to 10
    show x
  end

  for each item in mylist
    show item
  end

  while x is less than 100
    add 1 to x
  end

FUNCTIONS
  define a function called greet that takes name
    show "Hello,"
    show name
  end

  call greet with "Darrien"

LISTS
  make a list called fruits
  add "apple" to fruits
  show length of fruits
  for each item in fruits
    show item
  end

INPUT
  ask for name
  ask "What is your name?" for name

BUILT-INS (call with 'call' or 'set x to <builtin> of y')
  set result to square root of 144
  set result to uppercase name
  set result to length of message
  call random number
  set result to current year

For more: speek help math | speek help strings | speek help lists
`)
}

func printMathHelp() {
	fmt.Println(`
Math Built-ins
===============
  abs x                    absolute value
  square root of x         sqrt
  cube root of x           cbrt
  power of x to y          x raised to y (use: call power with x, y)
  ceiling of x             ceil
  floor of x               floor
  round x                  round to nearest integer
  truncate x               drop the decimal
  log of x                 natural logarithm
  log base 10 of x         log10
  log base 2 of x          log2
  sin of x                 sine
  cos of x                 cosine
  tan of x                 tangent
  arcsin of x              asin
  arccos of x              acos
  arctan of x              atan
  max of x and y           maximum
  min of x and y           minimum
  sign of x                -1, 0, or 1
  pi                       3.14159...
  e                        2.71828...
  clamp x between a and b  min(max(x,a),b)
  is nan x                 check if not-a-number
  is infinite x            check if infinite

Examples:
  set result to square root of 144
  set result to abs score
  set result to pi
  call clamp with x, 0, 100
`)
}

func printStringHelp() {
	fmt.Println(`
String Built-ins
=================
  length of x              character count
  uppercase x              convert to uppercase
  lowercase x              convert to lowercase
  reverse x                reverse the string
  trim x                   remove surrounding whitespace
  contains "sub" in x      check if substring exists
  starts with "x" in y     prefix check
  ends with "x" in y       suffix check
  replace x with y in z    find and replace
  split x by y             split into list
  join list x with y       join list items with separator
  index of x in y          position of substring (-1 if not found)
  substring of x from a to b  extract portion
  repeat x n times         repeat string
  count x in y             count occurrences
  is empty x               true if empty string
  pad left x to n with y   left-pad to length
  pad right x to n with y  right-pad to length
  first char of x          first character
  last char of x           last character
  char at n in x           character at index

Examples:
  set result to uppercase message
  set result to length of name
  set result to trim input
`)
}

func printListHelp() {
	fmt.Println(`
List Operations
===============
  make a list called fruits
  add "apple" to fruits
  append "banana" to fruits
  push "mango" into fruits
  remove "apple" from fruits
  pop from fruits
  remove last from fruits
  remove first from fruits

  length of list fruits
  first item in fruits
  last item in fruits
  get item 2 from fruits
  fruits at 0

  sort list fruits
  sort list fruits descending
  shuffle list fruits
  reverse list fruits
  unique items in fruits
  sum of fruits
  average of fruits
  slice fruits from 1 to 3

  for each item in fruits
    show item
  end

Examples:
  make a list called scores
  add 95 to scores
  add 87 to scores
  set total to sum of scores
  set avg to average of scores
`)
}
