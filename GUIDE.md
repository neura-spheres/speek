# Speek Language Guide

This guide covers everything you need to know to write programs in Speek. You do not need any programming experience to follow along. Just read it from top to bottom and try each example as you go.

---

## Table of Contents

1. [Running a Speek file](#1-running-a-speek-file)
2. [Variables](#2-variables)
3. [Printing output](#3-printing-output)
4. [Math](#4-math)
5. [Conditions (if / else)](#5-conditions-if--else)
6. [Loops](#6-loops)
7. [Functions](#7-functions)
8. [Lists](#8-lists)
9. [Input from the user](#9-input-from-the-user)
10. [Built-in functions](#10-built-in-functions)
11. [Files](#11-files)
12. [Comments](#12-comments)
13. [Tips and tricks](#13-tips-and-tricks)

---

## 1. Running a Speek file

Create a file with the `.spk` extension and run it from the terminal.

```
speek run myfile.spk
```

If you want to see how Speek reads each line before running, use the debug flag.

```
speek run myfile.spk --debug
```

If you just want to check for errors without actually running the program, use check.

```
speek check myfile.spk
```

You can also use the interactive shell to type and run code one line at a time.

```
speek repl
```

---

## 2. Variables

A variable is a place to store a value. You need to create a variable before you can use it.

### Creating a variable

```
make a variable called score
create x
declare result
```

All three lines do the same thing. Use whichever feels natural to you.

### Giving a variable a value

```
set score to 100
put "hello" into message
change score to 200
score = 42
```

Again, all of these do the same thing. Pick one style and stick with it.

### Creating and setting at the same time

You can write it on one line using a comma.

```
make a variable called score, set score to 100.
```

The comma means "and then do this next". The period ends the group of statements.

### Deleting a variable

```
delete score
remove x
clear result
```

After deleting, the variable no longer exists. Trying to use it after that will cause an error.

---

## 3. Printing output

To show something on the screen, use any of these.

```
show score
print "Hello, world!"
display message
say "Done"
log result
```

You can print text directly by wrapping it in double quotes, or print a variable by just writing its name.

```
make a variable called name
set name to "Darrien"
show "Hello"
show name
```

Output:
```
Hello
Darrien
```

---

## 4. Math

### Basic operations

```
add 5 to score
increase score by 10
subtract 3 from score
multiply score by 2
divide score by 4
```

These all modify the variable directly. So if score is 10 and you add 5, score becomes 15.

### Inline expressions

You can write math expressions directly when assigning a value.

```
make a variable called a, set a to 10.
make a variable called b, set b to 20.
make a variable called c, set c to 30.

make a variable called total
set total to a + b + c
show total
```

Output:
```
60
```

You can also write it like this.

```
total is a + b + c
```

Parentheses work too.

```
make a variable called result
result is (a + b) * c
show result
```

Output:
```
900
```

Supported operators:

| Symbol | What it does |
|--------|-------------|
| `+` | addition |
| `-` | subtraction |
| `*` | multiplication |
| `/` | division |
| `%` | remainder (modulo) |
| `**` | power (e.g. 2 ** 8 = 256) |

---

## 5. Conditions (if / else)

Use if to run code only when a certain condition is true.

```
make a variable called score
set score to 75

if score is greater than 50
  show "You passed"
end
```

### if / else

```
if score is greater than 50
  show "You passed"
else
  show "You failed"
end
```

### if / else if / else

```
if score is greater than 90
  show "A"
else if score is greater than 75
  show "B"
else if score is greater than 60
  show "C"
else
  show "F"
end
```

### Comparison words you can use

All of these work. Use whatever reads naturally to you.

| Meaning | Words you can use |
|---------|------------------|
| greater than | greater than, more than, above, over, exceeds |
| less than | less than, fewer than, below, under |
| greater than or equal | at least, no less than, greater than or equal to |
| less than or equal | at most, no more than, less than or equal to |
| equal | equals, is, same as, matches, equal to, identical to |
| not equal | not equal, not equal to, different from |
| divisible by | divisible by, is divisible by, a multiple of |

Example using different words:

```
if score exceeds 90
  show "Excellent"
end

if score at least 60
  show "Passing"
end

if score is divisible by 2
  show "Even number"
end
```

---

## 6. Loops

### Repeat a fixed number of times

```
loop 5 times
  show "hello"
end
```

Output:
```
hello
hello
hello
hello
hello
```

### Count from one number to another

```
for i from 1 to 5
  show i
end
```

Output:
```
1
2
3
4
5
```

### Loop while a condition is true

```
make a variable called x
set x to 1

while x is less than or equal to 5
  show x
  add 1 to x
end
```

Output:
```
1
2
3
4
5
```

You can also write it as:

```
keep going while x is less than 5
```

### For each item in a list

This is covered in the Lists section. See section 8.

### Breaking out of a loop early

Use break to stop a loop before it finishes.

```
for i from 1 to 10
  if i equals 5
    break
  end
  show i
end
```

Output:
```
1
2
3
4
```

### Skipping to the next iteration

Use continue or skip to jump to the next loop iteration without finishing the current one.

```
for i from 1 to 5
  if i equals 3
    continue
  end
  show i
end
```

Output:
```
1
2
4
5
```

---

## 7. Functions

A function is a block of code you can reuse. You define it once and call it whenever you need it.

### Defining a function with no parameters

```
define a function called greet
  show "Hello!"
end
```

### Calling the function

```
call greet
```

### Function with parameters

Parameters are values you pass in when calling the function.

```
define a function called greet that takes name
  show "Hello,"
  show name
end

call greet with "Darrien"
call greet with "Hannah"
```

Output:
```
Hello,
Darrien
Hello,
Hannah
```

### Function that returns a value

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

Output:
```
10
```

### Multiple parameters

Separate parameters with commas when calling.

```
define a function called add that takes a, b
  make a variable called total
  total is a + b
  return total
end

make a variable called sum
set sum to add with 3, 7
show sum
```

Output:
```
10
```

---

## 8. Lists

A list holds multiple values in order.

### Creating a list

```
make a list called fruits
```

### Adding items

```
add "apple" to fruits
add "banana" to fruits
add "mango" to fruits
```

### Showing all items

```
show fruits
```

### Looping through a list

```
for each fruit in fruits
  show fruit
end
```

Output:
```
apple
banana
mango
```

### Getting the length

```
show length of fruits
```

Output:
```
3
```

### Getting a specific item

Items are numbered starting from 1.

```
show first item in fruits
show last item in fruits
show fruits at 2
```

### Removing items

```
remove "banana" from fruits
pop from fruits
remove last from fruits
remove first from fruits
```

`pop` removes the last item. `remove last` and `remove first` do what they say.

### Sorting

```
sort list fruits
sort list fruits descending
```

### Other list operations

```
shuffle list fruits
reverse list fruits
unique items in fruits
```

### Math on a list of numbers

```
make a list called scores
add 80 to scores
add 90 to scores
add 70 to scores

show sum of scores
show average of scores
```

Output:
```
240
80
```

---

## 9. Input from the user

To ask the user to type something, use ask or read.

### Basic input

```
make a variable called name
ask for name
show name
```

### Input with a prompt message

```
make a variable called name
ask "What is your name?" for name
show name
```

### Reading a number

```
make a variable called age
read number for age
show age
```

---

## 10. Built-in functions

Speek comes with a lot of built-in functions you can use without defining anything. You call them naturally as part of a sentence.

### Math functions

```
set result to square root of 144
set result to absolute value of -5
set result to ceiling of 4.2
set result to floor of 4.9
set result to round of 4.5
set result to power of 2 and 8
set result to log of 100
set result to sin of 90
set result to cos of 0
```

Constants:

```
set result to pi
set result to infinity
```

Clamping a value between a min and max:

```
set result to clamp score between 0 and 100
```

### String functions

```
set result to uppercase name
set result to lowercase name
set result to length of name
set result to reverse name
set result to trim name
```

Checking contents:

```
set result to contains "hello" in message
set result to starts with "He" in message
set result to ends with "lo" in message
```

Modifying strings:

```
set result to replace "world" with "Darrien" in message
set result to repeat "ha" 3 times
set result to substring of message from 1 to 4
```

Splitting and joining:

```
make a list called words
set words to split message by " "

make a variable called joined
set joined to join words with ", "
```

Finding position:

```
set result to index of "a" in name
```

### Type checking and conversion

```
set result to type of score
set result to score as string
set result to score as number
set result to score as bool

set result to is number score
set result to is string name
set result to is list fruits
set result to is empty name
```

### Random

```
set result to random number
set result to random between 1 and 10
set result to random item from fruits
```

### Time

```
set result to current time
set result to current year
set result to current month
set result to current day
set result to current hour
set result to current minute
set result to current second
set result to format time result
```

Pausing the program:

```
wait 2 seconds
```

---

## 11. Files

You can read and write files directly from Speek.

### Reading a file

```
make a variable called content
set content to read file "data.txt"
show content
```

### Writing to a file

This creates the file if it does not exist, and overwrites it if it does.

```
write "Hello, world!" to file "output.txt"
```

### Appending to a file

This adds to the end of the file without erasing what is already there.

```
append "New line" to file "output.txt"
```

### Checking if a file exists

```
set result to file exists "data.txt"
show result
```

### Deleting a file

```
delete file "output.txt"
```

### Listing files in a folder

```
make a list called files
set files to list files in "."
for each f in files
  show f
end
```

---

## 12. Comments

Comments are lines that Speek ignores. Use them to leave notes in your code.

```
# This is a comment
-- This is also a comment
// This works too
note: this is a note
```

---

## 13. Tips and tricks

### Writing multiple statements on one line

You can use a comma to separate statements and a period to end a group.

```
make a variable called x, set x to 10. show x.
```

This is the same as writing:

```
make a variable called x
set x to 10
show x
```

### Speek is case-insensitive

Keywords are not case sensitive. All of these work the same.

```
SHOW "hello"
Show "hello"
show "hello"
```

### You have many ways to say the same thing

Speek understands synonyms. These all do the same thing.

```
make a variable called x
create x
declare x
initialize x
```

Pick whatever feels most natural and stick with it. There is no wrong choice.

### Debug mode shows you what Speek thinks you wrote

If your program is not doing what you expect, run it with --debug to see how each line is being interpreted.

```
speek run myfile.spk --debug
```

Example output:

```
Line  Source                    Interpretation
----  ------------------------  ------------------
1     make a variable called x  DECLARE x = null
2     set x to 10               ASSIGN x <- 10
3     show x                    PRINT x
```

### Speek suggests fixes when it does not understand a line

If you make a typo, Speek tries to guess what you meant.

```
shw "hello"
```

Output:
```
Error on line 1: I don't understand 'shw "hello"'
Did you mean: show?
```

---

That is everything you need to write real programs in Speek. Start small, try the examples, and build from there.
