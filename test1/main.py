import requests

# take a function that takes list of values and returns true if contains duplicates
def dupes(vals):
    valDict = {}
    for val in vals:
        if val in valDict:
            return True
        valDict[val] = 1
    return False

# false
# print(dupes([1, 2, 3]))
# true
# print(dupes([1, 1, 2, 3]))
# true
# print(dupes([1, 2, 1, 3]))
# true
# print(dupes([1, 1, 1, 1]))

# two lists
# one local
# one on server


def dupesRemote(vals):
    r = requests.get("http://127.0.0.1:8000")
    srvVals = r.json()
    combined = vals + srvVals
    return dupes(combined)

# True
# print(dupesRemote([1]))
# False
# print(dupesRemote([4, 5, 6]))
# True
# print(dupesRemote([4, 5, 6, 4]))

# data on server is large, 1tb
# strings of equal length, random
