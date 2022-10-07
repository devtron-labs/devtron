import os

def getFileExtension(file):
    filename, extension = os.path.splitext(file)
    return extension


def readFileAsString(file):
    with open(file, 'r') as file:
        return file.read()


def writeFileFromString(path, data):
    f = open(path, "w")
    f.write(data)
    f.close()

def read_lines(filename):
    '''
    Read each lines from a file
    '''
    with open(filename, 'r') as file:
        lines = file.readlines()
        rows = [line.strip() for line in lines]  # Strip newline char
        result = [row for row in rows if len(row) != 0]  # Filter empty lines
    return result