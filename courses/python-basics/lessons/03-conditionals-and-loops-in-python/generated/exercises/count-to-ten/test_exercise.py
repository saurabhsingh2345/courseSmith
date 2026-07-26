from exercise import count_to_ten

def test_count_to_ten():
    assert count_to_ten() == [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
    assert len(count_to_ten()) == 10
