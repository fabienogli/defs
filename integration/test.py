import hashlib
import requests

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

def upload(name):
	headers = {
		"Content-Type": "application/json"
	}
	multipart_form_data = {
        'file':  ('test', open(name, 'rb'), "multipart/form-data"),
        'filename': ('test', 'Test'),
    }
	test= {'file':open(name, 'rb')}
	#curl -X POST -F ttl="1 minute" -F filename=hello.txt -F file=@hello.txt localhost:8080/file
	data = {'filename':'test'}
	response = requests.post('http://localhost:8080/file', files=test, headers=headers, params=data)
	print(response.content)

def main():
	file_name = "Test"
	writeTestFile(file_name)
	hashFile(file_name)
	upload(file_name)


if __name__ == "__main__":
	main()
	
