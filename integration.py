import hashlib

BUF_SIZE = 65536

def writeTestFile(name):
	with open(name, "w+") as f:
		f.write("This is a test\n")
	
def hashFile(name):
	hasheur = hashlib.sha256()
	with open(name, "r") as f:
		buf = f.read()
		test = buf.encode('utf-8')
		hasheur.update(test)
	hashFile = name + ".sha256"
	with open(hashFile, "w+") as f:
		f.write(hasheur.hexdigest())

def main():
	file_name = "Test"
	writeTestFile(file_name)
	hashFile(file_name)


if __name__ == "__main__":
	main()
	
