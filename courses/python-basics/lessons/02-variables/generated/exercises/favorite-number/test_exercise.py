from exercise import favorite_number

def test_favorite_number():
    assert favorite_number("Ada", 7) == "Ada's favorite number is 7."

def test_favorite_number_zero():
    assert favorite_number("Ada", 0) == "Ada's favorite number is 0."
