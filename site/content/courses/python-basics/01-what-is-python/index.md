---
title: "What is Python?"
weight: 1
duration_sec: 231
outcomes:
  - "Explain what a programming language does"
  - "Install Python on your own computer"
  - "Run your very first line of Python"
---

{{< lesson-video src="final.mp4" captions="captions.vtt" >}}

## A Language For Talking To Computers

Let's talk about Python. It's a language, but not the kind we speak every day. Instead, it's a way to communicate with computers. You see, computers can't guess what we mean. They need clear instructions. A programming language like Python gives us a precise way to write those instructions. Python is popular because it reads almost like English, making it easier for beginners. It was created by Guido van Rossum and first released in 1991. Interestingly, it’s named after Monty Python, not the snake. Now, let's look at how Python translates our instructions into actions.  
[DIAGRAM: python-translator]

{{< diagram src="diagrams/python-translator.svg" caption="A left-to-right flowchart showing Python translating a person's code into computer actions" >}}

## Why Beginners Start With Python

So, why do many beginners start with Python? First, it has very little ceremony. You can write a single, readable line of code that does real work. Plus, there's a huge community out there. Almost any question you might have has an answer online. And what’s great is that the same language can scale from tiny scripts to serious software. It's flexible and powerful.

## Where Python Is Used In The Real World

Python is used in many exciting areas. For example, it powers websites and apps like Instagram and Spotify. It's also used for data analysis and science, making it like spreadsheets on steroids. Python can automate boring tasks, like renaming a thousand files in just a second. And when it comes to AI and machine learning, most of that is written in Python. Now, let’s see a visual of where Python is used.  
[DIAGRAM: where-python-runs]

{{< diagram src="diagrams/where-python-runs.svg" caption="A hub-and-spoke flowchart with a central node labelled 'Python' connected by a plain edge…" >}}

## Installing Python

Ready to get started? First, you'll need to install Python. Go to python.org/downloads and grab the latest version for your system. If you're on Windows, make sure to check the 'Add Python to PATH' box during installation. For macOS, the installer usually works without extra steps. Now, if you're using Linux, Python might already be installed, but this can vary by distribution. To check if it’s installed, you can open a terminal and type python3 --version. This will show you the version if it’s there.

## Your First Line Of Python

Now, let's write our first line of Python. A common first program is to make the computer say hello. To do this, we use a function called print(). A function is like a tool that performs a specific task. With print(), we can show text on the screen. Here’s how it looks:  
print("Hello, world!")  
When you run this, the computer will display:  
Hello, world!  
Let’s see this in action.  
[DEMO: open the Python REPL with python3, print a hello message, then print your own name, and exit]

## Python Does Math Too

Python can do math as well! You don’t need quotes for numbers. Let’s start with a simple example. You can use print() to show a number directly. For instance:  
print(42)  
This will display:  
42  
Now, you can also do calculations. For example, if you write:  
print(7 * 6)  
This will give you:  
42  
And if you add:  
print(10 + 5)  
You’ll see:  
15.  
This is a straightforward way to perform calculations.

## Putting Lines Together

Let’s talk about how a program works. A program is just a series of lines that run one after another, from top to bottom. For example, if we write:  
name = "Ada"  
print("Hello,", name)  
print("Welcome to Python!")  
When you run this, it will first greet Ada and then welcome you to Python. This shows how Python executes each line in order. Let’s put this into practice.  
[DEMO: create a two-line script hello.py with a heredoc and run it with python3 hello.py]

## Whats Next

What’s next? In our next lesson, we’ll dive into variables and how programs remember things. But before then, I encourage you to try a challenge: make print() say your own name. It’s a fun way to practice what you've learned.

## Check your understanding

{{< quiz src="quiz.json" >}}
