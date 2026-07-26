from exercise import greet

def test_greet():
    assert greet("Ada") == "Hello, Ada!"
    assert greet("Bob") == "Hello, Bob!"
    assert greet("Alice") == "Hello, Alice!"
