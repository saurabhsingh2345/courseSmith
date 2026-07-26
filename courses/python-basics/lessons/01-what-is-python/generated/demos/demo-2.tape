Output demos/demo-2.mp4
Set FontSize 34
Set Width 1440
Set Height 640
Set Padding 32
Set TypingSpeed 80ms
Set Shell bash
Set Theme { "background": "#070c15", "foreground": "#f2f5f8", "cursor": "#ffd43b", "yellow": "#ffd43b", "brightYellow": "#ffd43b" }

Type "cat > hello.py << 'EOF'"
Enter
Type "print('Hello, world!')"
Enter
Type "EOF"
Enter
Type "python3 hello.py"
Enter
Sleep 1s
Type "print(7 * 6)"
Enter
Type "print(10 + 5)"
Enter
Type "exit()"
Enter
Sleep 3s
