"""Open Terminal.app on ~/arena with a neutral prompt and no identifying title."""
import subprocess, sys, time
RC = sys.argv[1]
osa = lambda s: subprocess.run(["osascript","-e",s],capture_output=True,text=True).stdout.strip()
osa('tell application "Terminal" to activate')
time.sleep(1.5)
osa(f'tell application "Terminal" to do script "ZDOTDIR=/dev/null; source {RC}"')
time.sleep(1.5)
# strip anything identifying from the window chrome
osa('''tell application "Terminal"
    set custom title of front window to "arena"
end tell''')
print("terminal ready")
