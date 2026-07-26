from exercise import say_name

def test_say_name(capsys):
    say_name()
    captured = capsys.readouterr()
    assert captured.out == "Hello, Chris! Welcome to Python!\n"
