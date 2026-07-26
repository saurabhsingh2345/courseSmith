# 01-what-is-python — What is Python?

_Generated for human review. Leave feedback in `review-notes.yaml` (template at the bottom); the next pipeline run applies your notes and marks them resolved._

## Narration script

### a-language-for-talking-to-computers (~63s)

So, you wanna know what Python is? Well, computers only follow instructions, they can't guess what you mean. A programming language is a precise way of writing instructions. Python is one of the most popular ones, designed to read almost like English. It was created by Guido van Rossum, first released in 1991, and named after Monty Python, not the snake. Let's look at how Python works as a translator between you and a computer. It reads your code line by line and carries it out, so let's look at that in a picture

**Diagram `python-translator`** (shown at word 23):

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 400">
  <style>
    svg { --primary: #306998; --accent: #ffd43b; --bg: #ffffff; }
    .person { fill: var(--primary); stroke: none; }
    .computer { fill: var(--primary); stroke: none; }
    .screen { fill: var(--bg); stroke: var(--primary); stroke-width: 3; rx: 10; }
    .box { fill: var(--bg); stroke: var(--primary); stroke-width: 3; rx: 14; }
    .label { font-family: sans-serif; font-size: 20px; fill: #1f2937; text-anchor: middle; }
    .code { font-family: monospace; font-size: 18px; fill: #1f2937; }
    .arrow { stroke: var(--accent); stroke-width: 4; fill: none; marker-end: url(#head); }
  </style>
  <defs>
    <marker id="head" markerWidth="10" markerHeight="8" refX="9" refY="4" orient="auto">
      <path d="M0,0 L10,4 L0,8 z" fill="#ffd43b"/>
    </marker>
  </defs>
  <g id="background">
    <rect width="800" height="400" fill="var(--bg)"/>
  </g>
  <g id="person">
    <circle class="person" cx="150" cy="200" r="50"/>
    <circle cx="150" cy="150" r="30"/>
    <path d="M 120 220 L 150 250 L 180 220 z" fill="#ffffff"/>
    <text class="code" x="150" y="180">print("Hello")</text>
  </g>
  <g id="python-box">
    <rect class="box" x="350" y="120" width="100" height="80"/>
    <text class="label" x="400" y="160">Python interpreter</text>
    <line class="arrow" x1="250" y1="200" x2="350" y2="200"/>
  </g>
  <g id="label-under-box">
    <text class="label" x="400" y="240">reads your code line by line and carries it out</text>
  </g>
  <g id="computer">
    <rect x="550" y="150" width="100" height="100" fill="#ffffff" stroke="var(--primary)" stroke-width="3" rx="10"/>
    <rect class="screen" x="560" y="160" width="80" height="40"/>
    <text class="label" x="600" y="180">Hello</text>
    <line class="arrow" x1="450" y1="200" x2="550" y2="200"/>
  </g>
</svg>

### why-beginners-start-with-python (~46s)

So, why do beginners start with Python? Well, it's got very little ceremony, one readable line can do real work. There's a huge community, almost any question you'll have has an answer online. And the same language scales from tiny scripts to serious software. That's pretty cool, right? It's easy to get started and you can build on that as you learn more.

### where-python-is-used-in-the-real-world (~73s)

Now, where is Python used in the real world? It's used in websites and apps, like Instagram and Spotify backends. It's used in data analysis and science, like spreadsheets on steroids. It's used for automating boring tasks, like renaming 1,000 files in a second. And it's even used in AI and machine learning, most of it is written in Python. Let's take a look at some examples of where Python is used.

**Diagram `where-python-runs`** (shown at word 20):

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 600">
  <style>
    svg { --primary: #306998; --accent: #ffd43b; --bg: #ffffff; }
    .box { fill: var(--bg); stroke: var(--primary); stroke-width: 3; rx: 14; }
    .icon { fill: var(--primary); }
    .label { font-family: sans-serif; font-size: 18px; fill: #1f2937; text-anchor: middle; }
    .snake { fill: var(--accent); }
    .connector { stroke: var(--primary); stroke-width: 2; }
  </style>
  <g id="background">
    <rect width="800" height="600" fill="var(--bg)"/>
  </g>
  <g id="top-left-box">
    <rect class="box" x="100" y="100" width="200" height="150"/>
    <rect class="icon" x="120" y="120" width="40" height="40" rx="5"/>
    <text class="label" x="200" y="220">Websites</text>
  </g>
  <g id="top-right-box">
    <rect class="box" x="500" y="100" width="200" height="150"/>
    <rect class="icon" x="520" y="120" width="40" height="40" rx="5"/>
    <text class="label" x="600" y="220">Data &#38; Science</text>
  </g>
  <g id="bottom-left-box">
    <rect class="box" x="100" y="300" width="200" height="150"/>
    <rect class="icon" x="140" y="320" width="20" height="20" rx="5"/>
    <text class="label" x="200" y="420">Automation</text>
  </g>
  <g id="bottom-right-box">
    <rect class="box" x="500" y="300" width="200" height="150"/>
    <circle class="icon" cx="580" cy="340" r="20"/>
    <text class="label" x="600" y="420">AI &#38; Machine Learning</text>
  </g>
  <g id="snake">
    <circle class="snake" cx="400" cy="300" r="50"/>
  </g>
  <g id="connectors">
    <line class="connector" x1="400" y1="300" x2="150" y2="175"/>
    <line class="connector" x1="400" y1="300" x2="650" y2="175"/>
    <line class="connector" x1="400" y1="300" x2="150" y2="425"/>
    <line class="connector" x1="400" y1="300" x2="650" y2="425"/>
  </g>
</svg>

### installing-python (~74s)

Okay, so you wanna install Python. That's easy. Just go to python.org/downloads and grab the latest version for your system. On Windows, make sure to check the 'Add Python to PATH' box in the installer. On macOS, the installer just works. And on Linux, Python is usually already installed. To check it worked, open a terminal and type python3 --version. That should show you the version of Python you just installed.

### your-first-line-of-python (~51s)

Now, let's write your first line of Python. The traditional first program is to make the computer say hello. We use the print function, which takes whatever is inside the quotes and shows it on screen. So, let's try that out in a demo.

**Demo** (at word 25): open the Python REPL with python3, print a hello message, then print your own name, and exit

### python-does-math-too (~43s)

Python can also do math. You don't need quotes for numbers, Python calculates them. The REPL answers back immediately, like a very obedient calculator. So, let's try some math.

### putting-lines-together (~51s)

A program is just lines executed top to bottom, one after another. You can save them in a file, run the file, and Python performs each line in order. Let's try that out in a demo.

**Demo** (at word 20): create a two-line script hello.py with a heredoc and run it with python3 hello.py

### what-s-next (~46s)

So, what's next? In the next lesson, we'll cover variables, which is how programs remember things. And before then, try to make print say your own name. That's a challenge for you to try on your own.

## Quiz — Introduction to Python

**1. []** Who created the Python programming language?

- [x] Guido van Rossum
- [ ] Monty Python
- [ ] Alan Turing
- [ ] Steve Jobs

_Guido van Rossum created Python, and it's named after the British comedy group Monty Python, not the snake._

**2. []** What is the purpose of the print() function in Python?

- [ ] To store data in a variable
- [ ] To perform mathematical calculations
- [x] To show text on the screen
- [ ] To exit the program

_The print() function is used to display text on the screen, taking whatever is inside the quotes as its input._

**3. []** How do you check if Python is installed correctly on your system?

- [ ] By running a Python script
- [ ] By checking the Python version in the installer
- [x] By opening a terminal and typing python3 --version
- [ ] By restarting your computer

_To verify that Python is installed correctly, you can open a terminal and type python3 --version, which should display the version of Python you just installed._

**4. []** What happens when you run a Python program?

- [ ] The lines are executed in reverse order
- [ ] The lines are executed randomly
- [x] The lines are executed top to bottom, one after another
- [ ] The lines are ignored

_A Python program is executed line by line, from top to bottom, with each line being performed in order._

**5. []** What is the result of the expression print(7 * 6) in Python?

- [x] 42
- [ ] 12
- [ ] 7
- [ ] 6

_The expression print(7 * 6) calculates the product of 7 and 6, which is 42, and then prints the result._

## Automated quality flags

- TTS accuracy: overall WER 0.0% (flagged sections: none)
- pace: target 145 wpm ±15% (out-of-band: a-language-for-talking-to-computers; why-beginners-start-with-python; your-first-line-of-python; putting-lines-together; what-s-next)
- loudness: -25.5 → -16.1 LUFS (target -16), true peak -1.4 dBTP

## Leaving feedback

Add notes to `review-notes.yaml` in the course directory:

```yaml
lessons:
  01-what-is-python:
    notes:
      - note: "Lesson-wide feedback here"
    sections:
      a-language-for-talking-to-computers:
        - note: "Section-specific feedback here"
```
