from exercise import check_even_or_odd

def test_check_even_or_odd():
    assert check_even_or_odd(4) == 'Even'
    assert check_even_or_odd(3) == 'Odd'
    assert check_even_or_odd(0) == 'Even'
    assert check_even_or_odd(-1) == 'Odd'
