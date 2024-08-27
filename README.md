# Updated Project
### Quick Start:
1. Run in cmd:
```
git --branch new_ver clone https://github.com/nircoren/lightblocks.git
```
2. Add .env to root of project. Make sure its called .env <br></br>
3. Build images
```
docker-compose build
```
4. Run server

```
docker-compose up -d server
```
5. Run client

```
docker run --rm -it lightblocks-client
```
6. First prompt will be username, second should be a json in this format:
```
		[
			{
				"Action": "addItem",
				"Key": "1",
				"Value": "val1"
			},
			{
				"Action": "getAllItems"
			},
			{
				"Action": "addItem",
				"Key": "2",
				"Value": "val2"
			}
		]
```
  !!! might have issue with parsing the json on bash/wt in windows, better to use another terminal. <br />
  <br />
<b>Example multiple inputs:  </b><br/>
[input1.json](https://github.com/user-attachments/files/16645653/input1.json)  <br/>
[input2.json](https://github.com/user-attachments/files/16645651/input2.json)  <br/>
[input3.json](https://github.com/user-attachments/files/16645652/input3.json)  <br/>

You can try to run multiple instances of the program with different username for each input.
