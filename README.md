# Test task 1

### how to run the docker container
```bash
docker build -t file-server .
docker run -d -p 9999:9999 --name file-server-container file-server
```

### using
```bash
#  save method
curl -X POST -F "file=@newTest.txt" http://localhost:9999/save/path/newTest.txt

# serve method
curl http://localhost:9999/serve/path/newTest.txt --output privet.txt

# delete method
curl -X DELETE http://localhost:9999/delete/path/newTest.txt
```
  
