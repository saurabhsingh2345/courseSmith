Output demos/demo-2.mp4
Set FontSize 34
Set Width 1440
Set Height 640
Set Padding 32
Set TypingSpeed 80ms
Set Shell bash
Set Theme { "background": "#070c15", "foreground": "#f2f5f8", "cursor": "#ffd43b", "yellow": "#ffd43b", "brightYellow": "#ffd43b" }

Type "cat > about_me.py << 'EOF'"
Enter
Type "name = 'Ada'"
Enter
Type "age = 12"
Enter
Type "print(name, 'is', age)"
Enter
Type "print(name, 'turns', age + 1, 'next year')"
Enter
Type "EOF"
Enter
Type "python3 about_me.py"
Enter
Sleep 1s
Sleep 2s
Type "exit()"
Enter
