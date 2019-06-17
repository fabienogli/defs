import hashlib

BUF_SIZE = 65536

def writeTestFile(name):
	sha1 = hashlib.sha1()
	with open(name, "w+") as f:
		f.write("This is a test")
	with open(name, "r") as f:
		buf = f.read()
		test = buf.encode('utf-8')
		sha1.update(test)
	hashFile = name + ".sha1"
	with open(hashFile, "w+") as f:
		f.write(sha1.hexdigest())
	
def main():
	writeTestFile("Test")

if __name__ == "__main__":
	main()
	
