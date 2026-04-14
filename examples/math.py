from core import math


scores = [0.2, 0.4, 0.9]
print(math.mean(scores))

tree = math.kdtree.build([[0.0, 0.0], [1.0, 1.0], [3.0, 3.0]], metric="euclidean")
print(tree.nearest([0.8, 0.8], k=2))
