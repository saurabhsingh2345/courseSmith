from exercise import greet

def test_greet():
    assert greet("Ada") == "Hello, Ada!"

def test_greet_empty():
    assert greet("") == "Hello, !"
