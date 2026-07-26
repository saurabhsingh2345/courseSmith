---
title: "Variables: how programs remember"
weight: 2
duration_sec: 223
outcomes:
  - "Store a value in a variable and use it later"
  - "Choose clear, legal variable names"
  - "Update a variable and predict what prints"
---

{{< lesson-video src="final.mp4" captions="captions.vtt" >}}

## You Already Used One

Last lesson's challenge was to make print say your own name. If you wrote your name straight inside print, you had to repeat it every time. But in the final script, you probably had a line like this: name = "Ada". That line is a variable. This lesson is all about what that line really does. A variable is how a program remembers a value, so it can use it again later.

## Giving A Value A Name

When you write name = "Ada", you're storing the value "Ada" under the name name. Now, the equals sign here means "store this", not "equals" like in math. You put the name of the box on the left, and the value on the right. After that, whenever you write name, it means whatever is in that box. Let's look at that in a picture. 

Here’s a diagram showing how it works. You see that the label stays, but the contents can change. 

So when you run this code, it prints Ada twice because it looks inside the box named name.

{{< diagram src="diagrams/labeled-box.svg" caption="A left-to-right flowchart showing a variable as a labelled box" >}}

## Names Without Quotes Text With Quotes

When you write print("name"), it shows the text name. But if you write print(name), it shows what the variable holds. Quotes mean "these exact characters", while no quotes mean "look inside the box". 

So if you run this code, you'll see that the first print gives you the word name, and the second one gives you the value Ada. It’s a simple but important distinction.

## Picking Good Names

When choosing names for your variables, remember that legal names can use letters, digits, and underscores, but they can't start with a digit. Also, Python is case-sensitive, so age and Age are two different boxes. It's best to choose names that describe what's inside. For example, birth_year is clearer than just by. And if you have multi-word names, use underscores, like favorite_color instead of favoritecolor.

## Changing What Is Inside

Assigning again replaces the old value. The box keeps its label, but the contents are swapped. The old value is gone, and Python only remembers the latest assignment. Let's look at that in a picture. 

This diagram shows how assigning again replaces what was inside. When you run this code, you’ll see that the first print shows 0, and after we assign again, it shows 10.

{{< diagram src="diagrams/reassignment.svg" caption="A left-to-right flowchart showing variable reassignment as a before/after" >}}

## Variables Can Hold Numbers Too

Everything from last lesson's math works inside variables. You can even use a variable's current value to compute a new one. For example, if you start with apples = 3 and then write apples = apples + 2, you’re adding 2 to the current value. When you print apples, it gives you 5. It’s like counting your apples and finding out you have more than you thought!

## Variables Working Together

In real programs, you often mix several variables in one line. The print function can take multiple things separated by commas, and it prints them with spaces in between. For instance, if you have name = "Ada" and age = 12, you can print them together. This prints a nice sentence about Ada and her age. It’s a great way to share information in a clear way.

## Whats Next

In the next lesson, we’ll dive into text itself and explore what strings can do, from joining to changing case. But before then, here’s a challenge for you: create a script with two variables, your name and your favorite number, and print one sentence using both. It’s a fun way to practice what you’ve learned!

## Check your understanding

{{< quiz src="quiz.json" >}}
