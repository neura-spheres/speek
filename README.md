# Speek

![Speek Banner](public/image.png)

A programming language you write in plain English. No AI, no internet required, no syntax to memorize. Just type what you want to happen and it runs.

## Install

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/neura-spheres/speek/main/install.ps1 | iex
```

**macOS / Linux:**
```bash
curl -sSL https://raw.githubusercontent.com/neura-spheres/speek/main/install.sh | bash
```

Both scripts download the binary, add it to PATH, and install the VS Code extension automatically.

**Build from source:**
```bash
git clone https://github.com/neura-spheres/speek
cd speek
go build -o speek.exe .
speek install-vscode
```

## VS Code Extension

The extension is bundled inside the Speek binary. After installing Speek, run:

```bash
speek install-vscode
```

This installs syntax highlighting into every VS Code variant found on your machine (VS Code, Cursor, VSCodium). Restart VS Code after running it.

You get:
- Syntax highlighting for `.spk` files
- Play button in the editor toolbar to run the current file
- F5 runs the current `.spk` file in a terminal
- Right-click menu with Run and Check options
- Auto-indentation for blocks (loops, if, functions)
- Comment toggling with `#`

To update the extension after upgrading Speek:
```bash
speek install-vscode --force
```

The color scheme uses standard TextMate scopes so it works with any VS Code theme:

| Element | Scope | Typical Color |
|---|---|---|
| Comments | `comment.line` | Green/Gray |
| Strings | `string.quoted.double` | Orange/Yellow |
| Numbers | `constant.numeric` | Teal/Green |
| true/false/null | `constant.language` | Purple |
| Loop/if/else/end | `keyword.control` | Purple/Pink |
| make/create/declare | `keyword.other.declaration` | Blue |
| set/put/assign | `keyword.operator.assignment` | Blue |
| show/print/display | `keyword.other.output` | Blue |
| add/subtract/multiply | `keyword.operator.arithmetic` | Blue |
| sqrt/length/uppercase | `support.function.builtin` | Teal |
| to/from/by/with | `keyword.operator.connector` | Subtle |
| User variable names | `variable.other` | Default text |

The extension source is in `vscode-ext/` inside this repo.

## Quick Start

Create a file called `hello.spk`:

```
make a variable called name
set name to "Darrien"
show "Hello,"
show name
```

Run it:
```bash
speek run hello.spk
```

You can also separate statements with commas and periods on one line:
```
create variable x, put 5 into x. show x.
```

## How It Works

Speek has no AI or machine learning. When it starts, it compiles hundreds of regex patterns from a dictionary of synonyms. Every line you write gets matched against those patterns until one fits. If nothing matches, it suggests what you might have meant using edit distance.

This means:
- Starts in under 50ms
- Works completely offline
- Deterministic -- same input always gives the same output
- You can add your own synonym words to the dictionary

## Language Reference

### Variables

```
make a variable called score
create x
set score to 100
put "hello" into message
change score to 200
score = 42
delete score
```

### Print

```
show score
print "Hello, world!"
display message
log result
say "Done"
```

### Math

```
add 5 to score
increase score by 10
subtract 3 from score
multiply score by 2
divide score by 4
```

Inline expressions work anywhere a value is expected:

```
make a variable called total
set total to a + b + c
total is (a + b) * c
```

### Conditionals

```
if score is greater than 50
  show "high score"
else if score equals 50
  show "exactly 50"
else
  show "low score"
end
```

Comparison words you can use:
- greater than, more than, above, over, exceeds
- less than, fewer than, below, under
- at least, no less than, minimum
- at most, no more than, maximum
- equals, is, same as, matches
- not equal, is not, different from
- divisible by, is divisible by

### Loops

**Repeat N times:**
```
loop 10 times
  show "hello"
end
```

**Count from A to B:**
```
for x from 1 to 10
  show x
end
```

**While a condition is true:**
```
make a variable called x
set x to 0
while x is less than 10
  show x
  add 1 to x
end
```

**For each item in a list:**
```
for each fruit in fruits
  show fruit
end
```

Use `break` to exit a loop early.

### Functions

```
define a function called greet that takes name
  show "Hello,"
  show name
end

call greet with "Darrien"
call greet with "Hannah"
```

Functions can return values:
```
define a function called double that takes x
  make a variable called result
  set result to x
  multiply result by 2
  return result
end

make a variable called answer
set answer to double with 5
show answer
```

### Lists

```
make a list called fruits
add "apple" to fruits
add "banana" to fruits
add "mango" to fruits

show length of fruits

for each item in fruits
  show item
end

remove "banana" from fruits
pop from fruits
```

### Input

```
ask for name
ask "What is your name?" for name
read number for age
```

### Built-in Functions

```
set result to square root of 144
set result to uppercase message
set result to length of name
set result to abs score
set result to pi
```

**Math:**
```
abs x                     absolute value
square root of x          sqrt
cube root of x            cbrt
ceiling of x              ceil
floor of x                floor
round x                   round
log of x                  natural log
log base 10 of x          log10
sin of x                  sine
cos of x                  cosine
tan of x                  tangent
sign of x                 -1, 0, or 1
pi                        3.14159...
e                         2.71828...
clamp x between a and b   clamp to range
```

**Strings:**
```
length of x               character count
uppercase x               convert to uppercase
lowercase x               convert to lowercase
reverse x                 reverse the string
trim x                    strip whitespace
contains "sub" in x       true/false
starts with "x" in y      prefix check
ends with "x" in y        suffix check
replace x with y in z     find and replace
split x by y              split into list
join list x with y        join with separator
index of x in y           position of substring
substring of x from a to b  extract slice
repeat x n times          repeat string
is empty x                true if blank
```

**Type conversion:**
```
x as number               parse to number
x as string               convert to string
x as bool                 convert to boolean
type of x                 returns "number", "string", etc.
is number x               type check
is string x               type check
is list x                 type check
```

**List operations:**
```
length of list x          list size
first item in x           index 0
last item in x            last element
sort list x               sort ascending
sort list x descending    sort descending
shuffle list x            random order
reverse list x            reverse
unique items in x         remove duplicates
sum of x                  total of all numbers
average of x              mean
```

**Random:**
```
random number             float between 0 and 1
random between 1 and 10   random integer in range
random item from mylist   random list element
```

**Time:**
```
current time              unix timestamp
current year              calendar year
current month             month number
current day               day of month
current hour              0-23
current minute            0-59
current second            0-59
format time x             human readable date string
wait 2 seconds            sleep
```

**Files:**
```
read file "data.txt"
write "hello" to file "out.txt"
append "more" to file "out.txt"
file exists "data.txt"
delete file "data.txt"
list files in "."
```

**System:**
```
exit with 0
environment variable "PATH"
run command "ls -la"
```

## CLI Commands

```bash
speek run file.spk                  run a program
speek run file.spk --debug          show intent panel before running
speek run file.spk --confirm        ask for confirmation first
speek check file.spk                validate without running
speek repl                         interactive shell
speek install-vscode               install VS Code extension
speek install-vscode --force       reinstall / update extension
speek help                         show all commands
speek help math                    show math built-ins
speek help strings                 show string built-ins
speek help lists                   show list operations
speek version                      version info
```

Debug mode shows how every line is interpreted before it runs:
```
Line  Source                                    Interpretation
----  ----------------------------------------  ---------------
1     make a variable called score              DECLARE score = null
2     set score to 100                          ASSIGN score <- 100
3     show score                                PRINT score
```

## Example Programs

### FizzBuzz

```
make a variable called i
set i to 1
loop 20 times
  if i is divisible by 15
    show "FizzBuzz"
  else if i is divisible by 3
    show "Fizz"
  else if i is divisible by 5
    show "Buzz"
  else
    show i
  end
  add 1 to i
end
```

### Square Root Calculator

```
make a variable called x
set x to 144
make a variable called result
set result to square root of x
show result
```

### Working with Lists

```
make a list called scores
add 95 to scores
add 87 to scores
add 72 to scores
add 100 to scores

make a variable called total
set total to sum of scores
show total

make a variable called avg
set avg to average of scores
show avg

sort list scores descending
for each s in scores
  show s
end
```

## Docs

| File | Description |
|------|-------------|
| [GUIDE.md](GUIDE.md) | Full language tutorial with examples for every feature |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to add new words, statements, and built-ins |
| [LICENSE](LICENSE) | MIT License |
